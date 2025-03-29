package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/johannessarpola/poor-cache-go/internal/logger"
	"github.com/johannessarpola/poor-cache-go/internal/middleware"
	"github.com/johannessarpola/poor-cache-go/internal/rest"
	"github.com/johannessarpola/poor-cache-go/internal/store"
	"github.com/johannessarpola/poor-cache-go/internal/udp"
)

var Version = ""

func BuildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return info.Main.Version
}

func main() {
	r := gin.New()
	store := store.New()

	logger.SetServiceName("poor-cache-go")
	if Version != "" {
		logger.SetVersion(Version)
	} else {
		logger.SetVersion(BuildVersion())
	}

	r.Use(middleware.RequestLogger())

	v1group := r.Group("/api/v1")
	v1Svc := rest.New(store)

	rest.SetupRouter(v1group, v1Svc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create a context that listens for SIGTERM signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	go func() {
		if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
			logger.Errorf("Failed to start HTTP server %e", err)
		}
	}()

	udpServer := udp.New("0.0.0.0", 8081, store)
	go func() {
		if err := udpServer.Start(); err != nil {
			logger.Errorf("Failed to start UDP server %e", err)
		}
	}()

	// Wait for the SIGTERM signal
	<-ctx.Done()

	logger.Info("Received SIGTERM, shutting down gracefully")

	// cleanup
	udpServer.Close()
	store.Close()

	os.Exit(0)

}
