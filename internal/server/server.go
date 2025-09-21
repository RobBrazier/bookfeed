package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/RobBrazier/bookfeed/config"
	"github.com/RobBrazier/bookfeed/internal/cache"
	"github.com/RobBrazier/bookfeed/internal/feed"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port    int
	logger  *zerolog.Logger
	builder feed.Builder
}

func getSlogLevel(level zerolog.Level) slog.Leveler {
	switch level {
	case zerolog.DebugLevel:
		return slog.LevelDebug
	case zerolog.InfoLevel:
		return slog.LevelInfo
	case zerolog.WarnLevel:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

func getLogger() *zerolog.Logger {
	var writer io.Writer
	writer = os.Stdout
	context := zerolog.New(writer).With().Timestamp().Caller().Stack()
	logger := context.Logger()
	err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to load config")
	}
	if config.LogFormat() == "text" {
		writer = zerolog.NewConsoleWriter()
		logger = logger.Output(writer)
	} else {
		logger = logger.With().Str("service.name", "bookfeed").Logger()
	}
	level := config.LogLevel()
	logger = logger.Level(level)
	log.Logger = logger
	slogLevel := getSlogLevel(level)

	// Set up slog to use zerolog for compatibility with go-retryablehttp
	slog.SetDefault(slog.New(slogzerolog.Option{Level: slogLevel, Logger: &logger}.NewZerologHandler()))
	return &logger
}

func NewServer() *http.Server {
	logger := getLogger()
	port := config.Port()
	log.Info().Int("port", port).Msg("Started server")
	scheduler, _ := gocron.NewScheduler()
	scheduler.NewJob(
		gocron.DurationJob(1*time.Hour),
		gocron.NewTask(cache.SaveCache),
	)
	scheduler.Start()

	NewServer := &Server{
		port:    port,
		logger:  logger,
		builder: feed.NewHardcoverBuilder(),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	server.RegisterOnShutdown(func() {
		scheduler.Shutdown()
	})

	cache.LoadCache()

	return server
}
