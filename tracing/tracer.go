package tracing

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/lightstep/tracecontext.go/traceparent"
	"github.com/propertechnologies/monitor/context_util"
	log "github.com/propertechnologies/monitor/logging"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// EnvProjectID is the GCP project ID the trace exporter ships to.
	EnvProjectID = "GCP_PROJECT_ID"
	// EnvSampleRatio is the trace sample ratio in [0,1]. Unset or invalid means AlwaysSample.
	EnvSampleRatio = "MONITOR_TRACE_SAMPLE_RATIO"
)

type (
	Tracer struct {
		t trace.Tracer
		p *sdktrace.TracerProvider
	}

	HttpContext interface {
		GetHeader(string) string
	}
)

var (
	providerOnce sync.Once
	providerVal  *sdktrace.TracerProvider
	providerErr  error
)

// GetTracer returns a Tracer backed by a process-wide singleton TracerProvider.
// The first call initializes the provider; subsequent calls reuse it and only
// rebind the named instrumentation scope. Returns nil on init failure (logged).
//
// The provider is built from environment configuration:
//   - GCP_PROJECT_ID (required) — destination project for the Cloud Trace exporter
//   - MONITOR_TRACE_SAMPLE_RATIO (optional) — parent-based ratio sampler; defaults to AlwaysSample
//
// Callers should defer Shutdown on the returned Tracer in main.
func GetTracer(ctx context.Context, name string) *Tracer {
	tp, err := initProvider(ctx, name)
	if err != nil {
		log.Errorf(ctx, "tracing: failed to init provider: %v", err)
		return nil
	}

	return &Tracer{
		t: tp.Tracer("propertechnologies/" + name),
		p: tp,
	}
}

func initProvider(ctx context.Context, name string) (*sdktrace.TracerProvider, error) {
	providerOnce.Do(func() {
		projectID := os.Getenv(EnvProjectID)
		if projectID == "" {
			providerErr = fmt.Errorf("tracing: %s is not set", EnvProjectID)
			return
		}

		exporter, err := texporter.New(texporter.WithProjectID(projectID))
		if err != nil {
			providerErr = fmt.Errorf("tracing: create exporter: %w", err)
			return
		}

		res, err := resource.New(
			ctx,
			resource.WithDetectors(gcp.NewDetector()),
			resource.WithTelemetrySDK(),
			// Merge OTEL_RESOURCE_ATTRIBUTES / OTEL_SERVICE_NAME from env so
			// callers can stamp deployment.environment, service.namespace, etc.
			// without code changes per consumer.
			resource.WithFromEnv(),
			resource.WithAttributes(
				semconv.ServiceNameKey.String(name),
			),
		)
		if err != nil {
			providerErr = fmt.Errorf("tracing: build resource: %w", err)
			return
		}

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(samplerFromEnv()),
		)

		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))

		providerVal = tp
	})

	return providerVal, providerErr
}

func samplerFromEnv() sdktrace.Sampler {
	raw := os.Getenv(EnvSampleRatio)
	if raw == "" {
		return sdktrace.AlwaysSample()
	}
	ratio, err := strconv.ParseFloat(raw, 64)
	if err != nil || ratio < 0 || ratio > 1 {
		return sdktrace.AlwaysSample()
	}
	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))
}

// Shutdown flushes pending spans and stops the underlying provider. Safe to call
// once at process exit. Subsequent calls are no-ops.
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t == nil || t.p == nil {
		return nil
	}
	return t.p.Shutdown(ctx)
}

func AddRemoteSpanContext(ctx context.Context, traceID, spanID string) context.Context {
	tid, err := trace.TraceIDFromHex(traceID)
	if err != nil {
		log.Infof(ctx, "failed to parse traceID: %s", traceID)
		return ctx
	}

	sid, err := trace.SpanIDFromHex(spanID)
	if err != nil {
		log.Infof(ctx, "failed to parse spanID: %s", spanID)
		return ctx
	}

	traceState, _ := trace.TraceState{}.Insert("client_command", "run-app")

	ctx = trace.ContextWithRemoteSpanContext(
		ctx,
		trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    tid,
			SpanID:     sid,
			Remote:     true,
			TraceState: traceState,
		}),
	)

	return ctx
}

// Some headers "traceparent" or "proper-referer"
func GetTraceparent(c HttpContext, header string) (traceparent.TraceParent, error) {
	traceParent, err := traceparent.ParseString(c.GetHeader(header))
	if err != nil {
		return traceparent.TraceParent{}, err
	}

	return traceParent, nil
}

func (t *Tracer) Trace(ctx context.Context, name string, f func(context.Context)) {
	ctx, span := t.t.Start(ctx, name)
	defer span.End()

	f(ctx)
}

func (t *Tracer) TraceSpanLazyNaming(ctx context.Context, name func() string, f func(context.Context) error) error {
	ctx, span := t.t.Start(ctx, context_util.GetServiceName(ctx))
	defer span.End()

	err := f(ctx)

	if overwrite := name(); overwrite != "" {
		span.SetName(overwrite)
	}

	return err
}

func (t *Tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.t.Start(ctx, name, opts...)
}

func BuildTraceParent(ctx context.Context) traceparent.TraceParent {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		log.Infof(ctx, "failed to get span context")

		return traceparent.TraceParent{}
	}

	tp, err := traceparent.ParseString("00-" + spanCtx.TraceID().String() + "-" + spanCtx.SpanID().String() + "-00")
	if err != nil {
		log.Errorf(ctx, "failed to parse traceparent: %w", err)
	}

	return tp
}
