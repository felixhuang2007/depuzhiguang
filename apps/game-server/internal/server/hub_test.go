package server

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewHub(t *testing.T) {
	h := NewHub()
	assert.NotNil(t, h)
	assert.NotNil(t, h.connections)
	assert.NotNil(t, h.tablePlayers)
}

func TestHub_RegisterUnregister(t *testing.T) {
	h := NewHub()
	// Use a mock conn (nil is fine for Register/Unregister logic)
	h.Register("p1", nil)
	assert.Equal(t, 1, len(h.connections))

	h.Unregister("p1")
	assert.Equal(t, 0, len(h.connections))
}

func TestHub_JoinLeaveTable(t *testing.T) {
	h := NewHub()
	h.Register("p1", nil)
	h.Register("p2", nil)

	h.JoinTable("p1", "table1")
	h.JoinTable("p2", "table1")

	assert.Equal(t, 2, len(h.tablePlayers["table1"]))

	h.LeaveTable("p1", "table1")
	assert.Equal(t, 1, len(h.tablePlayers["table1"]))

	h.LeaveTable("p2", "table1")
	assert.Nil(t, h.tablePlayers["table1"])
}

func TestHub_SendToPlayer(t *testing.T) {
	h := NewHub()
	// Sending to unregistered player should not panic
	err := h.SendToPlayer("p1", Message{Type: MsgPong})
	assert.NoError(t, err)
}

// mockConn is a minimal websocket.Conn wrapper for testing
func newMockConn() *websocket.Conn {
	return nil // Hub handles nil gracefully in Send/Broadcast
}
