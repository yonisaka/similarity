package logging

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Info(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Debug(msg string, fields ...zapcore.Field)
	Named(name string) Logger
	With(...zapcore.Field) Logger
	Unwrap() *zap.Logger
}

type logger struct {
	logger *zap.Logger
}

func NewLogger() (Logger, error) {
	config := zap.NewProductionConfig()

	l, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build a Zap logger: %w", err)
	}

	return &logger{
		logger: l,
	}, nil
}

func (l *logger) Info(msg string, fields ...zapcore.Field) {
	l.logger.Info(msg, fields...)
}

func (l *logger) Error(msg string, fields ...zapcore.Field) {
	l.logger.Error(msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...zapcore.Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...zapcore.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *logger) Debug(msg string, fields ...zapcore.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *logger) Named(name string) Logger {
	return &logger{logger: l.logger.Named(name)}
}

func (l *logger) With(fields ...zapcore.Field) Logger {
	return &logger{logger: l.logger.With(fields...)}
}

func (l *logger) Unwrap() *zap.Logger {
	return l.logger
}
