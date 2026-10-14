package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jeejak/parcel-proof/evidencecheck"
	"github.com/jeejak/parcel-proof/gemini"
)

//go:embed index.html
var indexHTML []byte

// Basic check thresholds, fitted on the odd-numbered records of the
// synthetic eval set and scored on the even ones (Plan 0016 v0.10).
const (
	darkBelow = 78.5
	blurBelow = 31.5
)

// Upload limits. Four photos is what a check sends (D2).
const (
	maxUploadPhotos = 4
	maxPhotoBytes   = 5 << 20
	maxUploadBytes  = maxUploadPhotos*maxPhotoBytes + 64<<10
)

var (
	sampleIDPattern   = regexp.MustCompile(`^R[0-9]{2}$`)
	sampleFilePattern = regexp.MustCompile(`^[0-9A-Za-z_-]{1,40}\.jpe?g$`)
)

// Recorded is a Gemini result captured by the Phase 0 eval, served when
// no key is configured or the live call fails.
type Recorded struct {
	Observation   evidencecheck.Observation `json:"observation"`
	Model         string                    `json:"model"`
	PromptVersion string                    `json:"prompt_version"`
}

// Sample is one staged parcel.
type Sample struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Resi     string   `json:"resi"`
	Koli     int      `json:"koli"`
	Courier  string   `json:"courier"`
	Expected string   `json:"expected"` // the human grade
	Photos   []string `json:"photos"`   // file names under SAMPLES_DIR/<id>/
	Recorded Recorded `json:"recorded"`
}

type serverConfig struct {
	SamplesDir string
	APIKey     string
	Model      string
	Logger     *slog.Logger
	// Live overrides the Gemini checker (tests).
	Live evidencecheck.Checker
	// Store keeps live results across instances (Firestore on Cloud Run).
	// Nil keeps them in memory only.
	Store liveStore
}

type server struct {
	dir     string
	samples map[string]Sample
	order   []string
	live    evidencecheck.Checker
	store   liveStore
	model   string
	logger  *slog.Logger

	mu    sync.Mutex
	cache map[string]checkResult // live Gemini results, one call per sample per process
	locks map[string]*sync.Mutex // one in-flight call per sample
}

func newServer(cfg serverConfig) (*server, error) {
	raw, err := os.ReadFile(filepath.Join(cfg.SamplesDir, "samples.json"))
	if err != nil {
		return nil, fmt.Errorf("read samples: %w", err)
	}
	var list []Sample
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("parse samples: %w", err)
	}
	s := &server{dir: cfg.SamplesDir, samples: map[string]Sample{}, model: cfg.Model, store: cfg.Store, logger: cfg.Logger, cache: map[string]checkResult{}, locks: map[string]*sync.Mutex{}}
	if s.logger == nil {
		s.logger = slog.Default()
	}
	for _, sm := range list {
		if !sampleIDPattern.MatchString(sm.ID) || len(sm.Photos) == 0 {
			return nil, fmt.Errorf("bad sample %q", sm.ID)
		}
		for _, f := range sm.Photos {
			if !sampleFilePattern.MatchString(f) {
				return nil, fmt.Errorf("sample %s: bad photo name %q", sm.ID, f)
			}
		}
		s.samples[sm.ID] = sm
		s.order = append(s.order, sm.ID)
		s.locks[sm.ID] = &sync.Mutex{}
	}
	sort.Strings(s.order)
	switch {
	case cfg.Live != nil:
		s.live = cfg.Live
	case cfg.APIKey != "":
		c, err := gemini.New(context.Background(), gemini.Config{API: gemini.APIGemini, Model: cfg.Model, APIKey: cfg.APIKey})
		if err != nil {
			return nil, err
		}
		s.live = c
	}
	return s, nil
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.index)
	mux.HandleFunc("GET /api/samples", s.listSamples)
	mux.HandleFunc("GET /samples/{id}/{file}", s.samplePhoto)
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/history", s.history)
	mux.HandleFunc("POST /api/check/sample/{id}", s.checkSample)
	mux.HandleFunc("POST /api/check/sample/{id}/live", s.checkSampleLive)
	mux.HandleFunc("POST /api/check/upload", s.checkUpload)
	return secure(mux)
}

