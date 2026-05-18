# monitor

Shared observability layer for Proper Go services. Wraps OpenTelemetry with GCP
exporters and Proper conventions so adopting services get logs, traces, and
metrics with minimal boilerplate.

## Layout

| Package | Purpose |
| --- | --- |
| `logging` | Structured `slog` → GCP Cloud Logging, automatic trace correlation |
| `tracing` | OTel TracerProvider → GCP Cloud Trace |
| `metrics` | OTel MeterProvider → GCP Cloud Monitoring |
| `middleware/gin` | Per-route HTTP metrics + OTel trace instrumentation for Gin |
| `client` | HTTP client wrapper that propagates trace context and flow IDs |
| `properrors` | Typed error codes with subtype chaining |
| `context_util` | Service name / request ID / flow ID context plumbing |

## Environment

| Var | Required | Default | Notes |
| --- | --- | --- | --- |
| `GCP_PROJECT_ID` | yes | — | Destination project for trace + metrics + logging trace correlation. Init returns an error if unset. |
| `MONITOR_TRACE_SAMPLE_RATIO` | no | AlwaysSample | Float in `[0,1]`. Parent-based ratio sampler. |
| `MONITOR_METRICS_INTERVAL` | no | `60s` | Periodic reader interval (Go duration). |
| `OTEL_RESOURCE_ATTRIBUTES` | no | — | Standard OTel env var; merged into every metric + span as resource attributes. Use for `deployment.environment=prod`, `service.namespace=…`, etc. Comma-separated `key=value` pairs. |
| `OTEL_SERVICE_NAME` | no | — | Standard OTel env var; overrides the `service.name` passed to `GetTracer` / `metrics.Init`. |

## Bootstrap

Initialize logging first (no env required), then tracing and metrics. Defer
shutdowns in `main`.

```go
ctx := context.Background()
ctx = logging.SetLogger(ctx, logging.NewLogger())

tr := tracing.GetTracer(ctx, "my-service")
defer tr.Shutdown(ctx)

mp, err := metrics.Init(ctx, "my-service")
if err != nil {
    log.Fatalf("metrics init: %v", err)
}
defer mp.Shutdown(ctx)
```

`GetTracer` and `metrics.Init` are idempotent — repeated calls reuse the same
provider.

## Per-route HTTP metrics (Gin)

```go
import monitorgin "github.com/propertechnologies/monitor/middleware/gin"

router := gin.New()
router.Use(monitorgin.Middleware("my-service")...)
```

Emits:

- Counter `http.server.requests{http.route, http.method, http.status_class, service.name}`
- Histogram `http.server.duration` (ms) with the same attributes

`http.route` uses Gin's matched route pattern (`c.FullPath()`) — never the raw
URI — so attribute cardinality stays bounded. Unmatched routes (404s) are
bucketed into `unmatched`.

## Logging

```go
logging.Infof(ctx, "starting worker %s", name)
logging.Errorf(ctx, "failed: %v", err)
logging.Reportf(ctx, "panic during job: %v", err) // adds stack trace, routes to Cloud Error Reporting
```

When a span is active in `ctx`, log records automatically carry
`logging.googleapis.com/trace`, `spanId`, and `trace_sampled` so Cloud Logging
can link the entry to its trace.
