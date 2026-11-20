package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestListenAddr(t *testing.T) {
	tests := []struct {
		port string
		want string
	}{
		{port: "", want: ":8080"},
		{port: "   ", want: ":8080"},
		{port: "9090", want: ":9090"},
		{port: ":9090", want: ":9090"},
		{port: " 9090 ", want: ":9090"},
	}

	for _, tt := range tests {
		if got := listenAddr(tt.port); got != tt.want {
			t.Errorf("listenAddr(%q) = %q, want %q", tt.port, got, tt.want)
		}
	}
}

func TestNewServerWithoutFiles(t *testing.T) {
	server := newServer(":8080", "  ")
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestNewServerServesStaticFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("calculator"), 0o644); err != nil {
		t.Fatal(err)
	}

	server := newServer(":8080", dir)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "calculator" {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestHoldOpen(t *testing.T) {
	if os.Getenv("HOLD_SERVER") != "1" {
		t.Skip("set HOLD_SERVER=1 to keep the API listening for the frontend integration test")
	}

	server := newServer(listenAddr(os.Getenv("PORT")), os.Getenv("STATIC_DIR"))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		t.Fatal(err)
	}
}
