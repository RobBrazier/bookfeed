package feed

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/Masterminds/sprig/v3"
	"github.com/RobBrazier/bookfeed/internal/cache"
	"github.com/RobBrazier/bookfeed/internal/hardcover"
	"github.com/RobBrazier/bookfeed/internal/model"
	"github.com/gorilla/feeds"
)

type hardcoverBuilder struct {
	builder
	client       graphql.Client
	compilations bool
}

func (b hardcoverBuilder) cdnUrl(image hardcover.BookImage) string {
	url := url.QueryEscape(image.Url)
	return fmt.Sprintf("https://img.hardcover.app/enlarge?url=%s&width=%d&height=%d&type=webp", url, image.Width, image.Height)
}

func (b hardcoverBuilder) mapBook(source hardcover.Book) model.Book {
	var genres []string
	var authors []string
	for _, genre := range source.Genres {
		genres = append(genres, genre.Tag)
	}
	for _, author := range source.Contributions {
		authors = append(authors, author.Author.Name)
	}
	image := model.Image{
		Url: source.Image.Url,
	}
	if source.Image.Width != 0 && source.Image.Height != 0 {
		ratio := float32(source.Image.Width) / float32(source.Image.Height)
		image.Width = int(500 * ratio)
		image.Height = 500
		image.Url = b.cdnUrl(source.Image)
	}
	return model.Book{
		Id:          source.Id,
		Slug:        source.Slug,
		Link:        fmt.Sprintf("https://hardcover.app/books/%s", source.Slug),
		Title:       source.Title,
		ReleaseDate: source.ReleaseDate,
		Headline:    source.Headline,
		Description: source.Description,
		Compilation: source.Compilation,
		Image:       image,
		Authors:     authors,
		Genres:      genres,
		Series: model.Series{
			Title:    source.FeaturedSeries.Series.Name,
			Position: source.FeaturedSeries.Position,
		},
	}
}

func (b hardcoverBuilder) mapBooks(source []hardcover.Book) (books []model.Book) {
	for _, book := range source {
		books = append(books, b.mapBook(book))
	}
	return books
}

func (b hardcoverBuilder) buildUrl(slug string) string {
	return fmt.Sprintf("https://hardcover.app/%s", slug)
}

func (b *hardcoverBuilder) GetRecentReleases(ctx context.Context) (feeds.Feed, error) {
	loader := cache.CollectionLoaderFunc(func(ctx context.Context, key string) (collection model.Collection, err error) {
		now := time.Now()
		lastMonth := now.AddDate(0, -1, 0)
		slog.Info("Fetching recent releases")
		data, err := hardcover.RecentReleases(ctx, b.client, now, lastMonth)
		slog.Info("Retrieved recent releases data", "elapsed", time.Since(now))
		if err != nil {
			return
		}
		books := b.mapBooks(data.Books)
		return model.NewCollection("Recent", "upcoming/recent", books), nil
	})
	collection, err := cache.CollectionCache.Get(ctx, "hardcover/releases", loader)
	if err != nil {
		return feeds.Feed{}, err
	}
	return b.buildFeed("Hardcover: Recent Releases", b.buildUrl(collection.Slug), "", collection.Created, collection.Books)
}

func (b *hardcoverBuilder) GetAuthorReleases(ctx context.Context, slug string) (feeds.Feed, error) {
	loader := cache.CollectionLoaderFunc(func(ctx context.Context, key string) (collection model.Collection, err error) {
		now := time.Now()
		lastYear := now.AddDate(-1, 0, 0)
		slog.Info("Fetching releases", "author", slug)
		data, err := hardcover.RecentAuthorReleases(ctx, b.client, now, lastYear, []string{slug}, b.compilations)
		slog.Info("Retrieved author data", "author", slug, "elapsed", time.Since(now))
		if err != nil {
			return
		}
		if len(data.Authors) == 0 {
			err = fmt.Errorf("Author not found")
			return
		}
		author := data.Authors[0]
		authorName := author.Name
		var books []model.Book
		for _, contribution := range author.Contributions {
			books = append(books, b.mapBook(contribution.Book))
		}
		url := fmt.Sprintf("authors/%s", slug)
		return model.NewCollection(authorName, url, books), nil
	})
	collection, err := cache.CollectionCache.Get(ctx, fmt.Sprintf("hardcover/author/%s", slug), loader)
	if err != nil {
		return feeds.Feed{}, err
	}
	title := fmt.Sprintf("Hardcover Author Releases: %s", collection.Name)
	return b.buildFeed(title, b.buildUrl(collection.Slug), "", collection.Created, collection.Books)
}

