// Package gin provides the standard Proper HTTP server observability layer for
// Gin: per-route request counters, latency histograms, and OpenTelemetry trace
// propagation. Install once on the router:
//
//	router.Use(monitorgin.Middleware("my-service"))
//
// Routes are bucketed by Gin's matched path pattern (c.FullPath()) — never by
// the raw request URI — so attribute cardinality stays bounded.
package gin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/propertechnologies/monitor/metrics"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

const (
	meterName = "github.com/propertechnologies/monitor/middleware/gin"

	metricRequests = "http.server.requests"
	metricDuration = "http.server.duration"

	attrRoute       = "http.route"
	attrMethod      = "http.method"
	attrStatusClass = "http.status_class"
	attrService     = "service.name"
)

// durationBucketsMs overrides the OTel default explicit-bucket histogram
// boundaries on http.server.duration. The OTel default caps at 10000ms, which
// silently clips p95/p99 for batch / long-poll endpoints to a flat 10s. New
// range spans 5ms to 5min (the Cloud Run default request timeout) so slow
// endpoints become measurable and so we see latency creeping toward the
// platform timeout before requests get killed.
var durationBucketsMs = []float64{
	5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000, 30000, 60000, 120000, 300000,
}

// Middleware returns a Gin handler chain that emits per-route HTTP metrics and
// installs the standard otelgin tracing instrumentation. `serviceName` is added
// as an attribute on every metric — match it to the service.name resource set
// at tracing/metrics Init time.
func Middleware(serviceName string) []gin.HandlerFunc {
	tracingMW := otelgin.Middleware(serviceName)
	metricsMW := metricsMiddleware(serviceName)
	return []gin.HandlerFunc{tracingMW, metricsMW}
}

func metricsMiddleware(serviceName string) gin.HandlerFunc {
	meter := metrics.Meter(meterName)

	requests, err := meter.Int64Counter(
		metricRequests,
		otelmetric.WithDescription("Count of HTTP requests served, by route and status class"),
		otelmetric.WithUnit("{request}"),
	)
	if err != nil {
		requests = nil
	}

	duration, err := meter.Float64Histogram(
		metricDuration,
		otelmetric.WithDescription("HTTP request duration, by route and status class"),
		otelmetric.WithUnit("ms"),
		otelmetric.WithExplicitBucketBoundaries(durationBucketsMs...),
	)
	if err != nil {
		duration = nil
	}

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// FullPath returns the matched route pattern (e.g. "/users/:id") rather
		// than the raw URI. Empty for 404s — bucket those into "unmatched" to
		// keep cardinality bounded.
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		attrs := otelmetric.WithAttributes(
			attribute.String(attrRoute, route),
			attribute.String(attrMethod, c.Request.Method),
			attribute.String(attrStatusClass, statusClass(c.Writer.Status())),
			attribute.String(attrService, serviceName),
		)

		if requests != nil {
			requests.Add(c.Request.Context(), 1, attrs)
		}
		if duration != nil {
			elapsedMs := float64(time.Since(start)) / float64(time.Millisecond)
			duration.Record(c.Request.Context(), elapsedMs, attrs)
		}
	}
}

func statusClass(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	case code >= 300:
		return "3xx"
	case code >= 200:
		return "2xx"
	case code >= 100:
		return "1xx"
	default:
		// Negative / zero. Shouldn't happen in a normal request lifecycle but
		// guard against it so the attribute stays bounded.
		_ = strconv.Itoa(code)
		return "unknown"
	}
}
