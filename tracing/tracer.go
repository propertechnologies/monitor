package tracing

import (
	"context"
	"fmt"
	"os"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/lightstep/tracecontext.go/traceparent"
	"github.com/propertechnologies/monitor/context_util"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/appengine/log"
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

func GetTracer(ctx context.Context, name string) *Tracer {
	tr, err := buildTracer(ctx, name)
	if err != nil {
		log.Errorf(ctx, "failed to build tracer: %v", err)
		return nil
	}

	return tr
}

func (t *Tracer) Trace(ctx context.Context, name string, f func(context.Context)) {
	ctx, span := t.t.Start(ctx, name)

	f(ctx)

	span.SetName(context_util.GetServiceName(ctx))
	span.End()

	t.p.ForceFlush(ctx)
}

func (t *Tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.t.Start(ctx, name, opts...)
}

func GetTraceparent(c HttpContext) (traceparent.TraceParent, error) {
	// traceparent header info is sent here from bots given that traceparent header is overwriten by gcp
	traceParent, err := traceparent.ParseString(c.GetHeader("proper-referer"))
	if err == nil {
		return traceParent, nil
	}

	traceParent, err = traceparent.ParseString(c.GetHeader("traceparent"))
	if err != nil {
		return traceparent.TraceParent{}, err
	}

	return traceParent, nil
}

func buildTracer(ctx context.Context, name string) (*Tracer, error) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "proper-base" // default value
	}

	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
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
		return nil, fmt.Errorf("failed to create resource: %w", err)
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
	}, nil
}
