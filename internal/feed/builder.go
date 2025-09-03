package feed

import (
	"bytes"
	"context"
	"fmt"
	"hardcover-rss/internal/hardcover"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"embed"
	"text/template"

	"github.com/Khan/genqlient/graphql"
	"github.com/Masterminds/sprig/v3"
	"github.com/gorilla/feeds"
	"github.com/maypok86/otter/v2"
)

//go:embed templates/*
var fs embed.FS

type BookContext struct {
	Title string
	Books []hardcover.Book
}

type LoaderFunc = otter.LoaderFunc[string, feeds.Feed]

type builder struct {
	client       graphql.Client
	cache        *otter.Cache[string, feeds.Feed]
	templates    *template.Template
	language     string
	compilations bool
}

type Builder interface {
	GetRecentReleases(ctx context.Context) (feeds.Feed, error)
	GetAuthorReleases(ctx context.Context, author string) (feeds.Feed, error)
	GetSeriesReleases(ctx context.Context, series string) (feeds.Feed, error)
}

func cdnUrl(image hardcover.BookImageImages) string {
	url := url.QueryEscape(image.Url)
	return fmt.Sprintf("https://img.hardcover.app/enlarge?url=%s&width=%d&height=%d&type=webp", url, image.Width, image.Height)
}

func (b *builder) buildFeed(title, link string, books []hardcover.Book) (feeds.Feed, error) {
	feed := &feeds.Feed{
		Title:   title,
		Link:    &feeds.Link{Href: link},
		Created: time.Now(),
		Updated: time.Now(),
	}
	for _, book := range books {
		var authorName string
		if len(book.Contributions) > 0 {
			authorName = book.Contributions[0].Author.Name
		}
		var enclosure *feeds.Enclosure
		if url := book.Image.Url; url != "" {
			enclosure = &feeds.Enclosure{
				Url:  cdnUrl(book.Image),
				Type: "image/webp",
			}
		}

		item := &feeds.Item{
			Id:        strconv.Itoa(book.Id),
			Title:     book.Title,
			Link:      &feeds.Link{Href: fmt.Sprintf("https://hardcover.app/books/%s", book.Slug)},
			Author:    &feeds.Author{Name: authorName},
			Content:   b.renderContent(book),
			Created:   book.ReleaseDate,
			Enclosure: enclosure,
		}
		feed.Add(item)
	}

	return *feed, nil
}

func (b *builder) renderContent(book hardcover.Book) string {
	var buffer bytes.Buffer
	if book.Image.Width != 0 && book.Image.Height != 0 {
		book.Image.Ratio = float32(book.Image.Width) / float32(book.Image.Height)
		book.Image.Width = int(500 * book.Image.Ratio)
		book.Image.Height = 500
		book.Image.Url = cdnUrl(book.Image)
	}
	b.templates.ExecuteTemplate(&buffer, "content.tmpl", book)
	return buffer.String()
}

func (b *builder) GetRecentReleases(ctx context.Context) (feeds.Feed, error) {
	loader := LoaderFunc(func(ctx context.Context, key string) (feeds.Feed, error) {
		now := time.Now()
		lastMonth := now.AddDate(0, -1, 0)
		data, err := hardcover.RecentReleases(ctx, b.client, now, lastMonth, "")
		if err != nil {
			return feeds.Feed{}, err
		}
		title := "Hardcover: Recent Releases"
		url := "https://hardcover.app/upcoming/recent"
		return b.buildFeed(title, url, data.Books)
	})
	return b.cache.Get(ctx, "releases", loader)
}

func (b *builder) GetAuthorReleases(ctx context.Context, slug string) (feeds.Feed, error) {
	loader := LoaderFunc(func(ctx context.Context, key string) (feeds.Feed, error) {
		now := time.Now()
		authorSlug := strings.Split(key, "/")[1]
		data, err := hardcover.RecentAuthorReleases(ctx, b.client, now, slug, b.language, b.compilations)
		if err != nil {
			return feeds.Feed{}, err
		}
		var books []hardcover.Book
		var authorName string
		for _, contribution := range data.Contributions {
			if authorName == "" {
				authorName = contribution.Author.Name
			}
			books = append(books, contribution.Book)
		}
		title := fmt.Sprintf("Hardcover Author Releases: %s", authorName)
		url := fmt.Sprintf("https://hardcover.app/authors/%s", authorSlug)
		return b.buildFeed(title, url, books)
	})
	return b.cache.Get(ctx, fmt.Sprintf("author/%s", slug), loader)
}

func (b *builder) GetSeriesReleases(ctx context.Context, slug string) (feeds.Feed, error) {
	loader := LoaderFunc(func(ctx context.Context, key string) (feeds.Feed, error) {
		now := time.Now()
		data, err := hardcover.RecentSeriesReleases(ctx, b.client, now, slug, b.language, b.compilations)
		if err != nil {
			return feeds.Feed{}, err
		}
		var books []hardcover.Book
		var seriesName string
		if len(data.Series) > 0 {
			seriesName = data.Series[0].Name
		}
		for _, series := range data.BookSeries {
			books = append(books, series.Book)
		}

		title := fmt.Sprintf("Hardcover Series Releases: %s", seriesName)
		url := fmt.Sprintf("https://hardcover.app/series/%s", slug)
		return b.buildFeed(title, url, books)
	})
	return b.cache.Get(ctx, fmt.Sprintf("series/%s", slug), loader)
}

func newCache() *otter.Cache[string, feeds.Feed] {
	return otter.Must(&otter.Options[string, feeds.Feed]{
		MaximumSize:      10_000,
		ExpiryCalculator: otter.ExpiryCreating[string, feeds.Feed](6 * time.Hour),
	})
}

func NewBuilder() Builder {
	token := os.Getenv("HARDCOVER_TOKEN")
	client := hardcover.GetClient(token)
	return &builder{
		client: client,
		cache:  newCache(),
		templates: template.Must(
			template.New("base").Funcs(sprig.FuncMap()).ParseFS(fs, "templates/*.tmpl"),
		),
		language:     "eng",
		compilations: false,
	}
}
