package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandRank_String(t *testing.T) {
	assert.Equal(t, "Royal Flush", RoyalFlush.String())
	assert.Equal(t, "One Pair", OnePair.String())
	assert.Equal(t, "High Card", HighCard.String())
	assert.Equal(t, "Four of a Kind", FourOfAKind.String())
	assert.Equal(t, "Straight Flush", StraightFlush.String())
}
