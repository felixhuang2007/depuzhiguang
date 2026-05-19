package table

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayer_NewPlayer(t *testing.T) {
	p := NewPlayer("p1", 3, 1000)
	assert.Equal(t, "p1", p.ID)
	assert.Equal(t, 3, p.Seat)
	assert.Equal(t, 1000, p.Stack)
	assert.Equal(t, Waiting, p.Status)
}

func TestPlayer_IsActive(t *testing.T) {
	p := NewPlayer("p1", 0, 100)
	assert.True(t, p.IsActive())

	p.Status = Folded
	assert.False(t, p.IsActive())

	p.Status = AllIn
	assert.False(t, p.IsActive())
}

func TestPlayer_IsInHand(t *testing.T) {
	p := NewPlayer("p1", 0, 100)
	assert.False(t, p.IsInHand()) // Waiting is not in hand

	p.Status = Active
	assert.True(t, p.IsInHand())

	p.Status = Folded
	assert.False(t, p.IsInHand())

	p.Status = AllIn
	assert.True(t, p.IsInHand())
}

func TestPlayer_ResetForNewHand(t *testing.T) {
	p := NewPlayer("p1", 0, 100)
	p.Status = Folded
	p.Bet = 50
	p.TotalBet = 100

	p.ResetForNewHand()
	assert.Equal(t, Waiting, p.Status)
	assert.Equal(t, 0, p.Bet)
	assert.Equal(t, 0, p.TotalBet)
}
