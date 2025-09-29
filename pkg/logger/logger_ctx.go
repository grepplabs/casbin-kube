package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func VCf(ctx context.Context, level int, format string, args ...any) {
	l := FromContext(ctx).WithCallDepth(1).V(level)
	if !l.Enabled() {
		return
	}
	l.Info(fmt.Sprintf(format, args...))
}

func DebugCf(ctx context.Context, format string, args ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Debugf(format, args...)
}

func InfoCf(ctx context.Context, format string, args ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Infof(format, args...)
}

func ErrorCf(ctx context.Context, format string, args ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Errorf(format, args...)
}

func WarnCf(ctx context.Context, format string, args ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Warnf(format, args...)
}

func PanicCf(ctx context.Context, format string, args ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Panicf(format, args...)
}

func FatalCf(ctx context.Context, format string, args ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Fatalf(format, args...)
}

func VCw(ctx context.Context, level int, msg string, keysAndValues ...any) {
	l := FromContext(ctx).WithCallDepth(1).V(level)
	if !l.Enabled() {
		return
	}
	l.Info(msg, keysAndValues...)
}

func DebugCw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Debugw(msg, keysAndValues...)
}

func InfoCw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Infow(msg, keysAndValues...)
}

func WarnCw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Warnw(msg, keysAndValues...)
}

func ErrorCw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Errorw(msg, keysAndValues...)
}

func PanicCw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Panicw(msg, keysAndValues...)
}

func FatalCw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	logSink := LogSinkFromContext(ctx).WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Fatalw(msg, keysAndValues...)
}
