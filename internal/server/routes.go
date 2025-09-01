package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", s.HelloWorldHandler)

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 10*time.Second))

		r.Get("/recent", s.RecentHandler)
		r.Get("/recent/{format:[a-z]+}", s.RecentHandler)
		r.Get("/recent.{format:[a-z]+}", s.RecentHandler)

		r.Get("/author/{author:[a-z0-9-]+}", s.AuthorHandler)
		r.Get("/author/{author:[a-z0-9-]+}/{format:[a-z]+}", s.AuthorHandler)
		r.Get("/author/{author:[a-z0-9-]+}.{format:[a-z]+}", s.AuthorHandler)

		r.Get("/series/{series:[a-z0-9-]+}", s.SeriesHandler)
		r.Get("/series/{series:[a-z0-9-]+}/{format:[a-z]+}", s.SeriesHandler)
		r.Get("/series/{series:[a-z0-9-]+}.{format:[a-z]+}", s.SeriesHandler)
	})

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
