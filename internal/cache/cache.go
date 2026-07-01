package cache

import (
	"context"
	"errors"
	"time"

	"github.com/RobBrazier/bookfeed/config"
	"github.com/RobBrazier/bookfeed/internal/model"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/eko/gocache/lib/v4/store"
	goCacheStore "github.com/eko/gocache/store/go_cache/v4"
	redisStore "github.com/eko/gocache/store/redis/v4"
	gocache "github.com/patrickmn/go-cache"
	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	CollectionTTL = 12 * time.Hour
	UserTTL       = 24 * time.Hour
)

var (
	collectionCache *marshaler.Marshaler
	userCache       *marshaler.Marshaler
)

type (
	CollectionLoaderFunc     func(ctx context.Context, key string) (model.Collection, error)
	BulkCollectionLoaderFunc func(ctx context.Context, keys []string) (map[string]model.Collection, error)
	UserLoaderFunc           func(ctx context.Context, key string) (model.UserInterests, error)
)

func Init() {
	s := newStore()
	c := cache.New[any](s)

	collectionCache = marshaler.New(c)
	userCache = marshaler.New(c)
}

func newStore() store.StoreInterface {
	if redisURL := config.RedisURL(); redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatal().Err(err).Str("url", redisURL).Msg("Unable to parse REDIS_URL")
		}
		client := redis.NewClient(opt)
		log.Info().Msg("Using Redis cache store")
		return redisStore.NewRedis(client)
	}

	client := gocache.New(gocache.DefaultExpiration, 10*time.Minute)
	log.Info().Msg("Using in-memory cache store (no REDIS_URL configured)")
	return goCacheStore.NewGoCache(client)
}

func isNotFound(err error) bool {
	var notFound *store.NotFound
	return errors.As(err, &notFound)
}

func GetOrLoad[T any](
	ctx context.Context,
	m *marshaler.Marshaler,
	key string,
	ttl time.Duration,
	load func(ctx context.Context, key string) (T, error),
) (T, error) {
	var val T
	_, err := m.Get(ctx, key, &val)
	if err == nil {
		log.Debug().Str("key", key).Msg("Cache hit")
		return val, nil
	}
	if !isNotFound(err) {
		return val, err
	}
	log.Debug().Str("key", key).Msg("Cache miss, loading")
	val, err = load(ctx, key)
	if err != nil {
		return val, err
	}
	_ = m.Set(ctx, key, val, store.WithExpiration(ttl))
	return val, nil
}

func BulkGetOrLoad[T any](
	ctx context.Context,
	m *marshaler.Marshaler,
	keys []string,
	ttl time.Duration,
	load func(ctx context.Context, keys []string) (map[string]T, error),
) (map[string]T, error) {
	result := make(map[string]T)
	var missing []string

	for _, key := range keys {
		var val T
		_, err := m.Get(ctx, key, &val)
		if err == nil {
			log.Debug().Str("key", key).Msg("Cache hit")
			result[key] = val
		} else if isNotFound(err) {
			missing = append(missing, key)
		} else {
			return nil, err
		}
	}

	if len(missing) == 0 {
		return result, nil
	}

	log.Debug().Strs("keys", missing).Msg("Cache miss, loading")

	loaded, err := load(ctx, missing)
	if err != nil {
		return result, err
	}

	for key, val := range loaded {
		result[key] = val
		_ = m.Set(ctx, key, val, store.WithExpiration(ttl))
	}

	return result, nil
}

func GetIfPresent[T any](ctx context.Context, m *marshaler.Marshaler, key string) (T, bool) {
	var val T
	_, err := m.Get(ctx, key, &val)
	if err != nil {
		return val, false
	}
	return val, true
}

func GetCollection(
	ctx context.Context,
	key string,
	loader CollectionLoaderFunc,
) (model.Collection, error) {
	return GetOrLoad(ctx, collectionCache, key, CollectionTTL, loader)
}

func BulkGetCollection(
	ctx context.Context,
	keys []string,
	loader BulkCollectionLoaderFunc,
) (map[string]model.Collection, error) {
	return BulkGetOrLoad(ctx, collectionCache, keys, CollectionTTL, loader)
}

func GetUser(ctx context.Context, key string, loader UserLoaderFunc) (model.UserInterests, error) {
	return GetOrLoad(ctx, userCache, key, UserTTL, loader)
}

func UncachedKeys(ctx context.Context, keys []string) []string {
	var result []string
	for _, key := range keys {
		if _, ok := GetIfPresent[model.Collection](ctx, collectionCache, key); !ok {
			result = append(result, key)
		}
	}
	return result
}

func InvalidateCollection(ctx context.Context, key string) error {
	return collectionCache.Delete(ctx, key)
}
