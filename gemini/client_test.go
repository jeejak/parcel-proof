package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeejak/parcel-proof/evidencecheck"
)

const okObservation = `{"image_usable":"ok","label_visible":"yes","label_resi_read":"SPX1","package_sealed":"yes","item_visible":"yes","koli_seen":1}`

func newTest(t *testing.T, h http.HandlerFunc) *Checker {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := New(context.Background(), Config{
		Project: "proj", Location: "asia-southeast2", Model: "gemini-test",
		HTTPClient: srv.Client(), BaseURL: srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func reply(w http.ResponseWriter, text, finish string) {
	resp := map[string]any{
		"candidates": []any{map[string]any{
			"content":      map[string]any{"role": "model", "parts": []any{map[string]any{"text": text}}},
			"finishReason": finish,
		}},
		"usageMetadata": map[string]any{"promptTokenCount": 1300, "candidatesTokenCount": 40, "thoughtsTokenCount": 250},
		"modelVersion":  "gemini-test-001",
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func TestCheck_SendsWhatD2AllowsAndParses(t *testing.T) {
	var got generateRequest
	var path string
	c := newTest(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Errorf("bad request body: %v", err)
		}
		reply(w, okObservation, "STOP")
	})

	photo := []byte{0xff, 0xd8, 0xff, 0x01}
	obs, usage, err := c.Check(context.Background(), evidencecheck.Request{
		Photos: [][]byte{photo, photo}, Koli: 1, Courier: "SPX",
	})
	if err != nil {
		t.Fatal(err)
	}

	if want := "/v1/projects/proj/locations/asia-southeast2/publishers/google/models/gemini-test:generateContent"; path != want {
		t.Errorf("path = %s", path)
	}
	if obs.LabelResiRead == nil || *obs.LabelResiRead != "SPX1" || obs.ImageUsable != "ok" {
		t.Errorf("observation = %+v", obs)
	}
	// Reasoning tokens are reported apart from candidatesTokenCount but
	// billed at the output rate, so they count as output.
	if usage.InputTokens != 1300 || usage.OutputTokens != 290 || usage.ThoughtTokens != 250 || usage.Model != "gemini-test-001" {
		t.Errorf("usage = %+v", usage)
	}

	parts := got.Contents[0].Parts
	if len(parts) != 3 || parts[0].Text != evidencecheck.UserText(evidencecheck.Request{Photos: make([][]byte, 2), Koli: 1, Courier: "SPX"}) {
		t.Fatalf("parts = %+v", parts)
	}
	if parts[1].InlineData.MimeType != "image/jpeg" || parts[1].InlineData.Data != base64.StdEncoding.EncodeToString(photo) {
		t.Errorf("photo part = %+v", parts[1].InlineData)
	}
	if got.GenerationConfig.ResponseMimeType != "application/json" || got.GenerationConfig.ResponseSchema == nil {
		t.Errorf("generation config = %+v", got.GenerationConfig)
	}
	if got.SystemInstruction == nil || got.SystemInstruction.Parts[0].Text != evidencecheck.SystemInstruction {
		t.Error("system instruction missing")
	}
}

func TestCheck_SchemaFailuresWrapErrSchema(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"prose instead of json": func(w http.ResponseWriter, r *http.Request) { reply(w, "Looks fine to me.", "STOP") },
		"truncated":             func(w http.ResponseWriter, r *http.Request) { reply(w, okObservation[:30], "MAX_TOKENS") },
		// A parseable body is still discarded when the model did not
		// finish normally: the parser alone would accept these.
		"parseable but cut off": func(w http.ResponseWriter, r *http.Request) { reply(w, okObservation, "MAX_TOKENS") },
		"parseable but stopped for safety": func(w http.ResponseWriter, r *http.Request) {
			reply(w, okObservation, "SAFETY")
		},
		"blocked": func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"promptFeedback":{"blockReason":"SAFETY"},"usageMetadata":{"promptTokenCount":900}}`))
		},
	}
	for name, h := range cases {
		t.Run(name, func(t *testing.T) {
			_, usage, err := newTest(t, h).Check(context.Background(), evidencecheck.Request{Photos: [][]byte{{1}}, Koli: 1})
			if !errors.Is(err, evidencecheck.ErrSchema) {
				t.Fatalf("err = %v, want ErrSchema", err)
			}
			if usage.InputTokens == 0 {
				t.Error("usage not recorded on a schema failure")
			}
		})
	}
}

func TestCheck_HTTPErrorSurfacesAPIMessage(t *testing.T) {
	c := newTest(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"code":403,"message":"Permission denied on resource project proj.","status":"PERMISSION_DENIED"}}`))
	})
	_, _, err := c.Check(context.Background(), evidencecheck.Request{Photos: [][]byte{{1}}, Koli: 1})
	if err == nil || errors.Is(err, evidencecheck.ErrSchema) || !strings.Contains(err.Error(), "PERMISSION_DENIED") {
		t.Fatalf("err = %v", err)
	}
}

