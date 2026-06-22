package deepstack

import (
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRawConsoleLoggingVisually(t *testing.T) {
	logger := NewDeepStackLogger(NewRawConsoleHandler(slog.LevelDebug))
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	logger.Info("This is an info message", "key1", "value1", "key2", "value 2")
	logger.Error(logger.NewError("some-error", "key1", "value1"))
	logger.Error(errors.New("some-error"))
}

func TestRawConsoleLoggingVisuallyWrappedLine(t *testing.T) {
	logger := NewDeepStackLogger(NewRawConsoleHandler(slog.LevelDebug))
	longMessage := strings.Repeat("this-is-a-very-long-log-message-", 8)
	longValue := strings.Repeat("wrap-me-", 20)

	logger.Info(longMessage, "long_value", longValue)
}

func TestJsonConsoleLoggingVisually(t *testing.T) {
	logger := NewDeepStackLogger(NewJsonConsoleHandler(slog.LevelDebug))
	logger.Info("This is an info message", "key1", "value1", "key2", "value 2")
}

func newLogger(t *testing.T) (*DeepStackLoggerImpl, *LoggingBackendMock, *StackTracerMock) {
	loggingBackendMock := NewLoggingBackendMock(t)
	stackTracerMock := NewStackTracerMock(t)
	return &DeepStackLoggerImpl{
		logger:      loggingBackendMock,
		stackTracer: stackTracerMock,
	}, loggingBackendMock, stackTracerMock
}

func TestLogSkip(t *testing.T) {
	logger, backendMock, _ := newLogger(t)
	backendMock.EXPECT().ShouldLogBeSkipped(slog.LevelDebug).Return(true)
	logger.log(slog.LevelDebug, "msg")
	backendMock.AssertExpectations(t)
}

func TestLogDeepStackError(t *testing.T) {
	logger, backendMock, _ := newLogger(t)

	err := &DeepStackError{
		Message:    "some-error-cause",
		StackTrace: "trace",
		Context:    map[string]any{"key1": "value1"},
	}

	backendMock.EXPECT().ShouldLogBeSkipped(slog.LevelError).Return(false)

	expectedLogRecord := &Record{
		level:      slog.LevelError,
		msg:        "some-error-cause",
		attributes: []slog.Attr{slog.Any("key1", "value1"), slog.Any("stack_trace", "trace")},
	}
	backendMock.EXPECT().LogRecord(expectedLogRecord)

	logger.log(slog.LevelError, err)
	backendMock.AssertExpectations(t)
}

func TestLogNormalError(t *testing.T) {
	l, m, _ := newLogger(t)
	m.EXPECT().ShouldLogBeSkipped(slog.LevelError).Return(false)

	expected := &Record{
		level:      slog.LevelError,
		msg:        "some error",
		attributes: []slog.Attr{},
	}
	m.EXPECT().LogRecord(expected)

	l.log(slog.LevelError, errors.New("some error"))
	m.AssertExpectations(t)
}

func TestLogInvalidKeyType(t *testing.T) {
	l, m, _ := newLogger(t)

	expected := &Record{
		level:      slog.LevelInfo,
		msg:        "msg",
		attributes: []slog.Attr{slog.Any("key2", "value2")},
	}

	m.EXPECT().ShouldLogBeSkipped(slog.LevelInfo).Return(false)
	m.EXPECT().LogWarning(invalidKeyTypeMessage, []any{actualTypeField, reflect.TypeOf(0).String()})
	m.EXPECT().LogRecord(expected)

	l.log(slog.LevelInfo, "msg", 123, "value1", "key2", "value2")
	m.AssertExpectations(t)
}

func TestAddContextNormalError(t *testing.T) {
	logger, backendMock, stackTracerMock := newLogger(t)
	inputError := errors.New("some error")
	backendMock.EXPECT().LogWarning(invalidErrorTypeMessage)
	stackTracerMock.EXPECT().GetStackTrace().Return("some-stack-trace")
	createAndAssertDeepstackError(t, logger, inputError)
	backendMock.AssertExpectations(t)
	stackTracerMock.AssertExpectations(t)
}

func createAndAssertDeepstackError(t *testing.T, l *DeepStackLoggerImpl, inputError error) {
	outputError := l.AddContext(inputError, "key1", "value1", "key2", "value2")

	err, ok := outputError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "some error", err.Message)
	assert.Equal(t, 2, len(err.Context))
	assert.Equal(t, "value1", err.Context["key1"])
	assert.Equal(t, "value2", err.Context["key2"])
	assert.Equal(t, "some-stack-trace", err.StackTrace)
}

func TestAddContextDeepStackError(t *testing.T) {
	logger, backendMock, _ := newLogger(t)
	inputError := &DeepStackError{
		Message:    "some error",
		StackTrace: "some-stack-trace",
		Context:    map[string]any{"key1": "value1"},
	}
	outputError := logger.AddContext(inputError, "key2", "value2")

	err, ok := outputError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "some error", err.Message)
	assert.Equal(t, 2, len(err.Context))
	assert.Equal(t, "value1", err.Context["key1"])
	assert.Equal(t, "value2", err.Context["key2"])
	assert.Equal(t, "some-stack-trace", err.StackTrace)

	backendMock.AssertExpectations(t)
}

