package logger

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

type ILogger interface {
	Log(msg any, args ...any)
	Error(msg any, args ...any)
	Debug(msg any, args ...any)
	Info(msg any, args ...any)
	Warn(msg any, args ...any)
}

type LogMessage struct {
	CorrelationID string `json:"correlationId"`
	Message       any    `json:"message"`
}

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.NewWithOptions(os.Stderr, log.Options{
			ReportCaller:    true,
			ReportTimestamp: true,
			TimeFormat:      time.DateTime,
		}),
	}
}

func (l *Logger) Log(msg any, args ...any) {
	l.Logger.Info(msg, args...)
}

func (l *Logger) Error(msg any, args ...any) {
	l.Logger.Error(msg, args...)
}

func (l *Logger) Debug(msg any, args ...any) {
	l.Logger.Debug(msg, args...)
}

func (l *Logger) Info(msg any, args ...any) {
	l.Logger.Info(msg, args...)
}

func (l *Logger) Warn(msg any, args ...any) {
	l.Logger.Warn(msg, args...)
}
