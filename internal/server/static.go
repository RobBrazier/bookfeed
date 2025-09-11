package server

import (
	"github.com/RobBrazier/bookfeed/cmd/web"
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

	// Because of URLFormat middleware this gets mapped to robots.txt... and robots.anything
	r.Get("/robots", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(web.RobotsTxt))
	})

}
