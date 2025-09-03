package server

import (
	"log/slog"
	"mime"
	"net/http"

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

func (s *Server) writeFeed(format string, out *feeds.Feed, w http.ResponseWriter) {
	// Set Cloudflare cache header for 1 hour (3600 seconds)
	// This is shorter than our data cache (6 hours) to ensure freshness
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if len(out.Items) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
		return
	}

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
	author := r.PathValue("author")
	feed, err := s.builder.GetAuthorReleases(r.Context(), author)

	if err != nil {
		slog.Error("error retrieving author", "err", err)
	}

	s.writeFeed(format, &feed, w)
}

func (s *Server) SeriesHandler(w http.ResponseWriter, r *http.Request) {
	format, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	series := r.PathValue("series")
	feed, err := s.builder.GetSeriesReleases(r.Context(), series)
	if err != nil {
		slog.Error("error retrieving series", "err", err)
	}
	s.writeFeed(format, &feed, w)
}
