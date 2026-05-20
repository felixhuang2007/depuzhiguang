package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	l := New("test")
	assert.NotNil(t, l)
	l.Info("test message", "key", "value")
}
