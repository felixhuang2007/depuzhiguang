package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_Spawn(t *testing.T) {
	m := NewManager()
	m.Spawn(10)
	stats := m.Stats()
	assert.Equal(t, 10, stats["total"])
}

func TestManager_AssignToTable(t *testing.T) {
	m := NewManager()
	m.Spawn(5)

	var botID string
	for id := range m.bots {
		botID = id
		break
	}

	err := m.AssignToTable(botID, "table_1")
	assert.NoError(t, err)
	assert.Equal(t, "playing", m.bots[botID].Status)
}

func TestManager_StopAll(t *testing.T) {
	m := NewManager()
	m.Spawn(5)
	m.StopAll()
	stats := m.Stats()
	assert.Equal(t, 0, stats["total"])
}
