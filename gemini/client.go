// Package vertex is the Gemini implementation of evidencecheck.Checker
// (Plan 0016 D10). It speaks generateContent to one of two APIs:
//
//   - Vertex AI, the default, and the only one Phase 1 may use.
//   - The Gemini Developer API, for the Phase 0 eval on staged photos
//     only (Plan 0016 D11). Its free tier costs nothing, but Google may
//     use what is sent to it and human reviewers may read it, so no
//     production photo may ever go there.
//
// It calls the generateContent REST endpoint directly through
// google.golang.org/api/transport/http, which is already a dependency
// of this module, rather than pulling in a Gemini SDK. Credentials are
// Application Default Credentials with the cloud-platform scope: a
// service-account key file named by GOOGLE_APPLICATION_CREDENTIALS, or
// the metadata server when running on Google Cloud. No credential ever
// appears in a URL, so transport errors are safe to wrap as-is, a
// custom BaseURL is only accepted alongside an explicit HTTPClient, and
// the client never follows a redirect, so that a token minted for
// Vertex can never be sent elsewhere. The
// Gemini Developer API authenticates with an API key in the
// x-goog-api-key header instead, and the same rule holds: a custom
// BaseURL is never combined with a key.
package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"

	"github.com/jeejak/parcel-proof/evidencecheck"
)

const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

// The two APIs Config.API selects between.
const (
	APIVertex = "vertex"     // the default
	APIGemini = "gemini-api" // Phase 0 staged photos only (D11)
)

// Media resolution levels Config.MediaResolution accepts. Each sets how
// many tokens a photo is encoded into; empty leaves the model default.
const (
	MediaResolutionLow    = "low"
	MediaResolutionMedium = "medium"
	MediaResolutionHigh   = "high"
)

var mediaResolutionLevels = map[string]string{
	MediaResolutionLow:    "MEDIA_RESOLUTION_LOW",
	MediaResolutionMedium: "MEDIA_RESOLUTION_MEDIUM",
	MediaResolutionHigh:   "MEDIA_RESOLUTION_HIGH",
}

// CheckMediaResolution reports whether level is one New accepts: a
// MediaResolution* constant or empty.
func CheckMediaResolution(level string) error {
	if _, ok := mediaResolutionLevels[level]; !ok && level != "" {
		return fmt.Errorf("vertex: unknown media resolution %q: want low, medium, high or empty", level)
	}
	return nil
}

// defaultGeminiBaseURL is the Gemini Developer API host.
const defaultGeminiBaseURL = "https://generativelanguage.googleapis.com"

// ErrQuota and ErrAuth type the two refusals a caller should stop on
// rather than retry: HTTP 429, and HTTP 401 or 403.
var (
	ErrQuota = errors.New("vertex: quota or rate limit exceeded")
	ErrAuth  = errors.New("vertex: not authorised")
)

// maxResponseBytes bounds how much of a response is read. A valid
// answer is a few hundred bytes of JSON plus metadata.
const maxResponseBytes = 1 << 20

// Config selects the model. Project, Location and Model are required.
type Config struct {
	API      string // APIVertex (the default when empty) or APIGemini
	Project  string // GCP project id (Vertex only)
	Location string // e.g. "eu", "us", or "global" (Vertex only)
	Model    string // exact model id, recorded in Plan 0016 §7 Q4
	APIKey   string // Gemini Developer API key (APIGemini only)

	// MediaResolution is set on every photo part: one of the
	// MediaResolution* constants, or empty for the model's default (no
	// field is sent). It changes what the model sees, so a result
	// checked at one level is not the same check as at another (D6).
	MediaResolution string

	// HTTPClient overrides the authenticated client. Production leaves
	// both of these empty.
	//
	// BaseURL MUST be paired with an HTTPClient: New refuses the
	// combination of a custom host and the default credential chain,
	// because that chain mints a real OAuth bearer token for the
	// cloud-platform scope and would send it to whatever host was
	// named. A mock or a captive proxy would receive a usable Google
	// credential. See New.
	HTTPClient *http.Client
	BaseURL    string
}

// Checker implements evidencecheck.Checker against Vertex AI.
type Checker struct {
	cfg    Config
	client *http.Client
	url    string
	apiKey string // sent as x-goog-api-key; Gemini Developer API only
}

var _ evidencecheck.Checker = (*Checker)(nil)

// New builds a Checker. With no HTTPClient it resolves Application
// Default Credentials now, so a misconfigured deployment fails at boot
// rather than on the first record.
func New(ctx context.Context, cfg Config) (*Checker, error) {
	if err := CheckMediaResolution(cfg.MediaResolution); err != nil {
		return nil, err
	}
	switch cfg.API {
	case "", APIVertex:
		return newVertex(ctx, cfg)
	case APIGemini:
		return newGemini(cfg)
	}
	return nil, fmt.Errorf("vertex: unknown API %q", cfg.API)
}

