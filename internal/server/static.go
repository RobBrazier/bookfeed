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

}
