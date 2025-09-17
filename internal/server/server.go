package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RobBrazier/bookfeed/internal/cache"
	"github.com/RobBrazier/bookfeed/internal/feed"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port    int
	builder feed.Builder
}

func NewServer() (*http.Server, func()) {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	NewServer := &Server{
		port:    port,
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

	shutdown := func() {
		cache.SaveCache()
	}

	return server, shutdown
}
