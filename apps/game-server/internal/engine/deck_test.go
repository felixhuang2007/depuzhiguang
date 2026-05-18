package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDeck_Has52Cards(t *testing.T) {
	d := NewDeck()
	assert.Equal(t, 52, d.Len())
}

func TestNewDeck_AllCardsUnique(t *testing.T) {
	d := NewDeck()
	seen := make(map[Card]bool)
	for d.Len() > 0 {
		c, ok := d.Deal()
		assert.True(t, ok)
		assert.False(t, seen[c], "duplicate card: %v", c)
		seen[c] = true
	}
	assert.Equal(t, 52, len(seen))
}

func TestDeck_Shuffle(t *testing.T) {
	d1 := NewDeck()
	d2 := NewDeck()
	d2.Shuffle()

	same := 0
	for i := 0; i < 52; i++ {
		c1, _ := d1.Deal()
		c2, _ := d2.Deal()
		if c1 == c2 {
			same++
		}
	}
	assert.Less(t, same, 6, "deck was not shuffled")
}

func TestDeck_DealEmpty(t *testing.T) {
	d := NewDeck()
	for d.Len() > 0 {
		d.Deal()
	}
	_, ok := d.Deal()
	assert.False(t, ok)
}
