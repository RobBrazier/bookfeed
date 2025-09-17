package cache

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/RobBrazier/bookfeed/internal/model"
	"github.com/maypok86/otter/v2"
)

var (
	CollectionCache *otter.Cache[string, model.Collection]
	UserCache       *otter.Cache[string, model.UserInterests]
	cachePath       string
)

type CollectionLoaderFunc = otter.LoaderFunc[string, model.Collection]
type BulkCollectionLoaderFunc = otter.BulkLoaderFunc[string, model.Collection]
type UserLoaderFunc = otter.LoaderFunc[string, model.UserInterests]

func init() {
	CollectionCache = newCollectionCache()
	UserCache = newUserCache()

	cachePath = "."
	if value, ok := os.LookupEnv("CACHE_BACKUP"); ok {
		cachePath = value
	}
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
	collectionPath := path.Join(cachePath, "collection.gob")
	userPath := path.Join(cachePath, "user.gob")
	slog.Info(fmt.Sprintf("Loading collection cache from %s", collectionPath))
	if err := otter.LoadCacheFromFile(CollectionCache, collectionPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			slog.Warn("Load cache failed", "error", err)
		}
	}
	slog.Info(fmt.Sprintf("Loading user cache from %s", userPath))
	if err := otter.LoadCacheFromFile(UserCache, userPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			slog.Warn("Load cache failed", "error", err)
		}
	}
}

func SaveCache() {
	collectionPath := path.Join(cachePath, "collection.gob")
	userPath := path.Join(cachePath, "user.gob")
	slog.Info(fmt.Sprintf("Saving collection cache to %s", collectionPath))
	if err := otter.SaveCacheToFile(CollectionCache, collectionPath); err != nil {
		slog.Warn("Save cache failed", "error", err)
	}
	slog.Info(fmt.Sprintf("Saving user cache to %s", userPath))
	if err := otter.SaveCacheToFile(UserCache, userPath); err != nil {
		slog.Warn("Save cache failed", "error", err)
	}
}
