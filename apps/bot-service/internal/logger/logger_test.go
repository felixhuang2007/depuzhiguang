package logger

import (
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
