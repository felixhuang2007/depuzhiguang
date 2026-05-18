package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCard(t *testing.T) {
	c := NewCard(Spades, Ace)
	assert.Equal(t, Spades, c.Suit())
	assert.Equal(t, Ace, c.Rank())
	assert.Equal(t, "A♠", c.String())
}

func TestCardInvalid(t *testing.T) {
	c := NewCard(5, 15) // invalid suit and rank
	assert.Equal(t, InvalidSuit, c.Suit())
	assert.Equal(t, InvalidRank, c.Rank())
}

func TestCardComparison(t *testing.T) {
	aceSpades := NewCard(Spades, Ace)
	kingHearts := NewCard(Hearts, King)
	assert.True(t, aceSpades.Rank() > kingHearts.Rank())
}
