package server

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"mime"
	"net/http"
	"strings"
	"time"

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
	w.Header().Set("Last-Modified", out.Created.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	cacheExpiry := out.Created.Add(12 * time.Hour)
	remaining := cacheExpiry.Sub(time.Now().UTC())
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(remaining.Seconds())))

	switch format {
	case "rss":
		writeContentType("application/rss+xml", w)
		out.WriteRss(w)
	case "json":
		writeContentType("application/json", w)
		out.WriteJSON(w)
	default:
		writeContentType("application/atom+xml", w)
		out.WriteAtom(w)
	}
}

func (s *Server) RecentHandler(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(r.PathValue("format"))
	feed, err := s.builder.GetRecentReleases(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("error retrieving recent")
	}
	log.Info().Int("entries", len(feed.Items)).Msg("Generated feed for recent releases")
	s.writeFeed(format, &feed, w)
}

func (s *Server) AuthorHandler(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(r.PathValue("format"))
	author := strings.ToLower(r.PathValue("author"))
	log := log.With().Str("author", author).Logger()
	feed, err := s.builder.GetAuthorReleases(r.Context(), author)

	if err != nil {
		log.Error().Err(err).Msg("error retrieving author")
		s.notFound(err, w)
		return
	}
	log.Info().Int("entries", len(feed.Items)).Msg("Generated feed for author")
	s.writeFeed(format, &feed, w)
}

func (s *Server) SeriesHandler(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(r.PathValue("format"))
	series := strings.ToLower(r.PathValue("series"))
	log := log.With().Str("series", series).Logger()
	feed, err := s.builder.GetSeriesReleases(r.Context(), series)
	if err != nil {
		log.Error().Err(err).Msg("error retrieving series")
		s.notFound(err, w)
		return
	}
	log.Info().Int("entries", len(feed.Items)).Msg("Generated feed for series")
	s.writeFeed(format, &feed, w)
}

func (s *Server) MeHandler(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(r.PathValue("format"))
	user := strings.ToLower(r.PathValue("username"))
	filter := strings.ToLower(r.URL.Query().Get("filter"))
	log := log.With().Str("user", user).Str("filter", filter).Logger()
	feed, err := s.builder.GetUserReleases(r.Context(), user, filter)
	if err != nil {
		log.Error().Err(err).Msg("error retrieving user")
		s.notFound(err, w)
		return
	}
	log.Info().Int("entries", len(feed.Items)).Msg("Generated feed for user")
	s.writeFeed(format, &feed, w)
}
