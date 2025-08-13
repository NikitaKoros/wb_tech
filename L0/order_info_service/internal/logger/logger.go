package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
}

type Field = zapcore.Field

type ZapLogger struct {
	logger *zap.Logger
}

func NewLogger() (Logger, error) {
	config := zap.NewProductionConfig()
	
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig = encoderConfig
	
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	
	return &ZapLogger{logger: logger}, nil
}

func (l *ZapLogger) log(level zapcore.Level, msg string, fields ...Field) {
	processedFields := make([]Field, 0, len(fields))
	for _, field := range fields {
		if isSensitiveField(field) {
			field.String = "*****"
			field.Interface = nil
		}
		processedFields = append(processedFields, field)
	}
	
	switch level {
		case zapcore.InfoLevel:
			l.logger.Info(msg, processedFields...)
		case zapcore.ErrorLevel:
			l.logger.Error(msg, processedFields...)
		case zapcore.DebugLevel:
			l.logger.Debug(msg, processedFields...)
		case zapcore.WarnLevel:
			l.logger.Warn(msg, processedFields...)
	}
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.log(zapcore.InfoLevel, msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.log(zapcore.InfoLevel, msg, fields...)
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.log(zapcore.InfoLevel, msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.log(zapcore.InfoLevel, msg, fields...)
}

func isSensitiveField(field Field) bool {
	sensitiveFileds := map[string]struct{}{
		"address":{},
		"region":{},
		"email":{},
		"phone":{},
	}
	
	_, exists := sensitiveFileds[field.String]
	return exists
}