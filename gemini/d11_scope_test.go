package gemini

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// D11 lets the Phase 0 eval use the Gemini Developer API because its
// free tier costs nothing. The price is that Google may use what is
// sent there and human reviewers may read it, so it is sound only for
// staged photos. Phase 1 handles real sellers' packing photos and must
// use Vertex.
//
// That rule was carried by a doc comment, which is not a control: the
// server will import this package, and a config value or an env var
// could select the Gemini backend in production, sending real evidence
// to a tier that may train on it. The privacy rewrite is a
// non-negotiable gate in the umbrella CLAUDE.md, and this is the kind
// of thing it exists to stop.
//
// Rather than build a Phase 1 guard that has nothing to guard yet, this
// pins the rule that holds today: only the eval may name the Gemini
// backend. It fails the moment someone wires it in anywhere else, which
// is the moment the decision needs a human. Same shape as the CI job
// that greps the selfhost binary for AdminHandlers.
//
// If a second caller is ever legitimate, add it here deliberately and
// say in the plan why sending photos there is sound.
func TestGeminiBackendIsNamedByTheEvalOnly(t *testing.T) {
	root := moduleRoot(t)
	allowed := []string{
		"gemini",                       // the implementation
		filepath.Join("cmd", "eval"),   // the offline eval on staged photos
		filepath.Join("cmd", "server"), // the public demo: staged samples only
	}

	var offenders []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".claude", "node_modules", "vendor", "web", "dist":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		for _, a := range allowed {
			if strings.HasPrefix(rel, a+string(filepath.Separator)) {
				return nil
			}
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// The constant's own name, and its value as a literal, so
		// `API: "gemini-api"` is caught as well as `vertex.APIGemini`.
		if strings.Contains(string(b), "APIGemini") || strings.Contains(string(b), `"gemini-api"`) {
			offenders = append(offenders, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(offenders) > 0 {
		t.Fatalf("the Gemini Developer API backend is named outside the eval and the demo: %v\n"+
			"D11 allows it for staged photos only. Production photos must go to Vertex;\n"+
			"if this caller is legitimate, add it to the allowlist in this test and record why in docs/decisions.md.",
			offenders)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the test's working directory")
		}
		dir = parent
	}
}
