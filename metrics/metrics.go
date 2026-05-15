// Package metrics provides a process-wide OpenTelemetry MeterProvider wired to
// GCP Cloud Monitoring. It mirrors the bootstrap pattern of monitor/tracing:
// the first Init initializes a singleton provider, subsequent calls reuse it,
// and shutdown is the caller's responsibility (typically deferred in main).
//
// Required environment:
//   - GCP_PROJECT_ID — destination project for the Cloud Monitoring exporter
//
// Optional environment:
//   - MONITOR_METRICS_INTERVAL — periodic reader interval, Go duration string
//     (e.g. "30s", "1m"). Default 60s.
package metrics

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	EnvProjectID       = "GCP_PROJECT_ID"
	EnvMetricsInterval = "MONITOR_METRICS_INTERVAL"

	defaultExportInterval = 60 * time.Second
)

var (
	once    sync.Once
	mp      *sdkmetric.MeterProvider
	initErr error
)

// Provider is the handle returned by Init. Callers should keep it for the
// process lifetime and defer Shutdown in main.
type Provider struct {
	mp *sdkmetric.MeterProvider
}

// Init builds (or returns the already-built) MeterProvider, exports metrics
// to GCP Cloud Monitoring, and registers it as the global OTel meter provider.
//
// `serviceName` is set as the service.name resource attribute so dashboards can
// filter by service.
func Init(ctx context.Context, serviceName string) (*Provider, error) {
	once.Do(func() {
		projectID := os.Getenv(EnvProjectID)
		if projectID == "" {
			initErr = fmt.Errorf("metrics: %s is not set", EnvProjectID)
			return
		}

		exporter, err := mexporter.New(mexporter.WithProjectID(projectID))
		if err != nil {
			initErr = fmt.Errorf("metrics: create exporter: %w", err)
			return
		}

		res, err := resource.New(
			ctx,
			resource.WithDetectors(gcp.NewDetector()),
			resource.WithTelemetrySDK(),
			resource.WithAttributes(
				semconv.ServiceNameKey.String(serviceName),
			),
		)
		if err != nil {
			initErr = fmt.Errorf("metrics: build resource: %w", err)
			return
		}

		reader := sdkmetric.NewPeriodicReader(
			exporter,
			sdkmetric.WithInterval(intervalFromEnv()),
		)

		mp = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(reader),
		)
		otel.SetMeterProvider(mp)
	})

	if initErr != nil {
		return nil, initErr
	}
	return &Provider{mp: mp}, nil
}

func intervalFromEnv() time.Duration {
	raw := os.Getenv(EnvMetricsInterval)
	if raw == "" {
		return defaultExportInterval
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return defaultExportInterval
	}
	return d
}

// Shutdown flushes pending metrics and stops the periodic reader.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil || p.mp == nil {
		return nil
	}
	return p.mp.Shutdown(ctx)
}

// ForceFlush exports any buffered metrics immediately. Useful before a known
// process exit when Shutdown can't be deferred.
func (p *Provider) ForceFlush(ctx context.Context) error {
	if p == nil || p.mp == nil {
		return nil
	}
	return p.mp.ForceFlush(ctx)
}

// Meter returns a named OpenTelemetry Meter. The returned Meter is bound to
// the global provider — calls before Init resolve to a no-op meter and emit
// nothing, which is the OTel default.
func Meter(name string) metric.Meter {
	return otel.Meter(name)
}
