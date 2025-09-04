package server

import (
	"hardcover-feed/cmd/web"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func MountStatic(r *chi.Mux) {

	staticRoot, err := fs.Sub(web.Static, "static")
	if err != nil {
		log.Fatal(err)
	}

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServerFS(staticRoot)))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(web.Index))
	})
}