func (b *hardcoverBuilder) GetSeriesReleases(ctx context.Context, slug string) (feeds.Feed, error) {
	loader := cache.CollectionLoaderFunc(func(ctx context.Context, key string) (collection model.Collection, err error) {
		now := time.Now()
		lastYear := now.AddDate(-1, 0, 0)
		slog.Info("Fetching releases", "series", slug)
		data, err := hardcover.RecentSeriesReleases(ctx, b.client, now, lastYear, []string{slug}, b.compilations)
		slog.Info("Retrieved series data", "series", slug, "elapsed", time.Since(now))
		if err != nil {
			return
		}
		if len(data.Series) == 0 {
			err = fmt.Errorf("Series not found")
			return
		}
		var books []model.Book
		seriesName := data.Series[0].Name
		for _, series := range data.BookSeries {
			books = append(books, b.mapBook(series.Book))
		}

		url := fmt.Sprintf("series/%s", slug)
		return model.NewCollection(seriesName, url, books), nil
	})
	collection, err := cache.CollectionCache.Get(ctx, fmt.Sprintf("hardcover/series/%s", slug), loader)
	if err != nil {
		return feeds.Feed{}, err
	}
	title := fmt.Sprintf("Hardcover Series Releases: %s", collection.Name)
	return b.buildFeed(title, b.buildUrl(collection.Slug), "", collection.Created, collection.Books)
}

