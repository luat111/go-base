package middlewares

import (
	"context"
	"go-base/pkg/config"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(config config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		duration, _ := strconv.Atoi(config.GetOrDefault("TIME_OUT", "10"))
		convertToSecond := time.Duration(duration) * time.Second

		ctx, cancel := context.WithTimeout(c.Request.Context(), convertToSecond)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
