package tracing

import (
	"context"
	"os"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/lightstep/tracecontext.go/traceparent"
	"github.com/propertechnologies/monitor/context_util"
	log "github.com/propertechnologies/monitor/logging"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"
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

func GetTracer(ctx context.Context, name string) *Tracer {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "proper-base" // default value
	}

	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		log.Errorf(ctx, "failed to create exporter: %v", err)
		return nil
	}

	res, err := resource.New(
		ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(name),
		),
	)
	if err != nil {
		log.Errorf(ctx, "failed to create resource: %v", err)
		return nil
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	return &Tracer{
		t: otel.GetTracerProvider().Tracer("propertechnologies/" + name),
		p: tp,
	}
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

	f(ctx)

	span.End()

	t.p.ForceFlush(ctx)
}

func (t *Tracer) TraceSpanLazyNaming(ctx context.Context, name func() string, f func(context.Context) error) error {
	ctx, span := t.t.Start(ctx, context_util.GetServiceName(ctx))

	err := f(ctx)

	overwrite := name()
	if overwrite != "" {
		span.SetName(overwrite)
	}

	span.End()

	t.p.ForceFlush(ctx)

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
