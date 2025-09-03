package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/feeds"
)

func (s *Server) writeFeed(format string, out *feeds.Feed, w http.ResponseWriter) {
	// Set Cloudflare cache header for 1 hour (3600 seconds)
	// This is shorter than our data cache (6 hours) to ensure freshness
	w.Header().Set("Cache-Control", "public, max-age=3600")

	switch format {
	case "atom":
		out.WriteAtom(w)
	case "json":
		out.WriteJSON(w)
	default:
		out.WriteRss(w)
	}
}

func (s *Server) RecentHandler(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(r.PathValue("format"))
	feed, err := s.builder.GetRecentReleases(r.Context())
	if err != nil {
		slog.Error("error retrieving recent", "err", err)
	}
	s.writeFeed(format, &feed, w)
}

func (s *Server) AuthorHandler(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(r.PathValue("format"))
	author := r.PathValue("author")
	feed, err := s.builder.GetAuthorReleases(r.Context(), author)

	if err != nil {
		slog.Error("error retrieving author", "err", err)
	}

	s.writeFeed(format, &feed, w)
}

func (s *Server) SeriesHandler(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(r.PathValue("format"))
	series := r.PathValue("series")
	feed, err := s.builder.GetSeriesReleases(r.Context(), series)
	if err != nil {
		slog.Error("error retrieving series", "err", err)
	}
	s.writeFeed(format, &feed, w)
}
