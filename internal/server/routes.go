package server

import (
	"github.com/RobBrazier/bookfeed/cmd/web"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
	"github.com/go-chi/traceid"
	"github.com/golang-cz/devslog"
)

func getLogger(isLocal bool) (*slog.Logger, *httplog.Schema) {
	format := httplog.SchemaOTEL.Concise(isLocal)
	handlerOpts := &slog.HandlerOptions{
		AddSource:   !isLocal,
		ReplaceAttr: format.ReplaceAttr,
	}
	var handler slog.Handler
	if isLocal {
		handler = devslog.NewHandler(os.Stdout, &devslog.Options{
			SortKeys:           true,
			MaxErrorStackTrace: 5,
			MaxSlicePrintSize:  20,
			HandlerOptions:     handlerOpts,
		})
	} else {
		handler = traceid.LogHandler(
			slog.NewJSONHandler(os.Stdout, handlerOpts),
		)
	}

	logger := slog.New(handler)

	if !isLocal {
		logger = logger.With(slog.String("service.name", "bookfeed"))
	}
	return logger, format
}

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	isLocal := os.Getenv("APP_ENV") == "local"
	logger, format := getLogger(isLocal)

	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelInfo)

	r.Use(traceid.Middleware)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	if os.Getenv("LOG_REQUESTS") == "true" {
		r.Use(httplog.RequestLogger(logger, &httplog.Options{
			Level:         slog.LevelInfo,
			Schema:        format,
			RecoverPanics: true,
		}))
	}
	r.Use(middleware.Heartbeat("/up"))
	r.Use(middleware.URLFormat)

	MountStatic(r)

	// redirect root to hardcover
	r.Get("/", http.RedirectHandler("/hc", http.StatusTemporaryRedirect).ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 10*time.Second))

		r.Route("/hc", func(r chi.Router) {
			r.Get("/", templ.Handler(web.Hardcover()).ServeHTTP)
			r.Get("/recent", s.RecentHandler)
			r.Get("/author/{author:[a-zA-Z0-9-]+}", s.AuthorHandler)
			r.Get("/series/{series:[a-zA-Z0-9-]+}", s.SeriesHandler)
			r.Get("/me/{username:[a-zA-Z0-9-]+}", s.MeHandler)
		})

		r.Route("/jnc", func(r chi.Router) {
			r.Get("/", templ.Handler(web.JNovelClub()).ServeHTTP)
		})
	})

	return r
}
