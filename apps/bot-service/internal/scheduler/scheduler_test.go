package scheduler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_Assign(t *testing.T) {
	userIDs := make([]string, 20)
	for i := 0; i < 20; i++ {
		userIDs[i] = fmt.Sprintf("user_%d", i)
	}

	s, err := NewScheduler(userIDs, 3, 5, 7)
	require.NoError(t, err)
	tables := s.Assign()

	assert.Len(t, tables, 3)
	total := 0
	for _, table := range tables {
		assert.True(t, len(table.Users) >= 5 && len(table.Users) <= 7, "table size %d out of range", len(table.Users))
		total += len(table.Users)
	}
	assert.Equal(t, 20, total)
}

func TestScheduler_Rotate(t *testing.T) {
	userIDs := []string{"a", "b", "c", "d", "e", "f"}
	s, err := NewScheduler(userIDs, 1, 5, 6)
	require.NoError(t, err)
	first := s.Assign()
	s.Rotate()
	second := s.GetTables()

	assert.Equal(t, 1, len(second))
	assert.Equal(t, len(first[0].Users), len(second[0].Users))
	// Order should differ after rotation (very high probability)
	assert.NotEqual(t, first[0].Users, second[0].Users)
}

func TestScheduler_ShouldRotate(t *testing.T) {
	userIDs := []string{"a", "b", "c", "d", "e", "f"}
	s, err := NewScheduler(userIDs, 1, 5, 6)
	require.NoError(t, err)

	assert.False(t, s.ShouldRotate())
	for i := 0; i < 20; i++ {
		s.RecordHandPlayed()
	}
	assert.True(t, s.ShouldRotate())
	s.Rotate()
	assert.False(t, s.ShouldRotate())
}

func TestScheduler_NewScheduler_Invalid(t *testing.T) {
	userIDs := []string{"a", "b", "c"}
	_, err := NewScheduler(userIDs, 2, 5, 6)
	require.Error(t, err)
}

func TestScheduler_GetTables_Copy(t *testing.T) {
	userIDs := []string{"a", "b", "c", "d", "e", "f"}
	s, err := NewScheduler(userIDs, 1, 5, 6)
	require.NoError(t, err)
	s.Assign()
	copy1 := s.GetTables()
	copy2 := s.GetTables()
	assert.NotSame(t, &copy1, &copy2)
}
