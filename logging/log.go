package logging

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/propertechnologies/monitor/context_util"
)

type (
	Logger interface {
		Infof(ctx context.Context, format string, args ...interface{})
		Errorf(ctx context.Context, format string, args ...interface{})
		Warnf(ctx context.Context, format string, args ...interface{})
	}

	DefaultLogger struct{}

	report struct{}
)

func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{}
}

func (l *DefaultLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(message)
}

func (l *DefaultLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Infof(ctx, format, args...)
}

func (l *DefaultLogger) Reportf(ctx context.Context, format string, args ...interface{}) {
	l.Infof(ctx, format, args...)
}

func (l *DefaultLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.Infof(ctx, format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	log := GetLoggerOrDefault(ctx, &DefaultLogger{})
	log.Infof(ctx, format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	log := GetLoggerOrDefault(ctx, &DefaultLogger{})
	log.Errorf(ctx, format, args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	log := GetLoggerOrDefault(ctx, &DefaultLogger{})
	log.Warnf(ctx, format, args...)
}

func Reportf(ctx context.Context, format string, args ...interface{}) {
	log := GetLoggerOrDefault(ctx, &DefaultLogger{})
	ctx = context.WithValue(ctx, report{}, debug.Stack())

	log.Errorf(ctx, format, args...)
}

func GetLoggerOrDefault(ctx context.Context, actualLogger Logger) Logger {
	logger := loggerFromCtx(ctx)
	if logger != nil {
		actualLogger = logger
	}

	return actualLogger
}

func SetLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, "logger", logger)
}

func SetRootTaskID(ctx context.Context, taskID string) context.Context {
	return context_util.SetRootTaskID(ctx, taskID)
}

func loggerFromCtx(ctx context.Context) Logger {
	l := ctx.Value("logger")
	if l != nil {
		v, ok := l.(Logger)
		if ok {
			return v
		}
	}

	return nil
}
