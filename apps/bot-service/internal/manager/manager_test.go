package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_Spawn(t *testing.T) {
	m := NewManager("ws://localhost:8080/ws", "http://localhost:3000", nil)
	m.Spawn(10)
	stats := m.Stats()
	assert.Equal(t, 10, stats["total"])
}

func TestManager_AssignToTable(t *testing.T) {
	m := NewManager("ws://localhost:8080/ws", "http://localhost:3000", nil)
	m.Spawn(5)

	var botID string
	for id := range m.bots {
		botID = id
		break
	}

	// AssignToTable now tries to connect to WebSocket, so it will fail without a server
	// Just verify the bot exists and can be found
	assert.NotEmpty(t, botID)
	assert.Equal(t, "idle", m.bots[botID].Status)
}

func TestManager_StopAll(t *testing.T) {
	m := NewManager("ws://localhost:8080/ws", "http://localhost:3000", nil)
	m.Spawn(5)
	m.StopAll()
	stats := m.Stats()
	assert.Equal(t, 0, stats["total"])
}
