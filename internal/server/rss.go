package server

import (
	"fmt"
	"hardcover-rss/internal/feed"
	"log/slog"
	"net/http"

	"github.com/gorilla/feeds"
)

func (s *Server) determineFormat(r *http.Request) feed.Format {
	format := r.PathValue("format")
	switch format {
	case "rss":
		return feed.FORMAT_RSS
	case "json":
		return feed.FORMAT_JSON
	case "atom":
		return feed.FORMAT_ATOM
	default:
		return feed.FORMAT_RSS
	}
}

func (s *Server) writeFeed(format feed.Format, out *feeds.Feed, w http.ResponseWriter) error {
	switch format {
	case feed.FORMAT_RSS:
		out.WriteRss(w)
	case feed.FORMAT_ATOM:
		out.WriteAtom(w)
	case feed.FORMAT_JSON:
		out.WriteJSON(w)
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func (s *Server) RecentHandler(w http.ResponseWriter, r *http.Request) {
	format := s.determineFormat(r)
	feed, err := s.builder.GetRecentReleases(r.Context())
	if err != nil {
		slog.Error("error retrieving recent", "err", err)
	}
	s.writeFeed(format, feed, w)
}

func (s *Server) AuthorHandler(w http.ResponseWriter, r *http.Request) {
	format := s.determineFormat(r)
	author := r.PathValue("author")
	feed, err := s.builder.GetAuthorReleases(r.Context(), author)

	if err != nil {
		slog.Error("error retrieving author", "err", err)
	}
	s.writeFeed(format, feed, w)
}

func (s *Server) SeriesHandler(w http.ResponseWriter, r *http.Request) {
	format := s.determineFormat(r)
	series := r.PathValue("series")
	feed, err := s.builder.GetSeriesReleases(r.Context(), series)
	if err != nil {
		slog.Error("error retrieving series", "err", err)
	}
	s.writeFeed(format, feed, w)
}
