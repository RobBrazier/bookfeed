package server

import (
	"log/slog"
	"net/http"
	"os"
	"time"

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

		r.Get("/recent", s.RecentHandler)
		r.Get("/author/{author:[a-z0-9-]+}", s.AuthorHandler)
		r.Get("/series/{series:[a-z0-9-]+}", s.SeriesHandler)
	})

	return r
}
