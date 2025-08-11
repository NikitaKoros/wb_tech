package logger

import "go.uber.org/zap"

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

type ZapLogger struct {
	log *zap.Logger
}

func some() {
	
	zap := zap.Logger{}
	zap.DebugLevel
}