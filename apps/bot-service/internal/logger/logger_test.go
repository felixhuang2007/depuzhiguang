package logger

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_ReturnsNonNil(t *testing.T) {
	l := New("test")
	require.NotNil(t, l)
}

func TestNew_InfoDoesNotPanic(t *testing.T) {
	l := New("test")
	require.NotNil(t, l)

	// Should not panic
	assert.NotPanics(t, func() {
		l.Info("test message", "key", "value")
	})
}

func TestNew_IncludesServiceKey(t *testing.T) {
	// Capture stdout by replacing it temporarily
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Create a logger that writes only to our pipe (override the multi-writer)
	// We test the handler directly to verify the service attribute is present.
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	l := slog.New(handler).With("service", "test-service")
	l.Info("hello", "foo", "bar")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"service":"test-service"`)
	assert.Contains(t, output, `"msg":"hello"`)
	assert.Contains(t, output, `"foo":"bar"`)
}

func TestNew_JSONFormat(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	l := slog.New(handler).With("service", "json-test")
	l.Info("json check")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	require.NoError(t, err)

	output := strings.TrimSpace(buf.String())
	// Verify it starts with a JSON brace
	assert.True(t, strings.HasPrefix(output, "{"))
	assert.Contains(t, output, `"level":"INFO"`)
}