func (b *hardcoverBuilder) getUserInterests(ctx context.Context, username string) (model.UserInterests, error) {
	loader := cache.UserLoaderFunc(func(ctx context.Context, key string) (interests model.UserInterests, err error) {
		now := time.Now()
		earliest := now.AddDate(-2, 0, 0)
		slog.Info("Fetching user interests", "user", username)
		data, err := hardcover.UserInterests(ctx, b.client, username, earliest)
		slog.Info("Retrieved user interests", "user", username, "elapsed", time.Since(now))
		if err != nil {
			return
		}
		if len(data.Users) == 0 {
			err = fmt.Errorf("User not found")
			return
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

		return model.UserInterests{
			Series:  series,
			Authors: authors,
		}, nil
	})
	return cache.UserCache.Get(ctx, fmt.Sprintf("hardcover/user/%s", username), loader)
}

func (b hardcoverBuilder) uncachedKeys(keys []string) (result []string) {
	for _, key := range keys {
		if _, ok := cache.CollectionCache.GetIfPresent(key); !ok {
			result = append(result, key)
		}
	}
	return result
}

func (b hardcoverBuilder) extractSlugs(keys []string) map[string]string {
	result := make(map[string]string)
	for _, key := range keys {
		slug := key[strings.LastIndex(key, "/")+1:]
		result[slug] = key
	}
	return result
}

func (b *hardcoverBuilder) GetUserReleases(ctx context.Context, username, filter string) (feeds.Feed, error) {
	interests, err := b.getUserInterests(ctx, username)
	if err != nil {
		return feeds.Feed{}, err
	}

	var seriesKeys []string
	var authorKeys []string
	var descBuilder strings.Builder
	descBuilder.WriteString("Includes New Releases from:\n")
	if slices.Contains([]string{"", "series"}, filter) && len(interests.Series) > 0 {
		descBuilder.WriteString(fmt.Sprintf("Series: %s\n", strings.Join(interests.Series, ", ")))
		for _, item := range interests.Series {
			seriesKeys = append(seriesKeys, fmt.Sprintf("hardcover/series/%s", item))
		}
	}

	if slices.Contains([]string{"", "author"}, filter) && len(interests.Authors) > 0 {
		descBuilder.WriteString(fmt.Sprintf("Authors: %s\n", strings.Join(interests.Authors, ", ")))
		for _, item := range interests.Authors {
			authorKeys = append(seriesKeys, fmt.Sprintf("hardcover/author/%s", item))
		}
	}

	seriesLoader := cache.BulkCollectionLoaderFunc(func(ctx context.Context, keys []string) (map[string]model.Collection, error) {
		result := make(map[string]model.Collection)
		now := time.Now()
		earliest := now.AddDate(-1, 0, 0)
		uncachedKeys := b.uncachedKeys(seriesKeys)
		slugMapping := b.extractSlugs(uncachedKeys)
		slugs := slices.Collect(maps.Keys(slugMapping))
		slog.Info("Fetching releases", "series", slugs, "keys", uncachedKeys)
		data, err := hardcover.RecentSeriesReleases(ctx, b.client, now, earliest, slugs, b.compilations)
		slog.Info("Retrieved series data", "series", slugs, "elapsed", time.Since(now))
		for _, series := range data.Series {
			if cacheKey, ok := slugMapping[series.Slug]; ok {
				var books []model.Book
				for _, book := range data.BookSeries {
					if book.Series.Slug == series.Slug {
						books = append(books, b.mapBook(book.Book))
					}
				}
				result[cacheKey] = model.NewCollection(series.Name, series.Slug, books)
			}
		}
		return result, err
	})
	authorLoader := cache.BulkCollectionLoaderFunc(func(ctx context.Context, keys []string) (map[string]model.Collection, error) {
		result := make(map[string]model.Collection)
		now := time.Now()
		earliest := now.AddDate(-1, 0, 0)
		uncachedKeys := b.uncachedKeys(authorKeys)
		slugMapping := b.extractSlugs(uncachedKeys)
		slugs := slices.Collect(maps.Keys(slugMapping))
		slog.Info("Fetching releases", "author", slugs, "keys", uncachedKeys)
		data, err := hardcover.RecentAuthorReleases(ctx, b.client, now, earliest, slugs, b.compilations)
		slog.Info("Retrieved author data", "author", slugs, "elapsed", time.Since(now))
		for _, author := range data.Authors {
			if cacheKey, ok := slugMapping[author.Slug]; ok {
				var books []model.Book
				for _, contribution := range author.Contributions {
					books = append(books, b.mapBook(contribution.Book))
				}
				result[cacheKey] = model.NewCollection(author.Name, author.Slug, books)
			}
		}
		return result, err
	})
	seriesCollections, err := cache.CollectionCache.BulkGet(ctx, seriesKeys, seriesLoader)
	if err != nil {
		slog.Error("Unable to fetch series data for", "user", username, "error", err)
	}
	authorCollections, err := cache.CollectionCache.BulkGet(ctx, authorKeys, authorLoader)
	if err != nil {
		slog.Error("Unable to fetch author data for", "user", username, "error", err)
	}

	bookMapping := make(map[int]model.Book)
	for _, series := range seriesCollections {
		for _, book := range series.Books {
			if _, ok := bookMapping[book.Id]; !ok {
				bookMapping[book.Id] = book
			}
		}
	}
	for _, author := range authorCollections {
		for _, book := range author.Books {
			if _, ok := bookMapping[book.Id]; !ok {
				bookMapping[book.Id] = book
			}
		}
	}
	books := slices.Collect(maps.Values(bookMapping))

	slug := fmt.Sprintf("@%s", username)
	collection := model.NewCollection(username, slug, books)

	title := fmt.Sprintf("Hardcover User Releases: %s", username)
	return b.buildFeed(title, b.buildUrl(slug), descBuilder.String(), collection.Created, collection.Books)
}

func NewHardcoverBuilder() Builder {
	token := os.Getenv("HARDCOVER_TOKEN")
	client := hardcover.GetClient(token)
	return &hardcoverBuilder{
		client:       client,
		compilations: false,
		builder: builder{
			templates: template.Must(
				template.New("base").Funcs(sprig.FuncMap()).ParseFS(fs, "templates/*.tmpl"),
			),
		},
	}
}
