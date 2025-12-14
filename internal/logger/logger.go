package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new zap logger configured for the given environment
func NewLogger(env string) (*zap.Logger, error) {
	var cfg zap.Config

	switch env {
	case "production":
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.StacktraceKey = "" // Disable stacktrace in production logs
		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}
	case "test":
		// Minimal logging for tests
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	default: // development, staging
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.Development = true
	}

	logger, err := cfg.Build(zap.AddCaller(), zap.AddCallerSkip(0), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// NewNopLogger returns a no-op logger for testing
func NewNopLogger() *zap.Logger {
	return zap.NewNop()
}

// NewStdoutLogger creates a simple stdout logger
func NewStdoutLogger(level zapcore.Level) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		level,
	)

	return zap.New(core, zap.AddCaller())
}
