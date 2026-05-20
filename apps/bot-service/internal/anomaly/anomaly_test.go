package anomaly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetector_HighWinRate(t *testing.T) {
	d := NewDetector()
	for i := 0; i < 60; i++ {
		d.RecordResult("user_1", i < 50)
	}
	anomalies := d.Check()
	assert.Len(t, anomalies, 1)
	assert.Equal(t, "high_winrate", anomalies[0].Type)

	// After reset, should not trigger again with same stats
	anomalies = d.Check()
	assert.Empty(t, anomalies)
}

func TestDetector_BotStuck(t *testing.T) {
	d := NewDetector()
	for i := 0; i < 25; i++ {
		d.RecordAction("user_1", "fold")
	}
	anomalies := d.Check()
	assert.Len(t, anomalies, 1)
	assert.Equal(t, "bot_stuck", anomalies[0].Type)
}

func TestDetector_TableBias(t *testing.T) {
	d := NewDetector()
	for i := 0; i < 7; i++ {
		d.RecordTableWin("table_1", "user_1")
	}
	anomalies := d.Check()
	assert.Len(t, anomalies, 1)
	assert.Equal(t, "table_bias", anomalies[0].Type)
	assert.Equal(t, "table_1", anomalies[0].Data["tableId"])
}

func TestDetector_ResetTableWins(t *testing.T) {
	d := NewDetector()
	for i := 0; i < 7; i++ {
		d.RecordTableWin("table_1", "user_1")
	}
	d.ResetTableWins()
	anomalies := d.Check()
	assert.Empty(t, anomalies)
}
