package context_util

import (
	"context"
)

type (
	serviceNameKey string
)

const (
	sk serviceNameKey = "ServiceName"
)

func GetRequestID(ctx context.Context) string {
	return stringFromCtx(ctx, "RequestId")
}

func GetFlowID(ctx context.Context) string {
	return stringFromCtx(ctx, "FlowID")
}

func GetRootTaskID(ctx context.Context) string {
	return stringFromCtx(ctx, "RootTaskID")
}

func GetDebug(ctx context.Context) string {
	return stringFromCtx(ctx, "debug")
}

func GetEnv(ctx context.Context) string {
	return stringFromCtx(ctx, "env")
}

func GetServiceName(ctx context.Context) string {
	return stringFromCtx(ctx, sk)
}

func GetBotName(ctx context.Context) string {
	return stringFromCtx(ctx, "botname")
}

func IsProd(ctx context.Context) bool {
	return GetEnv(ctx) == "prod"
}

func IsDebugOn(ctx context.Context) bool {
	return (GetDebug(ctx) == "true" || GetEnv(ctx) == "local")
}

func SetDebugOn(ctx context.Context) context.Context {
	return context.WithValue(ctx, "debug", "true")
}

func SetServiceName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, sk, name)
}

func stringFromCtx(ctx context.Context, key interface{}) string {
	var value string

	l := ctx.Value(key)

	if l != nil {
		f, ok := l.(string)
		if ok {
			value = f
		}
	}

	return value
}
