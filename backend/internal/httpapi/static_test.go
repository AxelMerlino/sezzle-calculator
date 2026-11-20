package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStaticSite(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "static")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "secret.txt"), []byte("classified-value"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("calculator"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("console.log(1)"), 0o644); err != nil {
		t.Fatal(err)
	}

	mux := NewMux()
	AttachStatic(mux, dir)

	index := get(t, mux, "/")
	if index.Code != http.StatusOK || !strings.Contains(index.Body.String(), "calculator") {
		t.Fatalf("index status=%d body=%s", index.Code, index.Body.String())
	}

	asset := get(t, mux, "/assets/app.js")
	if asset.Code != http.StatusOK || asset.Body.String() != "console.log(1)" {
		t.Fatalf("asset status=%d body=%s", asset.Code, asset.Body.String())
	}
	if !strings.Contains(asset.Header().Get("Content-Type"), "javascript") {
		t.Fatalf("Content-Type = %q", asset.Header().Get("Content-Type"))
	}

	fallback := get(t, mux, "/not-a-page")
	if fallback.Code != http.StatusOK || !strings.Contains(fallback.Body.String(), "calculator") {
		t.Fatalf("fallback status=%d body=%s", fallback.Code, fallback.Body.String())
	}

	missingAsset := get(t, mux, "/assets/missing.js")
	if missingAsset.Code != http.StatusNotFound {
		t.Fatalf("missing asset status=%d", missingAsset.Code)
	}

	escape := get(t, mux, "/../secret.txt")
	if strings.Contains(escape.Body.String(), "classified-value") {
		t.Fatalf("escaped the static directory: %s", escape.Body.String())
	}

	api := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/calculate", strings.NewReader(`{"operation":"add","a":2,"b":2}`))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(api, req)
	if api.Code != http.StatusOK || !strings.Contains(api.Body.String(), `"result":4`) {
		t.Fatalf("api status=%d body=%s", api.Code, api.Body.String())
	}
}

func TestAttachStaticIgnoresBlankDir(t *testing.T) {
	mux := NewMux()
	AttachStatic(mux, "  ")
	rec := get(t, mux, "/")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestIsInside(t *testing.T) {
	root := filepath.Join(t.TempDir(), "static")
	if !isInside(root, filepath.Join(root, "index.html")) {
		t.Fatal("expected a file inside the root to be allowed")
	}
	outside := filepath.Join(root, "..", "secret.txt")
	if isInside(root, outside) {
		t.Fatal("expected a path outside the root to be rejected")
	}
}

func get(t *testing.T, handler http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
