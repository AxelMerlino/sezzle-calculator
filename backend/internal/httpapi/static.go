package httpapi

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// AttachStatic serves a built frontend from dir.
// The API routes stay on this mux; only paths they do not claim reach the files.
// A missing extensionless path returns index.html. A missing file, such as an
// asset, stays a 404 so the browser does not treat HTML as JavaScript.
func AttachStatic(mux *http.ServeMux, dir string) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return
	}
	mux.Handle("/", staticSite(dir))
}

func staticSite(dir string) http.Handler {
	root, err := filepath.Abs(dir)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "static files are unavailable", http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		target := filepath.Join(root, filepath.FromSlash(rel))
		if !isInside(root, target) {
			http.NotFound(w, r)
			return
		}

		info, err := os.Stat(target)
		if err != nil || info.IsDir() {
			if path.Ext(rel) != "" {
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, filepath.Join(root, "index.html"))
			return
		}
		http.ServeFile(w, r, target)
	})
}

func isInside(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
