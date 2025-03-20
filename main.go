package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/johannessarpola/poor-cache-go/internal/rest"
	"github.com/johannessarpola/poor-cache-go/internal/store"
)

func main() {
	store := store.New()
	r := gin.Default()

	rg := r.Group("/api/v1")
	rest.SetupRouter(rg, store)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(fmt.Sprintf(":%s", port))
}
