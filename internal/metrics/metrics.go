// Package metrics provides Prometheus metrics (go_example_up, http_requests_total, http_request_duration_seconds) and Fiber middleware.
package metrics

import (
	"bytes"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

var (
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
)

// RegisterGoExampleUp registers go_example_up gauge. Call once per service at startup.
func RegisterGoExampleUp(serviceName string) {
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "go_example_up",
			Help:        "Service is up.",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		func() float64 { return 1 },
	))
}

// RegisterHTTPMetrics registers HTTP request metrics. Call once per service before HTTPMiddleware.
func RegisterHTTPMetrics(serviceName string) {
	RegisterGoExampleUp(serviceName)
	if httpRequestsTotal != nil {
		return
	}
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

// HTTPMiddleware records request count and duration. Requires RegisterHTTPMetrics first.
func HTTPMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		method := c.Method()
		path := c.Path()
		if path == "" {
			path = "/"
		}
		err := c.Next()
		status := strconv.Itoa(c.Response().StatusCode())
		dur := time.Since(start).Seconds()
		if httpRequestsTotal != nil {
			httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		}
		if httpRequestDuration != nil {
			httpRequestDuration.WithLabelValues(method, path).Observe(dur)
		}
		return err
	}
}

// MetricsHandler serves Prometheus metrics in text format.
func MetricsHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		mfs, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		var buf bytes.Buffer
		format := expfmt.NewFormat(expfmt.TypeTextPlain)
		encoder := expfmt.NewEncoder(&buf, format)
		for _, mf := range mfs {
			if err := encoder.Encode(mf); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
			}
		}
		c.Set("Content-Type", string(format))
		return c.SendString(buf.String())
	}
}
