package feed

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/bookfeed/config"
	"github.com/RobBrazier/bookfeed/internal/cache"
	"github.com/RobBrazier/bookfeed/internal/hardcover"
	"github.com/RobBrazier/bookfeed/internal/model"
	"github.com/RobBrazier/bookfeed/internal/view/pages"
	"github.com/gorilla/feeds"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type hardcoverBuilder struct {
	builder
	client       graphql.Client
	compilations bool
}

func (b hardcoverBuilder) cdnUrl(image model.Image) string {
	url := url.QueryEscape(image.Url)
	return fmt.Sprintf(
		"https://img.hardcover.app/enlarge?url=%s&width=%d&height=%d&type=webp",
		url,
		image.Width,
		image.Height,
	)
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
		image.Url = b.cdnUrl(image)
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
	loader := cache.CollectionLoaderFunc(
		func(ctx context.Context, key string) (collection model.Collection, err error) {
			now := time.Now()
			lastMonth := now.AddDate(0, -1, 0)
			log.Info().Msg("Fetching recent releases")
			data, err := hardcover.RecentReleases(ctx, b.client, now, lastMonth)
			log.Info().Dur("elapsed", time.Since(now)).Msg("Retrieved recent releases data")
			if err != nil {
				return collection, err
			}
			books := b.mapBooks(data.Books)
			return model.NewCollection("Recent", "upcoming/recent", books), nil
		},
	)
	collection, err := cache.CollectionCache.Get(ctx, "hardcover/releases", loader)
	if err != nil {
		return feeds.Feed{}, err
	}
	return b.buildFeed(
		ctx,
		"Hardcover: Recent Releases",
		b.buildUrl(collection.Slug),
		"",
		collection.Created,
		collection.Books,
	)
}

func (b *hardcoverBuilder) authorLoader(ids ...int) cache.BulkCollectionLoaderFunc {
	return cache.BulkCollectionLoaderFunc(
		func(ctx context.Context, keys []string) (map[string]model.Collection, error) {
			result := make(map[string]model.Collection)
			now := time.Now()
			earliest := now.AddDate(-1, 0, 0)
			uncachedKeys := b.uncachedKeys(keys)
			slugMapping := b.extractSlugs(uncachedKeys)
			slugs := slices.Collect(maps.Keys(slugMapping))
			log := log.With().
				Strs("authors", slugs).
				Strs("uncached", uncachedKeys).
				Ints("ids", ids).
				Logger()
			log.Info().Msg("Fetching releases")
			var releases []hardcover.AuthorRelease
			if len(ids) > 0 {
				data, err := hardcover.RecentAuthorReleases(
					ctx,
					b.client,
					now,
					earliest,
					slugs,
					b.compilations,
				)
				if err != nil {
					return result, err
				}
				releases = data.Authors
			} else {
				data, err := hardcover.RecentAuthorReleasesById(
					ctx,
					b.client,
					now,
					earliest,
					ids,
					b.compilations,
				)
				if err != nil {
					return result, err
				}
				releases = data.Authors
			}
			log.Info().Dur("elapsed", time.Since(now)).Msg("Retrieved author data")
			for _, author := range releases {
				if cacheKey, ok := slugMapping[author.Slug]; ok {
					var books []model.Book
					for _, contribution := range author.Contributions {
						books = append(books, b.mapBook(contribution.Book))
					}
					result[cacheKey] = model.NewCollection(
						author.Name,
						fmt.Sprintf("authors/%s", author.Slug),
						books,
					)
				}
			}
			for _, uncachedKey := range uncachedKeys {
				if _, ok := result[uncachedKey]; !ok {
					// Prevent abuse from entry not found
					result[uncachedKey] = model.Collection{}
				}
			}
			return result, nil
		},
	)
}

func (b *hardcoverBuilder) GetAuthorReleases(
	ctx context.Context,
	slug string,
) (feed feeds.Feed, err error) {
	loader := b.authorLoader()
	key := fmt.Sprintf("hardcover/authors/%s", slug)
	collections, err := cache.CollectionCache.BulkGet(
		ctx,
		[]string{key},
		loader,
	)
	if err != nil {
		log.Error().Err(err).Msgf("error retrieving author via cache, key=%s", key)
		return feed, err
	}

	collection, ok := collections[key]
	if !ok {
		return feed, fmt.Errorf("error occurred fetching series %s", key)
	}
	if !collection.Found {
		return feed, fmt.Errorf("author not found")
	}
	title := fmt.Sprintf("Hardcover Author Releases: %s", collection.Name)
	return b.buildFeed(
		ctx,
		title,
		b.buildUrl(collection.Slug),
		"",
		collection.Created,
		collection.Books,
	)
}