// newGemini builds a Checker for the Gemini Developer API (D11). It
// needs no Google credentials and no project; the API key is the only
// credential, and it is never combined with a custom host.
func newGemini(cfg Config) (*Checker, error) {
	if cfg.Model == "" {
		return nil, errors.New("vertex: model is required")
	}
	if cfg.BaseURL != "" && cfg.APIKey != "" {
		return nil, errors.New("vertex: BaseURL cannot be combined with an APIKey; refusing to send a Gemini API key to a non-Google host")
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{}
	}
	client = refuseRedirects(client)
	base := cfg.BaseURL
	if base == "" {
		base = defaultGeminiBaseURL
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent", strings.TrimRight(base, "/"), cfg.Model)
	return &Checker{cfg: cfg, client: client, url: url, apiKey: cfg.APIKey}, nil
}

func newVertex(ctx context.Context, cfg Config) (*Checker, error) {
	if cfg.Project == "" || cfg.Location == "" || cfg.Model == "" {
		return nil, errors.New("vertex: project, location and model are required")
	}
	// Fail closed on "custom host plus real credentials". The eval's
	// -base-url path already passes a bare client, but that is a
	// caller's discipline; this makes the credential leak impossible
	// from inside the package instead of relying on every caller to
	// remember.
	if cfg.BaseURL != "" && cfg.HTTPClient == nil {
		return nil, errors.New("vertex: BaseURL requires an explicit HTTPClient; refusing to send Application Default Credentials to a non-Vertex host")
	}
	client := cfg.HTTPClient
	if client == nil {
		c, _, err := htransport.NewClient(ctx, option.WithScopes(cloudPlatformScope))
		if err != nil {
			return nil, fmt.Errorf("vertex: credentials: %w", err)
		}
		client = c
	}
	client = refuseRedirects(client)
	base := cfg.BaseURL
	if base == "" {
		base = defaultBaseURL(cfg.Location)
	}
	url := fmt.Sprintf("%s/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		strings.TrimRight(base, "/"), cfg.Project, cfg.Location, cfg.Model)
	return &Checker{cfg: cfg, client: client, url: url}, nil
}

// errRedirect is returned when the API answers with a redirect.
var errRedirect = errors.New("vertex: refusing to follow a redirect; credentials are sent only to the configured host")

// refuseRedirects returns a copy of c that never follows a redirect.
// Following one would send the credential to whatever host the 3xx
// names: net/http forwards custom headers such as x-goog-api-key on a
// redirect, and the oauth2 transport behind htransport.NewClient adds
// the bearer token to every hop. A generateContent call is not
// expected to redirect; if one ever does, it fails visibly instead of
// leaking. The caller's client is copied, not modified.
func refuseRedirects(c *http.Client) *http.Client {
	cp := *c
	cp.CheckRedirect = func(*http.Request, []*http.Request) error { return errRedirect }
	return &cp
}

// defaultBaseURL maps a location to its REST host. The us and eu
// multi-regions are the only pay-as-you-go endpoints that keep ML
// processing inside a jurisdiction, and they live on the .rep. hosts,
// not on a "<location>-aiplatform" host.
func defaultBaseURL(location string) string {
	switch location {
	case "global":
		return "https://aiplatform.googleapis.com"
	case "us", "eu":
		return "https://aiplatform." + location + ".rep.googleapis.com"
	}
	return "https://" + location + "-aiplatform.googleapis.com"
}

// Check sends one record's photos and returns the parsed observation.
// A response that does not match the schema returns an error wrapping
// evidencecheck.ErrSchema, with usage still filled in so the eval can
// count what the failed call cost.
func (c *Checker) Check(ctx context.Context, req evidencecheck.Request) (evidencecheck.Observation, evidencecheck.Usage, error) {
	usage := evidencecheck.Usage{Model: c.cfg.Model}
	if len(req.Photos) == 0 {
		return evidencecheck.Observation{}, usage, errors.New("vertex: no photos")
	}

	body, err := json.Marshal(buildRequest(req, c.cfg.MediaResolution))
	if err != nil {
		return evidencecheck.Observation{}, usage, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return evidencecheck.Observation{}, usage, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("x-goog-api-key", c.apiKey)
	}

	started := time.Now()
	resp, err := c.client.Do(httpReq)
	usage.Latency = time.Since(started)
	if err != nil {
		return evidencecheck.Observation{}, usage, fmt.Errorf("vertex: request: %w", err)
	}
	defer resp.Body.Close()
	// One byte past the limit, so an oversized body is detected rather
	// than quietly truncated into malformed JSON.
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return evidencecheck.Observation{}, usage, fmt.Errorf("vertex: read: %w", err)
	}
	oversized := len(raw) > maxResponseBytes
	if resp.StatusCode/100 != 2 {
		msg := apiErrorMessage(raw)
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return evidencecheck.Observation{}, usage, fmt.Errorf("%w: http %d: %s", ErrQuota, resp.StatusCode, msg)
		case http.StatusUnauthorized, http.StatusForbidden:
			return evidencecheck.Observation{}, usage, fmt.Errorf("%w: http %d: %s", ErrAuth, resp.StatusCode, msg)
		}
		return evidencecheck.Observation{}, usage, fmt.Errorf("vertex: http %d: %s", resp.StatusCode, msg)
	}

	// A successful response this large is the model ignoring the schema
	// and writing prose, which is a schema failure by D3, not a
	// transport problem. Truncating first and letting json.Unmarshal
	// fail would report it as "decode: unexpected end of JSON input"
	// and count it in the wrong bucket, so the eval's schema-failure
	// rate would understate reality.
	if oversized {
		return evidencecheck.Observation{}, usage, fmt.Errorf("%w: response exceeds %d bytes", evidencecheck.ErrSchema, maxResponseBytes)
	}

	var gr generateResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return evidencecheck.Observation{}, usage, fmt.Errorf("vertex: decode: %w", err)
	}
	usage.InputTokens = gr.UsageMetadata.PromptTokenCount
	// thoughtsTokenCount is reported apart from candidatesTokenCount but
	// billed at the output rate ("response and reasoning" on the Vertex
	// price list), so leaving it out under-reports cost.
	usage.ThoughtTokens = gr.UsageMetadata.ThoughtsTokenCount
	usage.OutputTokens = gr.UsageMetadata.CandidatesTokenCount + gr.UsageMetadata.ThoughtsTokenCount
	if gr.ModelVersion != "" {
		usage.Model = gr.ModelVersion
	}

	if len(gr.Candidates) == 0 {
		reason := gr.PromptFeedback.BlockReason
		if reason == "" {
			reason = "no candidates"
		}
		return evidencecheck.Observation{}, usage, fmt.Errorf("%w: %s", evidencecheck.ErrSchema, reason)
	}
	cand := gr.Candidates[0]
	if cand.FinishReason != "" && cand.FinishReason != "STOP" {
		return evidencecheck.Observation{}, usage, fmt.Errorf("%w: finish reason %s", evidencecheck.ErrSchema, cand.FinishReason)
	}
	var text strings.Builder
	for _, p := range cand.Content.Parts {
		text.WriteString(p.Text)
	}
	obs, err := evidencecheck.ParseObservation([]byte(text.String()))
	return obs, usage, err
}

