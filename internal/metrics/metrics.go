package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total"},
		[]string{"method", "path", "status"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_request_duration_seconds"},
		[]string{"method", "path"},
	)

	ReceptionCreated = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "reception_created_total"},
	)

	ProductsAdded = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "products_added_total"},
	)

	PvzCreated = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "pvz_created_total"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration, PvzCreated, ProductsAdded, ReceptionCreated)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := strconv.Itoa(c.Writer.Status())
		httpRequests.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		httpDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(time.Since(start).Seconds())
	}
}
