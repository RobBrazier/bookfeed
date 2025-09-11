package feed

import (
	"context"
	"fmt"
	"github.com/RobBrazier/bookfeed/internal/hardcover"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"slices"
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

type LoaderFunc = otter.LoaderFunc[string, feeds.Feed]

type UserInterests struct {
	Authors []string
	Series  []string
}

type builder struct {
	client       graphql.Client
	feed         *otter.Cache[string, feeds.Feed]
	user         *otter.Cache[string, UserInterests]
	templates    *template.Template
	compilations bool
}

type Builder interface {
	GetRecentReleases(ctx context.Context) (feeds.Feed, error)
	GetAuthorReleases(ctx context.Context, author string) (feeds.Feed, error)
	GetSeriesReleases(ctx context.Context, series string) (feeds.Feed, error)
	GetUserReleases(ctx context.Context, username, filter string) (feeds.Feed, error)
}

func cdnUrl(image hardcover.BookImage) string {
	url := url.QueryEscape(image.Url)
	return fmt.Sprintf("https://img.hardcover.app/enlarge?url=%s&width=%d&height=%d&type=webp", url, image.Width, image.Height)
}

func (b *builder) buildFeed(title, link, description string, books []hardcover.Book) (feeds.Feed, error) {
	created := time.Now()
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
	feed.Sort(func(a, b *feeds.Item) bool {
		return b.Created.Before(a.Created)
	})

	return *feed, nil
}

func (b *builder) renderContent(book hardcover.Book) string {
	var builder strings.Builder
	if book.Image.Width != 0 && book.Image.Height != 0 {
		ratio := float32(book.Image.Width) / float32(book.Image.Height)
		book.Image.Width = int(500 * ratio)
		book.Image.Height = 500
		book.Image.Url = cdnUrl(book.Image)
	}
	b.templates.ExecuteTemplate(&builder, "content.tmpl", book)
	return builder.String()
}

func (b *builder) GetRecentReleases(ctx context.Context) (feeds.Feed, error) {
	loader := LoaderFunc(func(ctx context.Context, key string) (feeds.Feed, error) {
		now := time.Now()
		lastMonth := now.AddDate(0, -1, 0)
		slog.Info("Fetching recent releases", "now", now, "earliest", lastMonth)
		data, err := hardcover.RecentReleases(ctx, b.client, now, lastMonth)
		slog.Info("Retrieved recent releases data", "elapsed", time.Since(now))
		if err != nil {
			return feeds.Feed{}, err
		}
		title := "Hardcover: Recent Releases"
		url := "https://hardcover.app/upcoming/recent"
		return b.buildFeed(title, url, "", data.Books)
	})
	return b.feed.Get(ctx, "releases", loader)
}

func (b *builder) GetAuthorReleases(ctx context.Context, slug string) (feeds.Feed, error) {
	loader := LoaderFunc(func(ctx context.Context, key string) (feeds.Feed, error) {
		now := time.Now()
		lastYear := now.AddDate(-1, 0, 0)
		slog.Info("Fetching releases", "author", slug, "now", now, "earliest", lastYear)
		data, err := hardcover.RecentAuthorReleases(ctx, b.client, now, lastYear, []string{slug}, b.compilations)
		slog.Info("Retrieved author data", "author", slug, "elapsed", time.Since(now))
		if err != nil {
			return feeds.Feed{}, err
		}
		var books []hardcover.Book
		if len(data.Authors) == 0 {
			return feeds.Feed{}, fmt.Errorf("Author not found")
		}
		author := data.Authors[0]
		authorName := author.Name
		for _, contribution := range author.Contributions {
			books = append(books, contribution.Book)
		}
		title := fmt.Sprintf("Hardcover Author Releases: %s", authorName)
		url := fmt.Sprintf("https://hardcover.app/authors/%s", slug)
		return b.buildFeed(title, url, "", books)
	})
	return b.feed.Get(ctx, fmt.Sprintf("author/%s", slug), loader)
}

func (b *builder) GetSeriesReleases(ctx context.Context, slug string) (feeds.Feed, error) {
	loader := LoaderFunc(func(ctx context.Context, key string) (feeds.Feed, error) {
		now := time.Now()
		lastYear := now.AddDate(-1, 0, 0)
		slog.Info("Fetching releases", "series", slug, "now", now, "earliest", lastYear)
		data, err := hardcover.RecentSeriesReleases(ctx, b.client, now, lastYear, []string{slug}, b.compilations)
		slog.Info("Retrieved series data", "series", slug, "elapsed", time.Since(now))
		if err != nil {
			return feeds.Feed{}, err
		}
		if len(data.Series) == 0 {
			return feeds.Feed{}, fmt.Errorf("Series not found")
		}
		var books []hardcover.Book
		seriesName := data.Series[0].Name
		for _, series := range data.BookSeries {
			books = append(books, series.Book)
		}

		title := fmt.Sprintf("Hardcover Series Releases: %s", seriesName)
		url := fmt.Sprintf("https://hardcover.app/series/%s", slug)
		return b.buildFeed(title, url, "", books)
	})
	return b.feed.Get(ctx, fmt.Sprintf("series/%s", slug), loader)
}

func (b *builder) getUserInterests(ctx context.Context, username string) (UserInterests, error) {
	loader := otter.LoaderFunc[string, UserInterests](func(ctx context.Context, key string) (UserInterests, error) {
		earliest := time.Now().AddDate(-2, 0, 0)
		data, err := hardcover.UserInterests(ctx, b.client, username, earliest)
		if err != nil {
			return UserInterests{}, err
		}
		if len(data.Users) == 0 {
			return UserInterests{}, fmt.Errorf("User not found")
		}
		authorCount := make(map[string]int)
		seriesCount := make(map[string]int)
		for _, book := range data.UserBooks {
			for _, contribution := range book.Book.Contributors {
				if slices.Contains([]string{"author", ""}, strings.ToLower(contribution.Contribution)) {
					authorCount[contribution.Author.Slug]++
				}
			}
			if slug := book.Book.FeaturedSeries.Series.Slug; slug != "" {
				seriesCount[slug]++
			}
		}
		var authors []string
		var series []string

		// only check feeds for authors that have > 1 book read
		for slug, count := range authorCount {
			if count > 1 {
				authors = append(authors, slug)
			}
		}
		for slug, count := range seriesCount {
			if count > 1 {
				series = append(series, slug)
			}
		}

		return UserInterests{
			Series:  series,
			Authors: authors,
		}, nil
	})
	return b.user.Get(ctx, fmt.Sprintf("user/%s", username), loader)
}

func (b *builder) GetUserReleases(ctx context.Context, username, filter string) (feeds.Feed, error) {
	interests, err := b.getUserInterests(ctx, username)
	if err != nil {
		return feeds.Feed{}, err
	}

	slog.Info("user interests", "authors", interests.Authors, "series", interests.Series)

	loader := LoaderFunc(func(ctx context.Context, key string) (feeds.Feed, error) {
		now := time.Now()
		earliest := now.AddDate(0, -3, 0)
		bookMap := make(map[int]hardcover.Book)
		if slices.Contains([]string{"", "series"}, filter) && len(interests.Series) > 0 {
			series, err := hardcover.RecentSeriesReleases(ctx, b.client, now, earliest, interests.Series, b.compilations)
			if err != nil {
				return feeds.Feed{}, err
			}
			for _, book := range series.BookSeries {
				if _, ok := bookMap[book.Book.Id]; !ok {
					bookMap[book.Book.Id] = book.Book
				}
			}
		}
		if slices.Contains([]string{"", "author"}, filter) && len(interests.Authors) > 0 {
			author, err := hardcover.RecentAuthorReleases(ctx, b.client, now, earliest, interests.Authors, b.compilations)
			if err != nil {
				return feeds.Feed{}, err
			}
			for _, author := range author.Authors {
				for _, book := range author.Contributions {
					if _, ok := bookMap[book.Book.Id]; !ok {
						bookMap[book.Book.Id] = book.Book
					}
				}
			}
		}

		books := slices.Collect(maps.Values(bookMap))

		title := fmt.Sprintf("Hardcover User Releases: %s", username)
		url := fmt.Sprintf("https://hardcover.app/@%s", username)
		var builder strings.Builder
		builder.WriteString("Includes New Releases from:\n")
		if filter == "" || filter == "author" {
			builder.WriteString(fmt.Sprintf("Authors: %s\n", strings.Join(interests.Authors, ", ")))
		}
		if filter == "" || filter == "series" {
			builder.WriteString(fmt.Sprintf("Series: %s\n", strings.Join(interests.Series, ", ")))
		}
		return b.buildFeed(title, url, builder.String(), books)
	})
	return b.feed.Get(ctx, fmt.Sprintf("user/%s/%s", username, filter), loader)
}

func feedCache() *otter.Cache[string, feeds.Feed] {
	return otter.Must(&otter.Options[string, feeds.Feed]{
		MaximumSize:      10_000,
		ExpiryCalculator: otter.ExpiryCreating[string, feeds.Feed](12 * time.Hour),
	})
}

func userCache() *otter.Cache[string, UserInterests] {
	return otter.Must(&otter.Options[string, UserInterests]{
		MaximumSize:      10_000,
		ExpiryCalculator: otter.ExpiryCreating[string, UserInterests](24 * time.Hour),
	})
}

func NewBuilder() Builder {
	token := os.Getenv("HARDCOVER_TOKEN")
	client := hardcover.GetClient(token)
	return &builder{
		client: client,
		feed:   feedCache(),
		user:   userCache(),
		templates: template.Must(
			template.New("base").Funcs(sprig.FuncMap()).ParseFS(fs, "templates/*.tmpl"),
		),
		compilations: false,
	}
}
