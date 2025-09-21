package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/RobBrazier/bookfeed/config"
	"github.com/RobBrazier/bookfeed/internal/view/pages"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/zerolog/hlog"
)

func formatPath(path string) string {
	formats := []string{"json", "atom", "rss"}
	regex := strings.Join(formats, "|")
	return fmt.Sprintf("%s.{format:(%s)}", path, regex)
}

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	if config.LogRequests() {
		r.Use(hlog.NewHandler(*s.logger))
		r.Use(hlog.RequestIDHandler("req_id", middleware.RequestIDHeader))
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

	MountStatic(r)

	// redirect root to hardcover
	r.Handle("/", http.RedirectHandler("/hc", http.StatusTemporaryRedirect))

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 10*time.Second))

		r.Route("/hc", func(r chi.Router) {
			r.With(middleware.NoCache).Handle("/", templ.Handler(pages.Hardcover()))
			r.Get(formatPath("/recent"), s.RecentHandler)
			r.Get(formatPath("/author/{author:[a-zA-Z0-9-]+}"), s.AuthorHandler)
			r.Get(formatPath("/series/{series:[a-zA-Z0-9-]+}"), s.SeriesHandler)
			r.Get(formatPath("/me/{username:[a-zA-Z0-9-]+}"), s.MeHandler)
		})

		r.Route("/jnc", func(r chi.Router) {
			r.With(middleware.NoCache).Handle("/", templ.Handler(pages.JNovelClub()))
		})
	})

	return r
}
