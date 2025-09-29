package logger

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	Level  string `env:"LOG_LEVEL"  envDefault:"info"`
	Format string `env:"LOG_FORMAT"`
}

const (
	LogFormatJson = "json"
	LogFormatText = "text"
)

func init() {
	l, err := NewLoggerFromEnv()
	if err != nil {
		LogSink = zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), os.Stderr, zap.NewAtomicLevel()))
		LogSink.Error("failed to init the logger, using default", zap.Any("error", err))
	} else {
		LogSink = l
	}
	Logger = zapr.NewLogger(LogSink)
}

func Init(cfg LogConfig) {
	l, err := NewZapLog(&cfg)
	if err != nil {
		LogSink.Error("failed to init the logger, keep default", zap.Any("error", err))
		return
	} else {
		LogSink = l
	}
	Logger = zapr.NewLogger(LogSink)
}

var Logger logr.Logger
var LogSink *zap.Logger

func NewLoggerFromEnv() (*zap.Logger, error) {
	cfg, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}

	return NewZapLog(cfg)
}

func ConfigFromEnv() (*LogConfig, error) {
	var cfg LogConfig
	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log config: %w", err)
	}
	return &cfg, nil
}

func NewZapLog(cfg *LogConfig) (*zap.Logger, error) { //nolint: cyclop
	if cfg.Format == "" {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			cfg.Format = LogFormatText
		} else {
			cfg.Format = LogFormatJson
		}
	}
	var level zapcore.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		return nil, fmt.Errorf("invalid log level %q", cfg.Level)
	}

	var zapLog *zap.Logger
	var err error

	switch cfg.Format {
	case LogFormatJson:
		logCfg := zap.NewProductionConfig()
		logCfg.Level = zap.NewAtomicLevelAt(level)
		logCfg.Sampling = nil
		zapLog, err = logCfg.Build()
	case LogFormatText:
		fallthrough
	default:
		lgcfg := zap.NewDevelopmentConfig()
		lgcfg.Level = zap.NewAtomicLevelAt(level)
		zapLog, err = lgcfg.Build()
	}
	if err != nil {
		return nil, err
	}
	return zapLog, nil
}

func FromRequest(r *http.Request) logr.Logger {
	return FromContext(r.Context())
}

func FromContext(ctx context.Context) logr.Logger {
	l, err := logr.FromContext(ctx)
	if err != nil {
		return Logger
	}
	return l
}

func WithContext(ctx context.Context, logger logr.Logger) context.Context {
	return logr.NewContext(ctx, logger)
}

func LogSinkFromContext(ctx context.Context) *zap.Logger {
	//nolint:forcetypeassert
	return FromContext(ctx).GetSink().(zapr.Underlier).GetUnderlying()
}
