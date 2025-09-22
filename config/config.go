package config

import (
	"slices"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

type config struct {
	Port int `default:"8080" envconfig:"PORT"`
	Log  struct {
		Level    string `default:"debug" envconfig:"LOG_LEVEL"`
		Format   string `default:"text"  envconfig:"LOG_FORMAT"`
		Requests bool   `default:"false" envconfig:"LOG_REQUESTS"`
	}
	Tokens struct {
		Hardcover string `envconfig:"HARDCOVER_TOKEN"`
	}
	Cache struct {
		StoragePath string `default:"." envconfig:"CACHE_STORAGE_PATH"`
	}
}

var cfg config

func LoadConfig() error {
	err := envconfig.Process("", &cfg)
	if err != nil {
		return err
	}
	return nil
}

func Config() config {
	return cfg
}

func Port() int {
	return cfg.Port
}

func LogLevel() zerolog.Level {
	switch strings.ToLower(cfg.Log.Level) {
	case "trace":
		return zerolog.TraceLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.DebugLevel
	}
}

func LogFormat() string {
	allowed := []string{"text", "json"}
	format := strings.ToLower(cfg.Log.Format)
	if slices.Contains(allowed, format) {
		return format
	}
	return "json"
}

func LogRequests() bool {
	return cfg.Log.Requests
}

func HardcoverToken() string {
	return cfg.Tokens.Hardcover
}

func CacheStorage() string {
	return cfg.Cache.StoragePath
}