func TestNew_ValidatesAndPicksHost(t *testing.T) {
	if _, err := New(context.Background(), Config{Project: "p", Location: "x"}); err == nil {
		t.Fatal("missing model accepted")
	}
	if got := defaultBaseURL("asia-southeast2"); got != "https://asia-southeast2-aiplatform.googleapis.com" {
		t.Errorf("regional host = %s", got)
	}
	if got := defaultBaseURL("global"); got != "https://aiplatform.googleapis.com" {
		t.Errorf("global host = %s", got)
	}
	// Multi-region endpoints use the .rep. hostnames; a "us-" prefix host
	// does not exist for them.
	if got := defaultBaseURL("us"); got != "https://aiplatform.us.rep.googleapis.com" {
		t.Errorf("us multi-region host = %s", got)
	}
	if got := defaultBaseURL("eu"); got != "https://aiplatform.eu.rep.googleapis.com" {
		t.Errorf("eu multi-region host = %s", got)
	}
}

// A custom host plus the default credential chain is the one
// combination that leaks: htransport mints a cloud-platform bearer
// token and Do would send it to whatever host BaseURL named. The eval
// avoids it by passing a bare client, but New must not depend on every
// caller remembering that.
func TestNew_RefusesBaseURLWithoutExplicitClient(t *testing.T) {
	_, err := New(context.Background(), Config{
		Project: "p", Location: "global", Model: "m",
		BaseURL: "http://127.0.0.1:1/mock",
	})
	if err == nil {
		t.Fatal("BaseURL with no HTTPClient was accepted; ADC would be sent to the mock host")
	}
	if !strings.Contains(err.Error(), "HTTPClient") {
		t.Errorf("error should name the missing HTTPClient, got %v", err)
	}

	// The supported shape still works and does not touch credentials.
	if _, err := New(context.Background(), Config{
		Project: "p", Location: "global", Model: "m",
		BaseURL: "http://127.0.0.1:1/mock", HTTPClient: &http.Client{},
	}); err != nil {
		t.Fatalf("BaseURL with an explicit client rejected: %v", err)
	}
}

func TestCheck_RejectsEmptyPhotoSet(t *testing.T) {
	c := newTest(t, func(w http.ResponseWriter, r *http.Request) { t.Error("no request expected") })
	if _, _, err := c.Check(context.Background(), evidencecheck.Request{}); err == nil {
		t.Fatal("empty photo set accepted")
	}
}