func secure(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.ServeHTTP(w, r)
	})
}

func (s *server) index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(indexHTML)
}

type sampleView struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Resi     string   `json:"resi"`
	Koli     int      `json:"koli"`
	Courier  string   `json:"courier"`
	Expected string   `json:"expected"`
	Photos   []string `json:"photos"`
}

func (s *server) listSamples(w http.ResponseWriter, _ *http.Request) {
	out := make([]sampleView, 0, len(s.order))
	for _, id := range s.order {
		sm := s.samples[id]
		urls := make([]string, len(sm.Photos))
		for i, f := range sm.Photos {
			urls[i] = "/samples/" + id + "/" + f
		}
		out = append(out, sampleView{sm.ID, sm.Title, sm.Resi, sm.Koli, sm.Courier, sm.Expected, urls})
	}
	writeJSON(w, http.StatusOK, out)
}

// samplePhoto serves only file names listed in samples.json, so no
// request path ever reaches the filesystem unchecked.
func (s *server) samplePhoto(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.samples[r.PathValue("id")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	file := r.PathValue("file")
	for _, f := range sm.Photos {
		if f == file {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Cache-Control", "public, max-age=86400")
			http.ServeFile(w, r, filepath.Join(s.dir, sm.ID, f))
			return
		}
	}
	http.NotFound(w, r)
}

type checkResult struct {
	Verdict     string                     `json:"verdict"`
	Reasons     []string                   `json:"reasons"`
	Observation *evidencecheck.Observation `json:"observation,omitempty"`
	Source      string                     `json:"source"` // basic | live | recorded
	Model       string                     `json:"model,omitempty"`
	Note        string                     `json:"note,omitempty"`
	CheckedAt   string                     `json:"checked_at,omitempty"`
	LatencyMS   int64                      `json:"latency_ms"`
}

type checkResponse struct {
	Basic         checkResult  `json:"basic"`
	Gemini        *checkResult `json:"gemini,omitempty"`
	LiveAvailable bool         `json:"live_available"`
}

// health is for Cloud Run and uptime checks: the process is up and the
// samples are loaded. It never calls a model.
func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "samples": len(s.samples), "gemini_live": s.live != nil})
}

// checkSample answers at once: the basic check, plus the Gemini result
// this instance already has (live) or the recorded one. It never waits
// on a model call, so a cold instance is as fast as a warm one. The
// live call is checkSampleLive.
func (s *server) checkSample(w http.ResponseWriter, r *http.Request) {
	s.respondSample(w, r, false)
}

// checkSampleLive runs Gemini on the sample now, unless this instance
// already did; see gemini for the limits.
func (s *server) checkSampleLive(w http.ResponseWriter, r *http.Request) {
	s.respondSample(w, r, true)
}

func (s *server) respondSample(w http.ResponseWriter, r *http.Request, live bool) {
	sm, ok := s.samples[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "unknown sample")
		return
	}
	photos, err := s.loadSample(sm)
	if err != nil {
		s.logger.Error("load sample", "id", sm.ID, "err", err)
		writeError(w, http.StatusInternalServerError, "could not read the sample photos")
		return
	}
	basic, err := runBasic(r.Context(), photos, sm.Koli, sm.Courier, sm.Resi)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var g checkResult
	if live {
		g = s.gemini(r.Context(), sm, photos)
	} else {
		g = s.cachedOrRecorded(r.Context(), sm)
	}
	writeJSON(w, http.StatusOK, checkResponse{Basic: basic, Gemini: &g, LiveAvailable: s.live != nil})
}

