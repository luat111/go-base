package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-base/pkg/logger"
	"go-base/pkg/tracing"
	"time"

	ormLogger "gorm.io/gorm/logger"
)

const SlowThreshold = 200 * time.Millisecond

type IDbLogger interface {
	LogMode(ormLogger.LogLevel) ormLogger.Interface
	Info(context.Context, string, ...any)
	Warn(context.Context, string, ...any)
	Error(context.Context, string, ...any)
	Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error)
}

type DBLogger struct {
	baseLogger logger.ILogger
}

type dbLog struct {
	CorrelationId string        `json:"correlationId"`
	Query         string        `json:"query"`
	Duration      time.Duration `json:"duration"`
	Args          []any         `json:"args,omitempty"`
}

func newDbLogger(logger logger.ILogger) *DBLogger {
	return &DBLogger{
		baseLogger: logger,
	}
}

func (l *DBLogger) LogMode(level ormLogger.LogLevel) ormLogger.Interface {
	return ormLogger.Default.LogMode(level)
}

func (l *DBLogger) Info(ctx context.Context, s string, args ...any) {
	ormLogger.Default.Info(ctx, s, args...)
}

func (l *DBLogger) Warn(ctx context.Context, s string, args ...any) {
	ormLogger.Default.Info(ctx, s, args...)
}

func (l *DBLogger) Error(ctx context.Context, s string, args ...any) {
	ormLogger.Default.Info(ctx, s, args...)
}

func (l *DBLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	trackingId := tracing.FromContext(ctx)
	elapsed := time.Since(begin)

	logMsg := dbLog{
		CorrelationId: trackingId,
	}

	switch {
	case err != nil && !errors.Is(err, ormLogger.ErrRecordNotFound):
		sql, rows := fc()

		duration := time.Since(begin)
		logMsg.Query = sql
		logMsg.Duration = duration

		l.baseLogger.Error(err, "rows", rows, logMsg)

	case elapsed > SlowThreshold:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", SlowThreshold)

		duration := time.Since(begin)
		logMsg.Query = sql
		logMsg.Duration = duration

		l.baseLogger.Warn(slowLog, "rows", rows, logMsg)

	default:
		sql, rows := fc()

		duration := time.Since(begin)
		logMsg.Query = sql
		logMsg.Duration = duration

		l.baseLogger.Info("DBLOG", "rows", rows, "Message", logMsg)
	}
}