func TestAddContextDeepStackError_DisabledWarnings(t *testing.T) {
	logger, backendMock, stackTracerMock := newLogger(t)
	stackTracerMock.EXPECT().GetStackTrace().Return("some-stack-trace")
	inputError := logger.NewError("some-error")

	backendMock.EXPECT().LogWarning(invalidKeyTypeMessage, []any{"actual_type", "int"})
	outputError := logger.AddContext(inputError, 1234, "key1", "key2", "value2")

	outputDeepstackError, ok := outputError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "some-error", outputDeepstackError.Message)
	assert.Equal(t, 1, len(outputDeepstackError.Context))
	assert.Equal(t, "value2", outputDeepstackError.Context["key2"])
	assert.Equal(t, "some-stack-trace", outputDeepstackError.StackTrace)
	backendMock.AssertExpectations(t)
	stackTracerMock.AssertExpectations(t)
}

func TestLogOddNumberOfKeyValues(t *testing.T) {
	logger, backendMock, _ := newLogger(t)
	backendMock.EXPECT().ShouldLogBeSkipped(slog.LevelInfo).Return(false)
	backendMock.EXPECT().LogWarning(oddKeyValuePairNumberMessage)
	backendMock.EXPECT().LogRecord(mock.Anything)
	logger.Info("some-message", "key1", "value1", "key2")
	backendMock.AssertExpectations(t)
}

func TestAddContextOddNumberOfKeyValues(t *testing.T) {
	logger, backendMock, _ := newLogger(t)
	backendMock.EXPECT().LogWarning(oddKeyValuePairNumberMessage)
	enrichedError := logger.AddContext(getSampleDeeptstackError(), "key1", "value1", "key2")
	backendMock.AssertExpectations(t)

	enrichedDeepStackError, ok := enrichedError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, 1, len(enrichedDeepStackError.Context))
	assert.Equal(t, "value1", enrichedDeepStackError.Context["key1"])
}

func getSampleDeeptstackError() *DeepStackError {
	deepstackError := &DeepStackError{
		Message:    "some-error",
		StackTrace: "some-stack-trace",
		Context:    map[string]any{},
	}
	return deepstackError
}

func TestEmptyKeyWarning(t *testing.T) {
	logger, backendMock, _ := newLogger(t)
	backendMock.EXPECT().LogWarning(emptySpacesInKeyMessage, []any{"key", "key 1"})
	enrichedError := logger.AddContext(getSampleDeeptstackError(), "key 1", "value1")
	backendMock.AssertExpectations(t)

	enrichedDeepStackError, ok := enrichedError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, 0, len(enrichedDeepStackError.Context))
}

func TestLogUnknownTypeWarning(t *testing.T) {
	logger, backendMock, _ := newLogger(t)
	backendMock.EXPECT().ShouldLogBeSkipped(slog.LevelInfo).Return(false)
	backendMock.EXPECT().LogWarning("an unknown type was passed to log function", []any{"unknown_msg_type", "struct {}"})

	logger.log(slog.LevelInfo, struct{}{})
	backendMock.AssertExpectations(t)
}

func TestLogSortsContextDeterministically(t *testing.T) {
	logger, backendMock, _ := newLogger(t)

	expected := &Record{
		level: slog.LevelInfo,
		msg:   "msg",
		attributes: []slog.Attr{
			slog.Any("key1", "value1"),
			slog.Any("key2", "value2"),
		},
	}

	backendMock.EXPECT().ShouldLogBeSkipped(slog.LevelInfo).Return(false)
	backendMock.EXPECT().LogRecord(expected)

	logger.Info("msg", "key2", "value2", "key1", "value1")
	backendMock.AssertExpectations(t)
}

func TestAddContextDeepStackError_DuplicateKeyKeepsOriginalValue(t *testing.T) {
	logger, backendMock, _ := newLogger(t)
	inputError := &DeepStackError{
		Message:    "some error",
		StackTrace: "some-stack-trace",
		Context:    map[string]any{"key1": "original"},
	}

	backendMock.EXPECT().LogWarning(duplicateContextFieldMessage, []any{keyField, "key1"})

	outputError := logger.AddContext(inputError, "key1", "replacement", "key2", "value2")

	err, ok := outputError.(*DeepStackError)
	assert.True(t, ok)
	assert.Equal(t, "some error", err.Message)
	assert.Equal(t, "some-stack-trace", err.StackTrace)
	assert.Equal(t, 2, len(err.Context))
	assert.Equal(t, "original", err.Context["key1"])
	assert.Equal(t, "value2", err.Context["key2"])
	backendMock.AssertExpectations(t)
}