// Plan 0016 D11: the Gemini Developer API backend puts the key in the
// x-goog-api-key header (never the URL, which ends up in logs and error
// strings) and needs no project, location or Google credentials.
func TestGeminiAPI_KeyInHeaderNotURL(t *testing.T) {
	var path, query, keyHeader, auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		keyHeader, auth = r.Header.Get("x-goog-api-key"), r.Header.Get("Authorization")
		reply(w, okObservation, "STOP")
	}))
	defer srv.Close()
	// New refuses a custom host together with a key (see
	// TestGeminiAPI_RefusesKeyWithCustomHost), so the key is set inside
	// the package here to observe how it is sent.
	c, err := New(context.Background(), Config{
		API: APIGemini, Model: "gemini-test", HTTPClient: srv.Client(), BaseURL: srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	c.apiKey = "secret-key"
	if _, _, err := c.Check(context.Background(), evidencecheck.Request{Photos: [][]byte{{1}}, Koli: 1}); err != nil {
		t.Fatal(err)
	}
	if path != "/v1beta/models/gemini-test:generateContent" {
		t.Errorf("path = %s", path)
	}
	if query != "" || strings.Contains(c.url, "secret-key") {
		t.Errorf("key leaked into the URL: query %q url %q", query, c.url)
	}
	if keyHeader != "secret-key" || auth != "" {
		t.Errorf("x-goog-api-key = %q, Authorization = %q", keyHeader, auth)
	}
	if got := defaultGeminiBaseURL; got != "https://generativelanguage.googleapis.com" {
		t.Errorf("gemini host = %s", got)
	}
}

func TestGeminiAPI_RequiresModelOnly(t *testing.T) {
	if _, err := New(context.Background(), Config{API: APIGemini, APIKey: "k"}); err == nil {
		t.Error("missing model accepted")
	}
	if _, err := New(context.Background(), Config{API: "something-else", Model: "m"}); err == nil {
		t.Error("unknown API accepted")
	}
}

// The eval stops on quota and permission errors instead of continuing
// to send requests an API is refusing, so those two must be typed.
func TestCheck_QuotaAndAuthErrorsAreTyped(t *testing.T) {
	for status, want := range map[int]error{429: ErrQuota, 401: ErrAuth, 403: ErrAuth, 500: nil} {
		c := newTest(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":{"message":"refused","status":"X"}}`))
		})
		_, _, err := c.Check(context.Background(), evidencecheck.Request{Photos: [][]byte{{1}}, Koli: 1})
		if err == nil {
			t.Fatalf("%d: no error", status)
		}
		if want != nil && !errors.Is(err, want) {
			t.Errorf("%d: %v is not %v", status, err, want)
		}
		if want == nil && (errors.Is(err, ErrQuota) || errors.Is(err, ErrAuth)) {
			t.Errorf("%d: typed as quota or auth: %v", status, err)
		}
	}
}

// Same fail-closed rule as Vertex credentials: a custom host must never
// receive the API key, so New refuses the combination outright.
func TestGeminiAPI_RefusesKeyWithCustomHost(t *testing.T) {
	_, err := New(context.Background(), Config{API: APIGemini, Model: "m", APIKey: "k", BaseURL: "http://127.0.0.1:1"})
	if err == nil {
		t.Fatal("custom host with an API key accepted")
	}
	c, err := New(context.Background(), Config{API: APIGemini, Model: "m", APIKey: "k"})
	if err != nil || c.url != defaultGeminiBaseURL+"/v1beta/models/m:generateContent" || c.apiKey != "k" {
		t.Fatalf("default host: %v %+v", err, c)
	}
}

// A 2xx body far over the read limit is the model ignoring the schema
// and writing prose. It must land in the schema-failure bucket, which
// is half of a D8 target: truncating and letting json.Unmarshal fail
// would report it as a decode error and understate the measured rate.
func TestCheck_OversizedResponseIsASchemaFailure(t *testing.T) {
	c := newTest(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Valid JSON that simply never ends inside the limit.
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"`))
		chunk := strings.Repeat("A", 64<<10)
		for written := 0; written < (1<<20)+(128<<10); written += len(chunk) {
			if _, err := w.Write([]byte(chunk)); err != nil {
				return
			}
		}
		_, _ = w.Write([]byte(`"}]}}]}`))
	})
	_, _, err := c.Check(context.Background(), evidencecheck.Request{Photos: [][]byte{{1}}, Koli: 1})
	if err == nil {
		t.Fatal("oversized response accepted")
	}
	if !errors.Is(err, evidencecheck.ErrSchema) {
		t.Fatalf("got %v, want it to wrap ErrSchema", err)
	}
}

// stampingTransport adds a bearer token to every request it carries,
// the way the oauth2 transport behind htransport.NewClient does,
// redirect hops included.
type stampingTransport struct{}

func (stampingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer vertex-token")
	return http.DefaultTransport.RoundTrip(r)
}

