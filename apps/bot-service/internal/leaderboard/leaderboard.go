package leaderboard

import (
	"sort"
	"sync"
)

type Entry struct {
	UserID   string
	Username string
	Value    float64
	Rank     int
}

type Leaderboard struct {
	data map[string]map[string]*Entry // metric -> userId -> entry
	mu   sync.RWMutex
}

func NewLeaderboard() *Leaderboard {
	return &Leaderboard{
		data: make(map[string]map[string]*Entry),
	}
}

func (lb *Leaderboard) Update(userID, username, metric string, delta float64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.data[metric] == nil {
		lb.data[metric] = make(map[string]*Entry)
	}
	if lb.data[metric][userID] == nil {
		lb.data[metric][userID] = &Entry{UserID: userID, Username: username}
	}
	lb.data[metric][userID].Value += delta
	lb.recalc(metric)
}

func (lb *Leaderboard) recalc(metric string) {
	entries := make([]*Entry, 0, len(lb.data[metric]))
	for _, e := range lb.data[metric] {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Value > entries[j].Value
	})
	for i, e := range entries {
		e.Rank = i + 1
	}
}

func (lb *Leaderboard) GetRanking(metric string) []*Entry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	entries := make([]*Entry, 0, len(lb.data[metric]))
	for _, e := range lb.data[metric] {
		entries = append(entries, &Entry{
			UserID:   e.UserID,
			Username: e.Username,
			Value:    e.Value,
			Rank:     e.Rank,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Value > entries[j].Value
	})
	return entries
}
