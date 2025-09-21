package server

import (
	"github.com/RobBrazier/bookfeed/assets"
	"github.com/rs/zerolog/log"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
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
		w.Write([]byte(assets.RobotsTxt))
	})

}
