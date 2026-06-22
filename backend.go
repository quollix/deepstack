package deepstack

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

type LoggingBackend interface {
	ShouldLogBeSkipped(level slog.Level) bool
	LogRecord(logRecord *Record)
	LogWarning(message string, kv ...any)
}

type LoggingBackendImpl struct {
	slog *slog.Logger
}

func (s *LoggingBackendImpl) ShouldLogBeSkipped(level slog.Level) bool {
	return !s.slog.Handler().Enabled(context.Background(), level)
}

func (s *LoggingBackendImpl) LogRecord(logRecord *Record) {
	s.logRecord(logRecord, 5)
}

func (s *LoggingBackendImpl) logRecord(logRecord *Record, skipFunctionTreeLevels int) {
	var pcs [1]uintptr
	runtime.Callers(skipFunctionTreeLevels, pcs[:])
	slogRecord := slog.NewRecord(time.Now(), logRecord.level, logRecord.msg, pcs[0])

	for _, attr := range logRecord.attributes {
		slogRecord.AddAttrs(attr)
	}

	_ = s.slog.Handler().Handle(context.Background(), slogRecord)
}

func (s *LoggingBackendImpl) LogWarning(message string, kv ...any) {
	if len(kv) == 0 {
		record := &Record{
			level:      slog.LevelWarn,
			msg:        message,
			attributes: make([]slog.Attr, 0),
		}
		s.logRecord(record, 6)
	} else if len(kv) == 2 {
		key, ok := kv[0].(string)
		if !ok {
			s.slog.Warn(invalidKeyTypeMessage, slog.Any("key", kv[0]))
			return
		}
		s.slog.Warn(message, slog.Any(key, kv[1]))
	}
}
