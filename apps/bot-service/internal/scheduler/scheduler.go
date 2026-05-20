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
	lastRotatedAt    int
	rotationInterval int
	mu               sync.RWMutex
	rng              *rand.Rand
}

func NewScheduler(userIDs []string, tableCount, minPerTable, maxPerTable int) (*Scheduler, error) {
	if tableCount <= 0 {
		return nil, fmt.Errorf("tableCount must be > 0")
	}
	if minPerTable <= 0 || maxPerTable < minPerTable {
		return nil, fmt.Errorf("invalid min/max per table: %d/%d", minPerTable, maxPerTable)
	}
	if len(userIDs) < tableCount*minPerTable {
		return nil, fmt.Errorf("not enough users (%d) for %d tables with min %d", len(userIDs), tableCount, minPerTable)
	}
	if len(userIDs) > tableCount*maxPerTable {
		return nil, fmt.Errorf("too many users (%d) for %d tables with max %d", len(userIDs), tableCount, maxPerTable)
	}
	return &Scheduler{
		userIDs:          userIDs,
		tableCount:       tableCount,
		minPerTable:      minPerTable,
		maxPerTable:      maxPerTable,
		rotationInterval: 20,
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (s *Scheduler) Assign() []TableAssignment {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.assignUnsafe()
}

func (s *Scheduler) assignUnsafe() []TableAssignment {
	shuffled := make([]string, len(s.userIDs))
	copy(shuffled, s.userIDs)
	s.rng.Shuffle(len(shuffled), func(i, j int) {
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
	s.lastRotatedAt = s.handsPlayed
	s.assignUnsafe()
}

func (s *Scheduler) ShouldRotate() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.handsPlayed > 0 && s.handsPlayed-s.lastRotatedAt >= s.rotationInterval
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
