package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"pvz/internal"
	"time"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := fmt.Sprintf("%d", c.Writer.Status())
		elapsed := time.Since(start).Seconds()

		internal.HttpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		internal.HttpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(elapsed)
	}
}
