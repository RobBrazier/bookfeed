package feed

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/RobBrazier/bookfeed/internal/model"

	"embed"
	"text/template"

	"github.com/gorilla/feeds"
)

//go:embed templates/*
var fs embed.FS

type builder struct {
	templates *template.Template
}

type Builder interface {
	GetRecentReleases(ctx context.Context) (feeds.Feed, error)
	GetAuthorReleases(ctx context.Context, author string) (feeds.Feed, error)
	GetSeriesReleases(ctx context.Context, series string) (feeds.Feed, error)
	GetUserReleases(ctx context.Context, username, filter string) (feeds.Feed, error)
}

func (b *builder) buildFeed(title, link, description string, created time.Time, books []model.Book) (feeds.Feed, error) {
	if description != "" {
		description = "\n" + description
	}
	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: link},
		Created:     created,
		Description: fmt.Sprintf("Generated on %s%s", created.Format("02 Jan 2006 15:04:05 (-0700)"), description),
		Updated:     created,
	}
	for _, book := range books {
		slog.Info("Processing book", "book", book)
		var authorName string
		if len(book.Authors) > 0 {
			authorName = book.Authors[0]
		}
		var enclosure *feeds.Enclosure
		if url := book.Image.Url; url != "" {
			enclosure = &feeds.Enclosure{
				Url:  book.Image.Url,
				Type: "image/webp",
			}
		}

		item := &feeds.Item{
			Id:        strconv.Itoa(book.Id),
			Title:     book.Title,
			Link:      &feeds.Link{Href: book.Link},
			Author:    &feeds.Author{Name: authorName},
			Content:   b.renderContent(book),
			Created:   book.ReleaseDate,
			Enclosure: enclosure,
		}
		feed.Add(item)
	}
	feed.Sort(func(a, b *feeds.Item) bool {
		return b.Created.Before(a.Created)
	})

	return *feed, nil
}

func (b *builder) renderContent(book model.Book) string {
	var builder strings.Builder
	b.templates.ExecuteTemplate(&builder, "content.tmpl", book)
	return builder.String()
}