func (b *hardcoverBuilder) seriesLoader(ids ...int) cache.BulkCollectionLoaderFunc {
	return cache.BulkCollectionLoaderFunc(
		func(ctx context.Context, keys []string) (map[string]model.Collection, error) {
			result := make(map[string]model.Collection)
			now := time.Now()
			earliest := now.AddDate(-1, 0, 0)
			uncachedKeys := keys
			if len(keys) > 1 {
				uncachedKeys = b.uncachedKeys(keys)
			}
			slugMapping := b.extractSlugs(uncachedKeys)
			slugs := slices.Collect(maps.Keys(slugMapping))
			log := log.With().
				Strs("series", slugs).
				Strs("uncached", uncachedKeys).
				Ints("ids", ids).
				Logger()
			log.Info().Msg("Fetching releases")
			var releases []hardcover.SeriesRelease
			if len(ids) > 0 {
				data, err := hardcover.RecentSeriesReleasesById(
					ctx,
					b.client,
					now,
					earliest,
					ids,
					b.compilations,
				)
				if err != nil {
					return result, err
				}
				releases = data.Series
			} else {
				data, err := hardcover.RecentSeriesReleases(
					ctx,
					b.client,
					now,
					earliest,
					slugs,
					b.compilations,
				)
				if err != nil {
					return result, err
				}
				releases = data.Series
			}
			log.Info().Dur("elapsed", time.Since(now)).Msg("Retrieved series data")
			for _, series := range releases {
				if cacheKey, ok := slugMapping[series.Slug]; ok {
					var books []model.Book
					for _, book := range series.BookSeries {
						books = append(books, b.mapBook(book.Book))
					}
					result[cacheKey] = model.NewCollection(
						series.Name,
						fmt.Sprintf("series/%s", series.Slug),
						books,
					)
				}
			}
			for _, uncachedKey := range uncachedKeys {
				if _, ok := result[uncachedKey]; !ok {
					// Prevent abuse from entry not found
					result[uncachedKey] = model.Collection{}
				}
			}
			return result, nil
		},
	)
}

func (b *hardcoverBuilder) GetSeriesReleases(
	ctx context.Context,
	slug string,
) (feed feeds.Feed, err error) {
	loader := b.seriesLoader()
	key := fmt.Sprintf("hardcover/series/%s", slug)
	collections, err := cache.CollectionCache.BulkGet(
		ctx,
		[]string{key},
		loader,
	)
	if err != nil {
		return feed, err
	}

	collection, ok := collections[key]
	if !ok {
		return feed, fmt.Errorf("error occurred fetching series %s", key)
	}
	if !collection.Found {
		return feed, fmt.Errorf("series not found")
	}
	title := fmt.Sprintf("Hardcover Series Releases: %s", collection.Name)
	return b.buildFeed(
		ctx,
		title,
		b.buildUrl(collection.Slug),
		"",
		collection.Created,
		collection.Books,
	)
}