// gemini returns the live result for a sample, calling the model at most
// once per sample per process; with no key, or when the call fails, it
// returns the result the Phase 0 eval recorded for the same photos.
func (s *server) gemini(ctx context.Context, sm Sample, photos [][]byte) checkResult {
	recorded := func(note string) checkResult { return recordedResult(sm, note) }
	if s.live == nil {
		return recorded("Recorded by the offline eval (prompt " + sm.Recorded.PromptVersion + "); no live key on this deployment.")
	}
	// One call per sample at a time: a second visitor waits for the
	// first call instead of spending another request of the free quota.
	lock := s.locks[sm.ID]
	lock.Lock()
	defer lock.Unlock()
	s.mu.Lock()
	if c, ok := s.cache[sm.ID]; ok {
		s.mu.Unlock()
		return c
	}
	s.mu.Unlock()

	cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	start := time.Now()
	obs, usage, err := s.live.Check(cctx, evidencecheck.Request{Photos: photos, Koli: sm.Koli, Courier: sm.Courier})
	if err != nil {
		s.logger.Warn("live gemini check failed; serving recorded result", "id", sm.ID, "err", err)
		return recorded("Gemini did not answer just now (the free tier may be busy), so this is the result the offline eval recorded for the same photos.")
	}
	d := evidencecheck.Decide(evidencecheck.Input{Obs: obs, Resi: sm.Resi, Koli: sm.Koli})
	model := usage.Model
	if model == "" {
		model = s.model
	}
	res := checkResult{Verdict: d.Verdict, Reasons: nonNil(d.Reasons), Observation: &obs, Source: "live", Model: model,
		LatencyMS: time.Since(start).Milliseconds(), CheckedAt: time.Now().UTC().Format(time.RFC3339)}
	s.mu.Lock()
	s.cache[sm.ID] = res
	s.mu.Unlock()
	if s.store != nil {
		sctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		if err := s.store.Save(sctx, sm.ID, sm.Title, res); err != nil {
			s.logger.Warn("store live result", "id", sm.ID, "err", err)
		}
		cancel()
	}
	return res
}

// stored returns the latest live result another instance saved, and
// keeps it in memory. A store that is slow or down never blocks the
// page: the caller falls back to the recorded result.
func (s *server) stored(ctx context.Context, id string) (checkResult, bool) {
	if s.store == nil {
		return checkResult{}, false
	}
	sctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	c, ok, err := s.store.Latest(sctx, id)
	if err != nil {
		s.logger.Warn("read stored live result", "id", id, "err", err)
		return checkResult{}, false
	}
	if !ok {
		return checkResult{}, false
	}
	s.mu.Lock()
	s.cache[id] = c
	s.mu.Unlock()
	return c, true
}

// history lists the most recent live Gemini checks, newest first.
func (s *server) history(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeJSON(w, http.StatusOK, []historyEntry{})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	h, err := s.store.Recent(ctx, 8)
	if err != nil {
		s.logger.Warn("read history", "err", err)
		writeJSON(w, http.StatusOK, []historyEntry{})
		return
	}
	writeJSON(w, http.StatusOK, h)
}

// checkUpload runs the basic check only. The photos are read into
// memory, checked and dropped; nothing is written and no model is
// called (D12).
func (s *server) checkUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "upload up to 4 photos of at most 5 MB each")
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()
	files := r.MultipartForm.File["photos"]
	if len(files) == 0 || len(files) > maxUploadPhotos {
		writeError(w, http.StatusBadRequest, "choose 1 to 4 photos")
		return
	}
	koli, _ := strconv.Atoi(r.FormValue("koli"))
	if koli < 1 {
		koli = 1
	}
	photos := make([][]byte, 0, len(files))
	for _, fh := range files {
		if fh.Size > maxPhotoBytes {
			writeError(w, http.StatusRequestEntityTooLarge, "each photo must be at most 5 MB")
			return
		}
		f, err := fh.Open()
		if err != nil {
			writeError(w, http.StatusBadRequest, "could not read a photo")
			return
		}
		small, err := evidencecheck.Downscale(io.LimitReader(f, maxPhotoBytes), evidencecheck.MaxEdge)
		_ = f.Close()
		if err != nil {
			writeError(w, http.StatusBadRequest, "each file must be a JPEG or PNG photo")
			return
		}
		photos = append(photos, small)
	}
	basic, err := runBasic(r.Context(), photos, koli, "", strings.TrimSpace(r.FormValue("resi")))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	basic.Note = "Uploaded photos get the basic check only. They are not stored and are never sent to Gemini."
	writeJSON(w, http.StatusOK, checkResponse{Basic: basic})
}

