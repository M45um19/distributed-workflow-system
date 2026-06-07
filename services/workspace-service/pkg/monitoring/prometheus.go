package monitoring

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var HTTPRequestCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "route", "status_code"},
)

func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		routePattern := c.FullPath()
		if routePattern == "" {
			routePattern = c.Request.URL.Path
		}

		statusCode := strconv.Itoa(c.Writer.Status())

		HTTPRequestCounter.WithLabelValues(
			c.Request.Method,
			routePattern,
			statusCode,
		).Inc()
	}
}
