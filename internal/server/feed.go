package server

import (
	"log/slog"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/feeds"
)

func writeContentType(mediaType string, w http.ResponseWriter) {
	params := map[string]string{
		"charset": "utf-8",
	}
	contentType := mime.FormatMediaType(mediaType, params)
	w.Header().Set("Content-Type", contentType)
}

func (s *Server) notFound(err error, w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(err.Error()))
}

func (s *Server) writeFeed(format string, out *feeds.Feed, w http.ResponseWriter) {
	// Set Cloudflare cache header for 1 hour (3600 seconds)
	// This is shorter than our data cache (6 hours) to ensure freshness
	w.Header().Set("Cache-Control", "public, max-age=3600")

	switch format {
	case "atom":
		writeContentType("application/atom+xml", w)
		out.WriteAtom(w)
	case "json":
		writeContentType("application/json", w)
		out.WriteJSON(w)
	default:
		writeContentType("application/rss+xml", w)
		out.WriteRss(w)
	}
}

func (s *Server) RecentHandler(w http.ResponseWriter, r *http.Request) {
	format, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	feed, err := s.builder.GetRecentReleases(r.Context())
	if err != nil {
		slog.Error("error retrieving recent", "err", err)
	}
	s.writeFeed(format, &feed, w)
}

func (s *Server) AuthorHandler(w http.ResponseWriter, r *http.Request) {
	format, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	author := strings.ToLower(r.PathValue("author"))
	feed, err := s.builder.GetAuthorReleases(r.Context(), author)

	if err != nil {
		slog.Error("error retrieving author", "err", err)
		s.notFound(err, w)
		return
	}

	s.writeFeed(format, &feed, w)
}

func (s *Server) SeriesHandler(w http.ResponseWriter, r *http.Request) {
	format, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	series := strings.ToLower(r.PathValue("series"))
	feed, err := s.builder.GetSeriesReleases(r.Context(), series)
	if err != nil {
		slog.Error("error retrieving series", "err", err)
		s.notFound(err, w)
		return
	}
	s.writeFeed(format, &feed, w)
}

func (s *Server) MeHandler(w http.ResponseWriter, r *http.Request) {
	format, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	series := strings.ToLower(r.PathValue("username"))
	filter := strings.ToLower(r.URL.Query().Get("filter"))
	feed, err := s.builder.GetUserReleases(r.Context(), series, filter)
	if err != nil {
		slog.Error("error retrieving series", "err", err)
		s.notFound(err, w)
		return
	}
	s.writeFeed(format, &feed, w)
}
