package leaderboard

import (
	"sync"
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
	assert.Equal(t, "user_2", ranks[0].UserID)
	assert.Equal(t, 1, ranks[0].Rank)
	assert.Equal(t, "user_1", ranks[1].UserID)
	assert.Equal(t, 2, ranks[1].Rank)
	assert.Equal(t, "user_3", ranks[2].UserID)
	assert.Equal(t, 3, ranks[2].Rank)
}

func TestLeaderboard_MultipleMetrics(t *testing.T) {
	lb := NewLeaderboard()
	lb.Update("u1", "A", "hands_won", 5)
	lb.Update("u1", "A", "gold_earned", 100)

	r1 := lb.GetRanking("hands_won")
	r2 := lb.GetRanking("gold_earned")
	assert.Len(t, r1, 1)
	assert.Len(t, r2, 1)
	assert.Equal(t, 5.0, r1[0].Value)
	assert.Equal(t, 100.0, r2[0].Value)
}

func TestLeaderboard_Empty(t *testing.T) {
	lb := NewLeaderboard()
	ranks := lb.GetRanking("nonexistent")
	assert.Empty(t, ranks)
}

func TestLeaderboard_Concurrent(t *testing.T) {
	lb := NewLeaderboard()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lb.Update("u1", "A", "metric", 1)
		}()
	}
	wg.Wait()
	ranks := lb.GetRanking("metric")
	assert.Len(t, ranks, 1)
	assert.Equal(t, 100.0, ranks[0].Value)
}
