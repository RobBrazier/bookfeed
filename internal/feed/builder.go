package feed

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/RobBrazier/bookfeed/internal/model"
	"github.com/RobBrazier/bookfeed/internal/view"
	"github.com/RobBrazier/bookfeed/internal/view/feed"
	"github.com/a-h/templ"
	"github.com/gorilla/feeds"
	"github.com/rs/zerolog/log"
)

type builder struct {
	provider view.ProviderData
}

type Builder interface {
	GetRecentReleases(ctx context.Context) (feeds.Feed, error)
	GetAuthorReleases(ctx context.Context, author string) (feeds.Feed, error)
	GetSeriesReleases(ctx context.Context, series string) (feeds.Feed, error)
	GetUserReleases(ctx context.Context, username, filter string) (feeds.Feed, error)
}

func (b *builder) buildFeed(
	ctx context.Context,
	title, link, description string,
	created time.Time,
	books []model.Book,
) (feeds.Feed, error) {
	if description != "" {
		description = "\n" + description
	}
	feed := &feeds.Feed{
		Title:   title,
		Link:    &feeds.Link{Href: link},
		Created: created,
		Description: fmt.Sprintf(
			"Generated on %s%s",
			created.Format("02 Jan 2006 15:04:05 (-0700)"),
			description,
		),
		Updated: created,
	}
	for _, book := range books {
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
		content, err := b.renderContent(ctx, book)
		if err != nil {
			log.Error().
				Err(err).
				Interface("book", book).
				Str("feed", title).
				Msg("Unable to render feed for book")
			continue
		}

		item := &feeds.Item{
			Id:        strconv.Itoa(book.Id),
			Title:     book.Title,
			Link:      &feeds.Link{Href: book.Link},
			Author:    &feeds.Author{Name: authorName},
			Content:   content,
			Created:   book.ReleaseDate,
			Enclosure: enclosure,
		}
		feed.Add(item)
	}
	feed.Sort(func(a, b *feeds.Item) bool {
		return b.Created.Before(a.Created)
	})

	// Limit feed result size
	maxItems := 25
	if len(feed.Items) > maxItems {
		feed.Items = feed.Items[:maxItems]
	}

	return *feed, nil
}

func (b *builder) renderContent(ctx context.Context, book model.Book) (string, error) {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)
	if err := feed.Feed(book, b.provider).Render(ctx, buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
