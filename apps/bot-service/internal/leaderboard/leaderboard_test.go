package leaderboard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeaderboard_UpdateAndRank(t *testing.T) {
	lb := NewLeaderboard()

	lb.Update("user_1", "Alice", "hands_won", 5)
	lb.Update("user_2", "Bob", "hands_won", 10)
	lb.Update("user_3", "Carol", "hands_won", 3)

	ranks := lb.GetRanking("hands_won")
	assert.Len(t, ranks, 3)
	assert.Equal(t, "user_2", ranks[0].UserID) // Bob highest
	assert.Equal(t, 1, ranks[0].Rank)
}
