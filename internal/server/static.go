package server

import (
	"io/fs"
	"net/http"

	"github.com/RobBrazier/bookfeed/assets"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func MountStatic(r *chi.Mux) {
	staticRoot, err := fs.Sub(assets.Static, "build")
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't extract static assets from embedded FS")
	}

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServerFS(staticRoot)))

	// Because of URLFormat middleware this gets mapped to robots.txt... and robots.anything
	r.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(assets.RobotsTxt))
	})
}
