package anomaly

import (
	"fmt"
	"sync"
)

type Anomaly struct {
	Type        string
	Severity    string
	Description string
	Data        map[string]interface{}
}

type userStats struct {
	handsPlayed      int
	handsWon         int
	consecutiveFolds int
	tableWins        map[string]int
}

type Detector struct {
	stats map[string]*userStats
	mu    sync.RWMutex
}

func NewDetector() *Detector {
	return &Detector{stats: make(map[string]*userStats)}
}

func (d *Detector) RecordResult(userID string, won bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stats[userID] == nil {
		d.stats[userID] = &userStats{tableWins: make(map[string]int)}
	}
	d.stats[userID].handsPlayed++
	if won {
		d.stats[userID].handsWon++
	}
}

func (d *Detector) RecordAction(userID, action string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stats[userID] == nil {
		d.stats[userID] = &userStats{tableWins: make(map[string]int)}
	}
	if action == "fold" {
		d.stats[userID].consecutiveFolds++
	} else {
		d.stats[userID].consecutiveFolds = 0
	}
}

func (d *Detector) RecordTableWin(tableID, userID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stats[userID] == nil {
		d.stats[userID] = &userStats{tableWins: make(map[string]int)}
	}
	d.stats[userID].tableWins[tableID]++
}

func (d *Detector) Check() []*Anomaly {
	d.mu.Lock()
	defer d.mu.Unlock()

	var anomalies []*Anomaly

	for uid, s := range d.stats {
		// Rule 1: High winrate
		if s.handsPlayed >= 50 {
			winRate := float64(s.handsWon) / float64(s.handsPlayed)
			if winRate > 0.40 {
				anomalies = append(anomalies, &Anomaly{
					Type:        "high_winrate",
					Severity:    "warning",
					Description: fmt.Sprintf("User %s winrate %.1f%% over %d hands", uid, winRate*100, s.handsPlayed),
					Data:        map[string]interface{}{"userId": uid, "winRate": winRate, "handsPlayed": s.handsPlayed},
				})
				s.handsPlayed = 0
				s.handsWon = 0
			}
		}

		// Rule 2: Bot stuck (consecutive folds)
		if s.consecutiveFolds >= 20 {
			anomalies = append(anomalies, &Anomaly{
				Type:        "bot_stuck",
				Severity:    "warning",
				Description: fmt.Sprintf("User %s folded %d consecutive hands", uid, s.consecutiveFolds),
				Data:        map[string]interface{}{"userId": uid, "consecutiveFolds": s.consecutiveFolds},
			})
			s.consecutiveFolds = 0
		}

		// Rule 3: Table bias
		for tid, wins := range s.tableWins {
			if wins >= 7 {
				anomalies = append(anomalies, &Anomaly{
					Type:        "table_bias",
					Severity:    "warning",
					Description: fmt.Sprintf("User %s won %d times at table %s", uid, wins, tid),
					Data:        map[string]interface{}{"userId": uid, "tableId": tid, "wins": wins},
				})
				delete(s.tableWins, tid)
			}
		}
	}

	return anomalies
}

func (d *Detector) ResetTableWins() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, s := range d.stats {
		s.tableWins = make(map[string]int)
	}
}