func buildRequest(req evidencecheck.Request, resolution string) generateRequest {
	var mr *mediaResolution
	if level, ok := mediaResolutionLevels[resolution]; ok {
		mr = &mediaResolution{Level: level}
	}
	parts := []part{{Text: evidencecheck.UserText(req)}}
	for _, p := range req.Photos {
		parts = append(parts, part{InlineData: &blob{
			MimeType: "image/jpeg",
			Data:     base64.StdEncoding.EncodeToString(p),
		}, MediaResolution: mr})
	}
	return generateRequest{
		SystemInstruction: &content{Parts: []part{{Text: evidencecheck.SystemInstruction}}},
		Contents:          []content{{Role: "user", Parts: parts}},
		GenerationConfig: generationConfig{
			Temperature:      0,
			ResponseMimeType: "application/json",
			ResponseSchema:   evidencecheck.ResponseSchema(),
		},
	}
}

// apiErrorMessage pulls error.message out of a Google API error body,
// falling back to a truncated raw body.
func apiErrorMessage(raw []byte) string {
	var e struct {
		Error struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &e) == nil && e.Error.Message != "" {
		if e.Error.Status != "" {
			return e.Error.Status + ": " + e.Error.Message
		}
		return e.Error.Message
	}
	s := string(raw)
	if len(s) > 300 {
		s = s[:300]
	}
	return s
}

// Wire types for generateContent. Only the fields this package uses.

type generateRequest struct {
	SystemInstruction *content         `json:"systemInstruction,omitempty"`
	Contents          []content        `json:"contents"`
	GenerationConfig  generationConfig `json:"generationConfig"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text            string           `json:"text,omitempty"`
	InlineData      *blob            `json:"inlineData,omitempty"`
	MediaResolution *mediaResolution `json:"mediaResolution,omitempty"`
}

type mediaResolution struct {
	Level string `json:"level"`
}

type blob struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type generationConfig struct {
	Temperature      float64        `json:"temperature"`
	ResponseMimeType string         `json:"responseMimeType"`
	ResponseSchema   map[string]any `json:"responseSchema"`
}

type generateResponse struct {
	Candidates []struct {
		Content      content `json:"content"`
		FinishReason string  `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		ThoughtsTokenCount   int `json:"thoughtsTokenCount"`
	} `json:"usageMetadata"`
	ModelVersion string `json:"modelVersion"`
}