func (b *hardcoverBuilder) getUserInterests(
	ctx context.Context,
	username string,
) (model.UserInterests, error) {
	log := log.With().Str("user", username).Logger()
	loader := cache.UserLoaderFunc(
		func(ctx context.Context, key string) (interests model.UserInterests, err error) {
			now := time.Now()
			earliest := now.AddDate(-2, 0, 0)
			log.Info().Msg("Fetching user interests")
			data, err := hardcover.UserInterests(ctx, b.client, username, earliest)
			log.Info().
				Dur("elapsed", time.Since(now)).
				Int("count", len(data.UserBooks)).
				Msg("Retrieved user interests")
			if err != nil {
				return interests, err
			}
			if len(data.Users) == 0 {
				return interests, nil
			}
			authorMapping := make(map[string]int)
			seriesMapping := make(map[string]int)
			authorCount := make(map[string]int)
			seriesCount := make(map[string]int)
			for _, book := range data.UserBooks {
				for _, contribution := range book.Book.Contributors {
					if slices.Contains(
						[]string{"author", ""},
						strings.ToLower(contribution.Contribution),
					) {
						slug := contribution.Author.Slug
						authorCount[slug]++
						authorMapping[slug] = contribution.Author.Id
					}
				}
				if slug := book.Book.FeaturedSeries.Series.Slug; slug != "" {
					seriesCount[slug]++
					seriesMapping[slug] = book.Book.FeaturedSeries.Series.Id
				}
			}
			var authors []model.Interest
			var series []model.Interest

			// only check feeds for authors that have > 1 book read
			for slug, count := range authorCount {
				if count > 1 {
					authors = append(authors, model.Interest{
						Slug: slug,
						Id:   authorMapping[slug],
					})
				}
			}
			for slug, count := range seriesCount {
				if count > 1 {
					series = append(series, model.Interest{
						Slug: slug,
						Id:   seriesMapping[slug],
					})
				}
			}

			return model.UserInterests{
				Series:  series,
				Authors: authors,
				Found:   true,
			}, nil
		},
	)
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

func (b hardcoverBuilder) collectKeys(
	collect bool,
	key string,
	items []model.Interest,
	builder *strings.Builder,
) ([]string, []int) {
	if !collect || len(items) == 0 {
		return []string{}, []int{}
	}
	caser := cases.Title(language.English)
	title := caser.String(key)

	var slugs []string
	var ids []int
	for _, item := range items {
		slugs = append(slugs, item.Slug)
		ids = append(ids, item.Id)
	}

	fmt.Fprintf(builder, "%s: %s\n", title, strings.Join(slugs, ", "))

	var keys []string
	for _, item := range slugs {
		keys = append(keys, fmt.Sprintf("hardcover/%s/%s", key, item))
	}
	return keys, ids
}

func (b *hardcoverBuilder) GetUserReleases(
	ctx context.Context,
	username, filter string,
) (feeds.Feed, error) {
	log := log.With().Str("user", username).Str("filter", filter).Logger()
	interests, err := b.getUserInterests(ctx, username)
	if err != nil {
		return feeds.Feed{}, err
	}
	if !interests.Found {
		return feeds.Feed{}, fmt.Errorf("user not found")
	}

	log.Info().Interface("interests", interests).Msg("Getting releases for interests")

	var descBuilder strings.Builder
	descBuilder.WriteString("Includes New Releases from:\n")

	seriesKeys, seriesIds := b.collectKeys(
		slices.Contains([]string{"", "series"}, filter),
		"series",
		interests.Series,
		&descBuilder,
	)

	authorKeys, authorIds := b.collectKeys(
		slices.Contains([]string{"", "author"}, filter),
		"authors",
		interests.Authors,
		&descBuilder,
	)

	type job struct {
		key    string
		keys   []string
		loader cache.BulkCollectionLoaderFunc
	}
	jobs := []job{}
	if len(seriesKeys) > 0 {
		jobs = append(jobs, job{
			key:    "series",
			keys:   seriesKeys,
			loader: b.seriesLoader(seriesIds...),
		})
	}
	if len(authorKeys) > 0 {
		jobs = append(jobs, job{
			key:    "author",
			keys:   authorKeys,
			loader: b.authorLoader(authorIds...),
		})
	}

	var wg sync.WaitGroup
	results := sync.Map{}
	wg.Add(len(jobs))
	for _, job := range jobs {
		go func() {
			defer wg.Done()
			result, err := cache.CollectionCache.BulkGet(ctx, job.keys, job.loader)
			if err != nil {
				log.Error().Err(err).Msgf("Unable to fetch %s data", job.key)
			}
			for key, value := range result {
				results.Store(key, value)
			}
		}()
	}
	wg.Wait()
	keys := []string{}
	keys = append(keys, seriesKeys...)
	keys = append(keys, authorKeys...)

	collections := make(map[string]model.Collection)
	for _, key := range keys {
		value, ok := results.Load(key)
		log := log.With().Str("key", key).Logger()
		if !ok {
			log.Warn().Msg("key not found in collected collections")
			continue
		}
		collection, ok := value.(model.Collection)
		if !ok {
			log.Warn().Msg("value not Collection type")
			continue
		}
		collections[key] = collection
	}

	bookMapping := make(map[int]model.Book)

	for _, collection := range collections {
		for _, book := range collection.Books {
			if _, ok := bookMapping[book.Id]; !ok {
				bookMapping[book.Id] = book
			}
		}
	}
	books := slices.Collect(maps.Values(bookMapping))

	slug := fmt.Sprintf("@%s", username)
	collection := model.NewCollection(username, slug, books)

	title := fmt.Sprintf("Hardcover User Releases: %s", username)
	return b.buildFeed(
		ctx,
		title,
		b.buildUrl(slug),
		descBuilder.String(),
		collection.Created,
		collection.Books,
	)
}

func NewHardcoverBuilder() Builder {
	token := config.HardcoverToken()
	client := hardcover.GetClient(token)
	return &hardcoverBuilder{
		client:       client,
		compilations: false,
		builder: builder{
			provider: pages.HardcoverProvider,
		},
	}
}
