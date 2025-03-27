package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/johannessarpola/poor-cache-go/internal/logger"
	"github.com/johannessarpola/poor-cache-go/internal/middleware"
	"github.com/johannessarpola/poor-cache-go/internal/rest"
	"github.com/johannessarpola/poor-cache-go/internal/store"
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

	r.Run(fmt.Sprintf(":%s", port))
}
