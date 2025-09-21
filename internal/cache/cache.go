package cache

import (
	"errors"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/RobBrazier/bookfeed/config"
	"github.com/RobBrazier/bookfeed/internal/model"
	"github.com/maypok86/otter/v2"
)

var (
	CollectionCache *otter.Cache[string, model.Collection]
	UserCache       *otter.Cache[string, model.UserInterests]
)

type CollectionLoaderFunc = otter.LoaderFunc[string, model.Collection]
type BulkCollectionLoaderFunc = otter.BulkLoaderFunc[string, model.Collection]
type UserLoaderFunc = otter.LoaderFunc[string, model.UserInterests]

func init() {
	CollectionCache = newCollectionCache()
	UserCache = newUserCache()
}

func newCollectionCache() *otter.Cache[string, model.Collection] {
	return otter.Must(&otter.Options[string, model.Collection]{
		MaximumSize:      10_000,
		ExpiryCalculator: otter.ExpiryCreating[string, model.Collection](12 * time.Hour),
	})
}

func newUserCache() *otter.Cache[string, model.UserInterests] {
	return otter.Must(&otter.Options[string, model.UserInterests]{
		MaximumSize:      10_000,
		ExpiryCalculator: otter.ExpiryCreating[string, model.UserInterests](24 * time.Hour),
	})
}

func LoadCache() {
	cachePath := config.CacheStorage()
	collectionPath := path.Join(cachePath, "collection.gob")
	userPath := path.Join(cachePath, "user.gob")
	log.Info().Str("path", collectionPath).Msg("Loading collection cache")
	if err := otter.LoadCacheFromFile(CollectionCache, collectionPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Error().Err(err).Msg("Load cache failed")
		}
	}
	log.Info().Str("path", userPath).Msg("Loading user cache")
	if err := otter.LoadCacheFromFile(UserCache, userPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Error().Err(err).Msg("Load cache failed")
		}
	}
}

func SaveCache() {
	cachePath := config.CacheStorage()
	collectionPath := path.Join(cachePath, "collection.gob")
	userPath := path.Join(cachePath, "user.gob")
	log.Info().Str("path", collectionPath).Msg("Saving collection cache")
	if err := otter.SaveCacheToFile(CollectionCache, collectionPath); err != nil {
		log.Error().Err(err).Msg("Save cache failed")
	}
	log.Info().Str("path", userPath).Msg("Saving user cache")
	if err := otter.SaveCacheToFile(UserCache, userPath); err != nil {
		log.Error().Err(err).Msg("Save cache failed")
	}
}
