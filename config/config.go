package config

import (
	"slices"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

type config struct {
	Port int `envconfig:"PORT" default:"8080"`
	Log  struct {
		Level    string `envconfig:"LOG_LEVEL" default:"debug"`
		Format   string `envconfig:"LOG_FORMAT" default:"text"`
		Requests bool   `envconfig:"LOG_REQUESTS" default:"false"`
	}
	Tokens struct {
		Hardcover string `envconfig:"HARDCOVER_TOKEN"`
	}
	Cache struct {
		StoragePath string `envconfig:"CACHE_STORAGE_PATH" default:"."`
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
