package server

import (
	"bookfeed/cmd/web"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	logFormat := httplog.SchemaOTEL.Concise(true)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: logFormat.ReplaceAttr,
	}))

	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelInfo)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level:         slog.LevelInfo,
		Schema:        httplog.SchemaOTEL,
		RecoverPanics: true,
	}))
	r.Use(middleware.Heartbeat("/up"))
	r.Use(middleware.URLFormat)

	MountStatic(r)

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 10*time.Second))

		r.Route("/hc", func(r chi.Router) {
			r.Get("/", templ.Handler(web.Hardcover()).ServeHTTP)
			r.Get("/recent", s.RecentHandler)
			r.Get("/author/{author:[a-zA-Z0-9-]+}", s.AuthorHandler)
			r.Get("/series/{series:[a-zA-Z0-9-]+}", s.SeriesHandler)
			// r.Get("/me/{username:[a-zA-Z0-9-]+}", s.MeHandler)
		})

		r.Route("/jnc", func(r chi.Router) {
			r.Get("/", templ.Handler(web.JNovelClub()).ServeHTTP)
		})
	})

	return r
}
