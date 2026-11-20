package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"calculator/internal/httpapi"
)

func main() {
	staticDir := strings.TrimSpace(os.Getenv("STATIC_DIR"))
	addr := listenAddr(os.Getenv("PORT"))
	if staticDir != "" {
		log.Printf("serving files from %s", staticDir)
	}
	log.Printf("listening on %s", addr)
	if err := newServer(addr, staticDir).ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func newServer(addr, staticDir string) *http.Server {
	mux := httpapi.NewMux()
	if strings.TrimSpace(staticDir) != "" {
		httpapi.AttachStatic(mux, staticDir)
	}
	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
	}
}

func listenAddr(port string) string {
	port = strings.TrimSpace(port)
	if port == "" {
		return ":8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}
