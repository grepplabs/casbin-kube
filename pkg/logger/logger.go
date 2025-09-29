package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func Vf(level int, format string, args ...any) {
	l := Logger.WithCallDepth(1).V(level)
	if !l.Enabled() {
		return
	}
	l.Info(fmt.Sprintf(format, args...))
}

func Printf(format string, args ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Infof(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Errorf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Warnf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Fatalf(format, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Infow(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Warnw(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Errorw(msg, keysAndValues...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Panicw(msg, keysAndValues...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	logSink := LogSink.WithOptions(zap.AddCallerSkip(1))
	logSink.Sugar().Fatalw(msg, keysAndValues...)
}
