package deepstack

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRawConsoleHandlerQuotesSpacedStringValues(t *testing.T) {
	var buf bytes.Buffer
	handler := ConsoleHandler{
		w:    &buf,
		opts: &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug},
	}

	record := slog.NewRecord(time.Date(2026, time.May, 15, 12, 0, 0, 0, time.UTC), slog.LevelInfo, "hello", 0)
	record.AddAttrs(slog.String("plain", "value1"), slog.String("spaced", "value 2"), slog.Any("count", 3))

	err := handler.Handle(context.Background(), record)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), `plain=value1`)
	assert.Contains(t, buf.String(), `spaced="value 2"`)
	assert.Contains(t, buf.String(), `count=3`)
}

func TestRawConsoleHandlerRendersStackTraceInSameWrite(t *testing.T) {
	var buf bytes.Buffer
	handler := ConsoleHandler{
		w:    &buf,
		opts: &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug},
	}

	record := slog.NewRecord(time.Date(2026, time.May, 15, 12, 0, 0, 0, time.UTC), slog.LevelError, "boom", 0)
	record.AddAttrs(slog.String("stack_trace", "main.func\n\t/path/file.go:42\n"))

	err := handler.Handle(context.Background(), record)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "\x1b[31m")
	assert.Contains(t, buf.String(), "main.func")
	assert.Contains(t, buf.String(), "/path/file.go:42")
}

func TestConsoleHandlerEnabledRespectsConfiguredLevel(t *testing.T) {
	handler := ConsoleHandler{
		w:    &bytes.Buffer{},
		opts: &slog.HandlerOptions{Level: slog.LevelWarn},
	}

	assert.False(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelWarn))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelError))
}

func TestConsoleHandlerWithAttrsAddsStaticAttributes(t *testing.T) {
	var originalBuf bytes.Buffer
	var enrichedBuf bytes.Buffer
	originalHandler := ConsoleHandler{
		w:    &originalBuf,
		opts: &slog.HandlerOptions{Level: slog.LevelDebug},
	}

	enrichedHandler := originalHandler.WithAttrs([]slog.Attr{slog.String("handler_key", "handler_value")}).(ConsoleHandler)
	enrichedHandler.w = &enrichedBuf
	record := slog.NewRecord(time.Date(2026, time.May, 15, 12, 0, 0, 0, time.UTC), slog.LevelInfo, "hello", 0)

	err := originalHandler.Handle(context.Background(), record)
	assert.NoError(t, err)
	err = enrichedHandler.Handle(context.Background(), record)
	assert.NoError(t, err)

	assert.NotContains(t, originalBuf.String(), "handler_key=handler_value")
	assert.Contains(t, enrichedBuf.String(), "handler_key=handler_value")
}