func runBasic(ctx context.Context, photos [][]byte, koli int, courier, resi string) (checkResult, error) {
	start := time.Now()
	c := evidencecheck.BasicChecker{DarkBelow: darkBelow, BlurBelow: blurBelow, ReadBarcodes: true}
	obs, _, err := c.Check(ctx, evidencecheck.Request{Photos: photos, Koli: koli, Courier: courier})
	if err != nil {
		return checkResult{}, fmt.Errorf("basic check: %w", err)
	}
	d := evidencecheck.Decide(evidencecheck.Input{Obs: obs, Resi: resi, Koli: koli})
	return checkResult{Verdict: d.Verdict, Reasons: nonNil(d.Reasons), Observation: &obs, Source: "basic",
		LatencyMS: time.Since(start).Milliseconds()}, nil
}

// recordedResult is the Gemini result the offline eval recorded for the
// same photos, decided by the same rules.
func recordedResult(sm Sample, note string) checkResult {
	obs := sm.Recorded.Observation
	d := evidencecheck.Decide(evidencecheck.Input{Obs: obs, Resi: sm.Resi, Koli: sm.Koli})
	return checkResult{Verdict: d.Verdict, Reasons: nonNil(d.Reasons), Observation: &obs, Source: "recorded",
		Model: sm.Recorded.Model, Note: note}
}

func (s *server) cachedOrRecorded(ctx context.Context, sm Sample) checkResult {
	s.mu.Lock()
	c, ok := s.cache[sm.ID]
	s.mu.Unlock()
	if ok {
		return c
	}
	if c, ok := s.stored(ctx, sm.ID); ok {
		return c
	}
	if s.live == nil {
		return recordedResult(sm, "Recorded by the offline eval (prompt "+sm.Recorded.PromptVersion+"); no live key on this deployment.")
	}
	return recordedResult(sm, "Recorded by the offline eval for these photos (prompt "+sm.Recorded.PromptVersion+"). Press \"Run Gemini live\" to call the model now.")
}

// warm checks every sample once in the background, so the first visitor
// does not wait for a cold model call. A failed call is not cached and
// is tried again on the next visit.
func (s *server) warm(ctx context.Context) {
	if s.live == nil {
		return
	}
	for _, id := range s.order {
		sm := s.samples[id]
		// A live result another instance stored is as good as a new call
		// and costs no quota.
		if _, ok := s.stored(ctx, id); ok {
			continue
		}
		photos, err := s.loadSample(sm)
		if err != nil {
			continue
		}
		s.gemini(ctx, sm, photos)
	}
	s.logger.Info("warmed live gemini results", "samples", len(s.order))
}

func (s *server) loadSample(sm Sample) ([][]byte, error) {
	selected := evidencecheck.SelectPhotos(sm.Photos)
	out := make([][]byte, 0, len(selected))
	for _, f := range selected {
		raw, err := os.ReadFile(filepath.Join(s.dir, sm.ID, f)) //nolint:gosec // G304: names come from samples.json, validated at start
		if err != nil {
			return nil, err
		}
		small, err := evidencecheck.Downscale(bytes.NewReader(raw), evidencecheck.MaxEdge)
		if err != nil {
			return nil, err
		}
		out = append(out, small)
	}
	if len(out) == 0 {
		return nil, errors.New("no photos")
	}
	return out, nil
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
