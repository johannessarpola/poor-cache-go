package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/johannessarpola/poor-cache-go/internal/logger"
)

func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		ctx.Next()

		// End timer
		end := time.Now()
		latency := end.Sub(start)

		// Log request details
		logger.Infof("Request: %s %s%s, Status: %d, Latency: %v",
			ctx.Request.Method,
			ctx.Request.Host,
			ctx.Request.URL.Path,
			ctx.Writer.Status(),
			latency)
	}
}
