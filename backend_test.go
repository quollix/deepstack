package deepstack

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingBackendShouldLogBeSkippedRespectsHandlerLevel(t *testing.T) {
	var buf bytes.Buffer
	backend := newLoggingBackendForTest(&buf, slog.LevelWarn)

	assert.True(t, backend.ShouldLogBeSkipped(slog.LevelInfo))
	assert.False(t, backend.ShouldLogBeSkipped(slog.LevelWarn))
	assert.False(t, backend.ShouldLogBeSkipped(slog.LevelError))
}

func TestLoggingBackendLogRecordForwardsMessageAndAttrs(t *testing.T) {
	var buf bytes.Buffer
	backend := newLoggingBackendForTest(&buf, slog.LevelDebug)
	record := &Record{
		level: slog.LevelInfo,
		msg:   "hello",
		attributes: []slog.Attr{
			slog.String("key1", "value1"),
			slog.Int("count", 3),
		},
	}

	backend.LogRecord(record)

	output := decodeLastLogRecord(t, &buf)
	assert.Equal(t, "INFO", output["level"])
	assert.Equal(t, "hello", output["msg"])
	assert.Equal(t, "value1", output["key1"])
	assert.Equal(t, float64(3), output["count"])
}

func TestLoggingBackendLogWarningWithoutContext(t *testing.T) {
	var buf bytes.Buffer
	backend := newLoggingBackendForTest(&buf, slog.LevelDebug)

	backend.LogWarning("plain warning")

	output := decodeLastLogRecord(t, &buf)
	assert.Equal(t, "WARN", output["level"])
	assert.Equal(t, "plain warning", output["msg"])
}

func TestLoggingBackendLogWarningWithContext(t *testing.T) {
	var buf bytes.Buffer
	backend := newLoggingBackendForTest(&buf, slog.LevelDebug)

	backend.LogWarning("warning with context", "key1", "value1")

	output := decodeLastLogRecord(t, &buf)
	assert.Equal(t, "WARN", output["level"])
	assert.Equal(t, "warning with context", output["msg"])
	assert.Equal(t, "value1", output["key1"])
}

func TestLoggingBackendLogWarningWithInvalidKeyType(t *testing.T) {
	var buf bytes.Buffer
	backend := newLoggingBackendForTest(&buf, slog.LevelDebug)

	backend.LogWarning("warning with invalid key", 123, "value1")

	output := decodeLastLogRecord(t, &buf)
	assert.Equal(t, "WARN", output["level"])
	assert.Equal(t, invalidKeyTypeMessage, output["msg"])
	assert.Equal(t, float64(123), output["key"])
}

func newLoggingBackendForTest(buf *bytes.Buffer, level slog.Level) *LoggingBackendImpl {
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level})
	return &LoggingBackendImpl{slog: slog.New(handler)}
}

func decodeLastLogRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	records := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var output map[string]any
	err := json.Unmarshal([]byte(records[len(records)-1]), &output)
	assert.NoError(t, err)
	return output
}
