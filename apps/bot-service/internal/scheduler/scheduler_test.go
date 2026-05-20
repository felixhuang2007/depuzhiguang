package scheduler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScheduler_Assign(t *testing.T) {
	userIDs := make([]string, 20)
	for i := 0; i < 20; i++ {
		userIDs[i] = fmt.Sprintf("user_%d", i)
	}

	s := NewScheduler(userIDs, 3, 5, 7)
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
	s := NewScheduler(userIDs, 1, 5, 6)
	s.Assign()
	s.Rotate()
	// After rotation, same users but different order
	assert.Equal(t, 1, len(s.GetTables()))
}
