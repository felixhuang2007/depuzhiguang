package scheduler

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type TableAssignment struct {
	TableID string
	Users   []string
}

type Scheduler struct {
	userIDs          []string
	tableCount       int
	minPerTable      int
	maxPerTable      int
	tables           []TableAssignment
	handsPlayed      int
	rotationInterval int
	mu               sync.RWMutex
}

func NewScheduler(userIDs []string, tableCount, minPerTable, maxPerTable int) *Scheduler {
	return &Scheduler{
		userIDs:          userIDs,
		tableCount:       tableCount,
		minPerTable:      minPerTable,
		maxPerTable:      maxPerTable,
		rotationInterval: 20,
	}
}

func (s *Scheduler) Assign() []TableAssignment {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.assignUnsafe()
}

func (s *Scheduler) assignUnsafe() []TableAssignment {
	rand.Seed(time.Now().UnixNano())
	shuffled := make([]string, len(s.userIDs))
	copy(shuffled, s.userIDs)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	s.tables = make([]TableAssignment, s.tableCount)
	baseSize := len(shuffled) / s.tableCount
	remainder := len(shuffled) % s.tableCount

	idx := 0
	for i := 0; i < s.tableCount; i++ {
		size := baseSize
		if i < remainder {
			size++
		}
		s.tables[i] = TableAssignment{
			TableID: fmt.Sprintf("sim-table-%d", i),
			Users:   shuffled[idx : idx+size],
		}
		idx += size
	}
	return s.tables
}

func (s *Scheduler) Rotate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.assignUnsafe()
}

func (s *Scheduler) ShouldRotate() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.handsPlayed > 0 && s.handsPlayed%s.rotationInterval == 0
}

func (s *Scheduler) RecordHandPlayed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handsPlayed++
}

func (s *Scheduler) GetTables() []TableAssignment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]TableAssignment, len(s.tables))
	copy(result, s.tables)
	return result
}
