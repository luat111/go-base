package logger

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

type ILogger interface {
	SetLoggerPrefix(prefix string)
	Log(msg any, args ...any)
	Error(msg any, args ...any)
	Debug(msg any, args ...any)
	Info(msg any, args ...any)
	Warn(msg any, args ...any)
	Debugf(string, ...any)
}

type LogMessage struct {
	CorrelationID string `json:"correlationId"`
	Message       any    `json:"message"`
}

type Logger struct {
	*log.Logger
}

func NewLogger(prefix string) *Logger {
	return &Logger{
		Logger: log.NewWithOptions(os.Stderr, log.Options{
			Prefix:          "[" + prefix + "]",
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

func (l *Logger) Debugf(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

func (l *Logger) SetLoggerPrefix(prefix string) {
	l.SetPrefix(prefix)
}
