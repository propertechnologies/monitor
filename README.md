# monitor
This library allows logging, error reporting and 
tracing of requests by using google platform.
- Logging
```golang
import "github.com/propertechnologies/monitor/logging"

ctx := logging.SetLogger(context.Background(), logging.NewLogger())
log.Infof(ctx, "No tracer found in context, running bot anyway.")
```

- Tracing
```golang
import "github.com/propertechnologies/monitor/tracing"

ctx = tracing.AddRemoteSpanContext(ctx, "traceID", "spanID")
tr := tracing.GetTracer(ctx, "serviceName")
	if tr != nil {
		err := tr.TraceSpanLazyNaming(
			ctx,
			func() string { return "serviceName" },
			MyCodeToBeTraced,
		)
		if err != nil {
			log.Reportf(ctx, "%v", err)
		}

		return
	}

func MyCodeToBeTraced(ctx context.Context) error{}
```

- Reporting
```golang
import "github.com/propertechnologies/monitor/logging"

ctx := logging.SetLogger(context.Background(), logging.NewLogger())

err := fmt.Errorf("some error %s", "!!!!")
log.Reportf(ctx, "No tracer found in context",err)
```

- Client
```golang
import "github.com/propertechnologies/monitor/client"

DoRequest(ctx context.Context, method, url string, body io.Reader)
DoRequestWithContentType(ctx context.Context, method, url string, body io.Reader, contentType string)
SetAuthorizationheader(request *http.Request)
```