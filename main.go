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

	v1group := r.Group("/api/v1")
	v1Svc := rest.New(store)

	rest.SetupRouter(v1group, v1Svc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(fmt.Sprintf(":%s", port))
}
