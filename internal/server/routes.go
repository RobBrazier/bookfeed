package server

import (
	"net/http"
	"os"
	"time"

	"github.com/RobBrazier/bookfeed/cmd/web"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-chi/traceid"
	"github.com/rs/zerolog/hlog"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(traceid.Middleware)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	if os.Getenv("LOG_REQUESTS") == "true" {
		r.Use(hlog.NewHandler(*s.logger))
		r.Use(hlog.RequestIDHandler("req_id", "Request-Id"))
		r.Use(hlog.MethodHandler("method"))
		r.Use(hlog.URLHandler("url"))
		r.Use(hlog.UserAgentHandler("user_agent"))
		r.Use(hlog.RefererHandler("referer"))
		r.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Send()
		}))
	}
	r.Use(middleware.Heartbeat("/up"))
	r.Use(middleware.URLFormat)

	MountStatic(r)

	// redirect root to hardcover
	r.Handle("/", http.RedirectHandler("/hc", http.StatusTemporaryRedirect))

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 10*time.Second))

		r.Route("/hc", func(r chi.Router) {
			r.With(middleware.NoCache).Handle("/", templ.Handler(web.Hardcover()))
			r.Get("/recent", s.RecentHandler)
			r.Get("/author/{author:[a-zA-Z0-9-]+}", s.AuthorHandler)
			r.Get("/series/{series:[a-zA-Z0-9-]+}", s.SeriesHandler)
			r.Get("/me/{username:[a-zA-Z0-9-]+}", s.MeHandler)
		})

		r.Route("/jnc", func(r chi.Router) {
			r.With(middleware.NoCache).Handle("/", templ.Handler(web.JNovelClub()))
		})
	})

	return r
}
