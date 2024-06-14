package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/propertechnologies/monitor/context_util"
	"go.opentelemetry.io/otel/trace"
)

type (
	GCPLoggerWrapper struct {
		logger *slog.Logger
	}
)

func NewLogger() *GCPLoggerWrapper {
	// Use json as our base logging format.
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: replacer})
	// Add span context attributes when Context is passed to logging calls.
	instrumentedHandler := handlerWithSpanContext(jsonHandler)
	// Set this handler as the global slog handler.
	var l = slog.New(instrumentedHandler)
	slog.SetDefault(l)

	return &GCPLoggerWrapper{
		logger: l,
	}
}

func (g *GCPLoggerWrapper) Infof(ctx context.Context, format string, args ...interface{}) {
	slog.InfoContext(ctx, fmt.Sprintf(format, args...))
}

func (g *GCPLoggerWrapper) Errorf(ctx context.Context, format string, args ...interface{}) {
	slog.ErrorContext(ctx, fmt.Sprintf(format, args...))
}

func handlerWithSpanContext(handler slog.Handler) *spanContextLogHandler {
	return &spanContextLogHandler{Handler: handler}
}

// spanContextLogHandler is an slog.Handler which adds attributes from the
// span context.
type spanContextLogHandler struct {
	slog.Handler
}

// Handle overrides slog.Handler's Handle method. This adds attributes from the
// span context to the slog.Record.
func (t *spanContextLogHandler) Handle(ctx context.Context, record slog.Record) error {
	s := trace.SpanContextFromContext(ctx)
	if s.IsValid() {
		// Add trace context attributes following Cloud Logging structured log format described
		// in https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
		record.AddAttrs(
			slog.Any("logging.googleapis.com/trace", "projects/proper-base/traces/"+s.TraceID().String()),
		)
		record.AddAttrs(
			slog.Any("logging.googleapis.com/spanId", s.SpanID()),
		)
		record.AddAttrs(
			slog.Bool("logging.googleapis.com/trace_sampled", s.TraceFlags().IsSampled()),
		)
	}

	record.AddAttrs(
		slog.String("app", context_util.GetServiceName(ctx)),
	)
	record.AddAttrs(
		slog.String("rid", context_util.GetRequestID(ctx)),
	)
	record.AddAttrs(
		slog.String("flow-id", context_util.GetFlowID(ctx)),
	)
	record.AddAttrs(
		slog.String("root-task-id", context_util.GetRootTaskID(ctx)),
	)

	return t.Handler.Handle(ctx, record)
}

func replacer(groups []string, a slog.Attr) slog.Attr {
	// Rename attribute keys to match Cloud Logging structured log format
	switch a.Key {
	case slog.LevelKey:
		a.Key = "severity"
		// Map slog.Level string values to Cloud Logging LogSeverity
		// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
		if level := a.Value.Any().(slog.Level); level == slog.LevelWarn {
			a.Value = slog.StringValue("WARNING")
		}
	case slog.TimeKey:
		a.Key = "timestamp"
	case slog.MessageKey:
		a.Key = "message"
	}

	return a
}
