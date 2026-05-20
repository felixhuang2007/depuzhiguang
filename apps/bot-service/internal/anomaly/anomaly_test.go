package anomaly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetector_HighWinRate(t *testing.T) {
	d := NewDetector()
	// 50 wins out of 60 hands = 83% winrate
	for i := 0; i < 60; i++ {
		d.RecordResult("user_1", i < 50)
	}
	anomalies := d.Check()
	assert.Len(t, anomalies, 1)
	assert.Equal(t, "high_winrate", anomalies[0].Type)
}

func TestDetector_BotStuck(t *testing.T) {
	d := NewDetector()
	for i := 0; i < 25; i++ {
		d.RecordAction("user_1", "fold")
	}
	anomalies := d.Check()
	assert.True(t, len(anomalies) > 0)
}
