package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jeejak/parcel-proof/evidencecheck"
)

func jpegBytes(t *testing.T, v uint8) []byte {
	t.Helper()
	img := image.NewGray(image.Rect(0, 0, 120, 90))
	for i := range img.Pix {
		img.Pix[i] = v
		if i%7 == 0 {
			img.Pix[i] = 255 - v // edges, so the photo is not blurry
		}
	}
	var b bytes.Buffer
	_ = jpeg.Encode(&b, img, nil)
	return b.Bytes()
}

func stage(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "R01"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"01.jpg", "02.jpg"} {
		if err := os.WriteFile(filepath.Join(dir, "R01", f), jpegBytes(t, 200), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	samples := `[{"id":"R01","title":"Good","resi":"SPX1","koli":1,"courier":"SPX","expected":"kuat","photos":["01.jpg","02.jpg"],
		"recorded":{"observation":{"image_usable":"ok","label_visible":"yes","label_resi_read":"SPX1","package_sealed":"yes","item_visible":"yes","koli_seen":1},"model":"m","prompt_version":"p"}}]`
	if err := os.WriteFile(filepath.Join(dir, "samples.json"), []byte(samples), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func post(t *testing.T, h http.Handler, path string, body *bytes.Buffer, ctype string) (int, checkResponse) {
	t.Helper()
	if body == nil {
		body = &bytes.Buffer{}
	}
	req := httptest.NewRequest(http.MethodPost, path, body)
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out checkResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestSampleCheck_RecordedWithoutKey(t *testing.T) {
	s, err := newServer(serverConfig{SamplesDir: stage(t)})
	if err != nil {
		t.Fatal(err)
	}
	code, out := post(t, s.routes(), "/api/check/sample/R01", nil, "")
	if code != 200 || out.Gemini == nil || out.Gemini.Source != "recorded" || out.Gemini.Verdict != "kuat" || out.Basic.Source != "basic" {
		t.Fatalf("code %d out %+v", code, out)
	}
}

func TestSampleCheck_LiveOncePerSample(t *testing.T) {
	var calls atomic.Int32
	live := evidencecheck.FakeChecker{Fn: func(context.Context, evidencecheck.Request) (evidencecheck.Observation, evidencecheck.Usage, error) {
		calls.Add(1)
		return evidencecheck.Observation{ImageUsable: "dark", LabelVisible: "unsure", PackageSealed: "yes", ItemVisible: "yes"}, evidencecheck.Usage{Model: "live-model"}, nil
	}}
	s, err := newServer(serverConfig{SamplesDir: stage(t), Live: live})
	if err != nil {
		t.Fatal(err)
	}
	h := s.routes()
	for i := 0; i < 3; i++ {
		_, out := post(t, h, "/api/check/sample/R01/live", nil, "")
		if out.Gemini.Source != "live" || out.Gemini.Verdict != "lemah" {
			t.Fatalf("run %d: %+v", i, out.Gemini)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("model called %d times, want once per sample", calls.Load())
	}
	// Once this instance has a live result, the default answer uses it.
	if _, out := post(t, h, "/api/check/sample/R01", nil, ""); out.Gemini.Source != "live" || out.Gemini.CheckedAt == "" {
		t.Fatalf("default after live: %+v", out.Gemini)
	}
}

// The default answer never waits on the model, so a cold instance is
// as fast as a warm one; it serves the recorded result and says how to
// run it live.
func TestSampleCheck_DefaultNeverCallsModel(t *testing.T) {
	live := evidencecheck.FakeChecker{Fn: func(context.Context, evidencecheck.Request) (evidencecheck.Observation, evidencecheck.Usage, error) {
		t.Error("the default sample check called the model")
		return evidencecheck.Observation{}, evidencecheck.Usage{}, nil
	}}
	s, _ := newServer(serverConfig{SamplesDir: stage(t), Live: live})
	_, out := post(t, s.routes(), "/api/check/sample/R01", nil, "")
	if out.Gemini.Source != "recorded" || !out.LiveAvailable || !strings.Contains(out.Gemini.Note, "Run Gemini live") {
		t.Fatalf("%+v", out)
	}
}

func TestHealth(t *testing.T) {
	s, _ := newServer(serverConfig{SamplesDir: stage(t)})
	rec := httptest.NewRecorder()
	s.routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

func TestSampleCheck_LiveFailureFallsBackToRecorded(t *testing.T) {
	live := evidencecheck.FakeChecker{Fn: func(context.Context, evidencecheck.Request) (evidencecheck.Observation, evidencecheck.Usage, error) {
		return evidencecheck.Observation{}, evidencecheck.Usage{}, errors.New("http 429")
	}}
	s, _ := newServer(serverConfig{SamplesDir: stage(t), Live: live})
	_, out := post(t, s.routes(), "/api/check/sample/R01/live", nil, "")
	if out.Gemini.Source != "recorded" || !strings.Contains(out.Gemini.Note, "did not answer") {
		t.Fatalf("%+v", out.Gemini)
	}
}

// D12: an uploaded photo never reaches the model.
func TestUpload_BasicOnlyNeverCallsModel(t *testing.T) {
	live := evidencecheck.FakeChecker{Fn: func(context.Context, evidencecheck.Request) (evidencecheck.Observation, evidencecheck.Usage, error) {
		t.Error("an uploaded photo was sent to the model")
		return evidencecheck.Observation{}, evidencecheck.Usage{}, nil
	}}
	s, _ := newServer(serverConfig{SamplesDir: stage(t), Live: live})
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("photos", "dark.jpg")
	_, _ = fw.Write(jpegBytes(t, 20))
	_ = mw.WriteField("resi", "SPX1")
	_ = mw.Close()
	code, out := post(t, s.routes(), "/api/check/upload", &body, mw.FormDataContentType())
	if code != 200 || out.Gemini != nil || out.Basic.Verdict != "lemah" || out.Basic.Reasons[0] != "image_dark" {
		t.Fatalf("code %d out %+v", code, out)
	}
}

func TestUpload_Rejects(t *testing.T) {
	s, _ := newServer(serverConfig{SamplesDir: stage(t)})
	h := s.routes()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for i := 0; i < 5; i++ {
		fw, _ := mw.CreateFormFile("photos", "p.jpg")
		_, _ = fw.Write(jpegBytes(t, 200))
	}
	_ = mw.Close()
	if code, _ := post(t, h, "/api/check/upload", &body, mw.FormDataContentType()); code != http.StatusBadRequest {
		t.Fatalf("five photos: %d", code)
	}
	body.Reset()
	mw = multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("photos", "x.jpg")
	_, _ = fw.Write([]byte("not an image"))
	_ = mw.Close()
	if code, _ := post(t, h, "/api/check/upload", &body, mw.FormDataContentType()); code != http.StatusBadRequest {
		t.Fatalf("garbage: %d", code)
	}
}

func TestSamplePhoto_ServesListedFilesOnly(t *testing.T) {
	s, _ := newServer(serverConfig{SamplesDir: stage(t)})
	h := s.routes()
	get := func(p string) int {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		return rec.Code
	}
	if c := get("/samples/R01/01.jpg"); c != 200 {
		t.Fatalf("listed photo: %d", c)
	}
	for _, p := range []string{"/samples/R01/samples.json", "/samples/R02/01.jpg", "/samples/R01/..%2fsamples.json"} {
		if c := get(p); c == 200 {
			t.Errorf("%s served", p)
		}
	}
	if code, _ := post(t, h, "/api/check/sample/R99", nil, ""); code != http.StatusNotFound {
		t.Errorf("unknown sample: %d", code)
	}
}

func TestNewServer_RejectsBadPhotoNames(t *testing.T) {
	dir := stage(t)
	_ = os.WriteFile(filepath.Join(dir, "samples.json"), []byte(`[{"id":"R01","photos":["../../etc/passwd"]}]`), 0o644)
	if _, err := newServer(serverConfig{SamplesDir: dir}); err == nil {
		t.Fatal("accepted a traversal photo name")
	}
}

// fakeStore is an in-memory liveStore.
type fakeStore struct {
	mu      sync.Mutex
	latest  map[string]checkResult
	history []historyEntry
	failGet bool
}

func newFakeStore() *fakeStore { return &fakeStore{latest: map[string]checkResult{}} }

func (f *fakeStore) Latest(_ context.Context, id string) (checkResult, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failGet {
		return checkResult{}, false, errors.New("firestore down")
	}
	c, ok := f.latest[id]
	return c, ok, nil
}

func (f *fakeStore) Save(_ context.Context, id, title string, res checkResult) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.latest[id] = res
	f.history = append([]historyEntry{{SampleID: id, Title: title, Result: res}}, f.history...)
	return nil
}

func (f *fakeStore) Recent(_ context.Context, n int) ([]historyEntry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.history) > n {
		return f.history[:n], nil
	}
	return f.history, nil
}

func darkLive(calls *atomic.Int32) evidencecheck.Checker {
	return evidencecheck.FakeChecker{Fn: func(context.Context, evidencecheck.Request) (evidencecheck.Observation, evidencecheck.Usage, error) {
		calls.Add(1)
		return evidencecheck.Observation{ImageUsable: "dark", LabelVisible: "unsure", PackageSealed: "yes", ItemVisible: "yes"}, evidencecheck.Usage{Model: "live-model"}, nil
	}}
}

// A live result is saved to the store, and a fresh instance (new server,
// same store) serves it on the default path without calling the model.
func TestStore_LiveResultSurvivesANewInstance(t *testing.T) {
	dir := stage(t)
	st := newFakeStore()
	var calls atomic.Int32
	a, _ := newServer(serverConfig{SamplesDir: dir, Live: darkLive(&calls), Store: st})
	if _, out := post(t, a.routes(), "/api/check/sample/R01/live", nil, ""); out.Gemini.Source != "live" {
		t.Fatalf("live: %+v", out.Gemini)
	}
	b, _ := newServer(serverConfig{SamplesDir: dir, Live: darkLive(&calls), Store: st})
	_, out := post(t, b.routes(), "/api/check/sample/R01", nil, "")
	if out.Gemini.Source != "live" || out.Gemini.Verdict != "lemah" || calls.Load() != 1 {
		t.Fatalf("fresh instance: %+v, model calls %d", out.Gemini, calls.Load())
	}
	// Warming a fresh instance reads the store instead of the model.
	c, _ := newServer(serverConfig{SamplesDir: dir, Live: darkLive(&calls), Store: st})
	c.warm(context.Background())
	if calls.Load() != 1 {
		t.Fatalf("warm called the model with a stored result: %d calls", calls.Load())
	}
}

// A store that fails never breaks the page: the recorded result shows.
func TestStore_FailureFallsBackToRecorded(t *testing.T) {
	st := newFakeStore()
	st.failGet = true
	s, _ := newServer(serverConfig{SamplesDir: stage(t), Store: st})
	code, out := post(t, s.routes(), "/api/check/sample/R01", nil, "")
	if code != 200 || out.Gemini.Source != "recorded" {
		t.Fatalf("code %d %+v", code, out.Gemini)
	}
}

func TestHistory(t *testing.T) {
	st := newFakeStore()
	var calls atomic.Int32
	s, _ := newServer(serverConfig{SamplesDir: stage(t), Live: darkLive(&calls), Store: st})
	h := s.routes()
	post(t, h, "/api/check/sample/R01/live", nil, "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/history", nil))
	var got []historyEntry
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != 200 || len(got) != 1 || got[0].SampleID != "R01" || got[0].Result.Source != "live" {
		t.Fatalf("history %d %s", rec.Code, rec.Body.String())
	}
	// Without a store the history is an empty list, not an error.
	s2, _ := newServer(serverConfig{SamplesDir: stage(t)})
	rec = httptest.NewRecorder()
	s2.routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/history", nil))
	if rec.Code != 200 || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("no store: %d %s", rec.Code, rec.Body.String())
	}
}