// Plan 0016 D11: the key goes to Google's host and nowhere else, and
// a token minted for Vertex is never sent elsewhere. net/http forwards
// custom headers such as x-goog-api-key on a redirect, and the oauth2
// transport stamps every hop, so following a redirect would hand the
// credential to whatever host the 3xx names. The client New returns
// must refuse to follow one, for both APIs.
func TestNew_ClientDoesNotFollowRedirects(t *testing.T) {
	var leaked []string
	elsewhere := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leaked = append(leaked, "key="+r.Header.Get("x-goog-api-key")+" auth="+r.Header.Get("Authorization"))
		reply(w, okObservation, "STOP")
	}))
	defer elsewhere.Close()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, elsewhere.URL+r.URL.Path, http.StatusTemporaryRedirect)
	}))
	defer api.Close()

	// Gemini: the client New builds for the real host, with a key. New
	// refuses a custom host with a key, so the URL is pointed at the
	// redirecting server inside the package.
	g, err := New(context.Background(), Config{API: APIGemini, Model: "m", APIKey: "secret-key"})
	if err != nil {
		t.Fatal(err)
	}
	g.url = api.URL + "/v1beta/models/m:generateContent"

	// Vertex: a caller-supplied client whose transport stamps a token.
	callerClient := &http.Client{Transport: stampingTransport{}}
	v, err := New(context.Background(), Config{
		Project: "p", Location: "eu", Model: "m", HTTPClient: callerClient, BaseURL: api.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	for name, c := range map[string]*Checker{"gemini-api": g, "vertex": v} {
		if _, _, err := c.Check(context.Background(), evidencecheck.Request{Photos: [][]byte{{1}}, Koli: 1}); err == nil {
			t.Errorf("%s: a redirected request succeeded", name)
		}
	}
	if len(leaked) > 0 {
		t.Fatalf("credentials reached the redirect target: %v", leaked)
	}
	if callerClient.CheckRedirect != nil {
		t.Error("New modified the caller's http.Client")
	}
}

// Media resolution is per photo part. Left empty, the request carries
// no mediaResolution at all, so the model's default applies and the
// production request is byte-for-byte what it was before the option.
func TestCheck_MediaResolution(t *testing.T) {
	photo := []byte{0xff, 0xd8, 0xff, 0x01}
	send := func(level string) []map[string]any {
		t.Helper()
		var body struct {
			Contents []struct {
				Parts []map[string]any `json:"parts"`
			} `json:"contents"`
		}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Errorf("bad request body: %v", err)
			}
			reply(w, okObservation, "STOP")
		}))
		defer srv.Close()
		c, err := New(context.Background(), Config{
			Project: "proj", Location: "eu", Model: "gemini-test", MediaResolution: level,
			HTTPClient: srv.Client(), BaseURL: srv.URL,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := c.Check(context.Background(), evidencecheck.Request{Photos: [][]byte{photo, photo}, Koli: 1}); err != nil {
			t.Fatal(err)
		}
		return body.Contents[0].Parts
	}

	for _, p := range send("") {
		if _, ok := p["mediaResolution"]; ok {
			t.Fatalf("default request carries mediaResolution: %v", p)
		}
	}

	for level, want := range map[string]string{
		MediaResolutionLow:    "MEDIA_RESOLUTION_LOW",
		MediaResolutionMedium: "MEDIA_RESOLUTION_MEDIUM",
		MediaResolutionHigh:   "MEDIA_RESOLUTION_HIGH",
	} {
		parts := send(level)
		if _, ok := parts[0]["mediaResolution"]; ok {
			t.Errorf("%s: text part carries mediaResolution", level)
		}
		for _, p := range parts[1:] {
			mr, _ := p["mediaResolution"].(map[string]any)
			if mr["level"] != want {
				t.Errorf("%s: photo part mediaResolution = %v, want level %s", level, p["mediaResolution"], want)
			}
		}
	}
}

func TestNew_RejectsUnknownMediaResolution(t *testing.T) {
	for _, api := range []string{APIVertex, APIGemini} {
		_, err := New(context.Background(), Config{
			API: api, Project: "p", Location: "eu", Model: "m", MediaResolution: "ultra",
			HTTPClient: &http.Client{}, BaseURL: "http://127.0.0.1:1",
		})
		if err == nil || !strings.Contains(err.Error(), "media resolution") {
			t.Errorf("%s: err = %v, want a media resolution error", api, err)
		}
	}
}
