package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

func getLevel() zerolog.Level {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.WarnLevel
	case "panic":
		return zerolog.WarnLevel
	default:
		return zerolog.DebugLevel
	}
}

func getLogger(isLocal bool) *zerolog.Logger {
	var writer io.Writer
	writer = os.Stdout
	if isLocal {
		writer = zerolog.NewConsoleWriter()
	}
	context := zerolog.New(writer).With().Timestamp().Caller().Stack()
	if !isLocal {
		context = context.Str("service.name", "bookfeed")
	}
	logger := context.Logger().Level(getLevel())
	log.Logger = logger

	// Set up slog to use zerolog for compatibility with go-retryablehttp
	slog.SetDefault(slog.New(slogzerolog.Option{Logger: &logger}.NewZerologHandler()))
	return &logger
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	isLocal := os.Getenv("APP_ENV") == "local"
	logger := getLogger(isLocal)
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
