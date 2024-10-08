package tracing

import (
	"context"

	_ "cloud.google.com/go/pubsub"
	"github.com/propertechnologies/monitor/context_util"
)

func SetServiceName(ctx context.Context, name string) context.Context {
	return context_util.SetServiceName(ctx, name)
}

func GetServiceName(ctx context.Context) string {
	return context_util.GetServiceName(ctx)
}
