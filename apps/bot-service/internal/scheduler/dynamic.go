package scheduler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type TableSnapshot struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MaxSeats int    `json:"max_seats"`
	Occupied int    `json:"occupied"`
	Status   string `json:"status"`
}

type ScheduleAction struct {
	Type    string // "assign", "unassign", "refill"
	BotID   string
	TableID string
	Amount  int
}

type DynamicScheduler struct {
	apiBaseURL      string
	client          *http.Client
	botIDs          []string
	active          map[string]*ActiveBot
	standby         map[string]struct{}
	mu              sync.RWMutex
	tickInterval    time.Duration
	minStandby      int
	targetTableSize int
	handsMin        int
	handsMax        int
	refillThreshold int
}

type ActiveBot struct {
	BotID       string
	TableID     string
	MaxHands    int
	HandsPlayed int
}

func NewDynamicScheduler(apiBaseURL string, botIDs []string) *DynamicScheduler {
	ds := &DynamicScheduler{
		apiBaseURL:      apiBaseURL,
		client:          &http.Client{Timeout: 5 * time.Second},
		botIDs:          botIDs,
		active:          make(map[string]*ActiveBot),
		standby:         make(map[string]struct{}),
		tickInterval:    5 * time.Second,
		minStandby:      5,
		targetTableSize: 5,
		handsMin:        5,
		handsMax:        20,
		refillThreshold: 500,
	}
	for _, id := range botIDs {
		ds.standby[id] = struct{}{}
	}
	return ds
}

func (ds *DynamicScheduler) Tick() []ScheduleAction {
	tables, err := ds.fetchTables()
	if err != nil {
		return nil
	}

	ds.mu.Lock()
	defer ds.mu.Unlock()

	var actions []ScheduleAction

	// 1. Unassign bots that reached max hands
	for botID, ab := range ds.active {
		if ab.HandsPlayed >= ab.MaxHands {
			actions = append(actions, ScheduleAction{Type: "unassign", BotID: botID})
			delete(ds.active, botID)
			ds.standby[botID] = struct{}{}
		}
	}

	// 2. Assign standby bots to tables that need filling
	for _, t := range tables {
		if t.Occupied >= ds.targetTableSize {
			continue
		}
		needed := ds.targetTableSize - t.Occupied
		for i := 0; i < needed; i++ {
			if len(ds.standby) <= ds.minStandby {
				break
			}
			botID := ds.pickStandby()
			if botID == "" {
				break
			}
			maxHands := ds.handsMin + rand.Intn(ds.handsMax-ds.handsMin+1)
			ds.active[botID] = &ActiveBot{
				BotID:    botID,
				TableID:  t.ID,
				MaxHands: maxHands,
			}
			delete(ds.standby, botID)
			actions = append(actions, ScheduleAction{Type: "assign", BotID: botID, TableID: t.ID})
		}
	}

	return actions
}

func (ds *DynamicScheduler) pickStandby() string {
	for id := range ds.standby {
		return id
	}
	return ""
}

func (ds *DynamicScheduler) fetchTables() ([]TableSnapshot, error) {
	resp, err := ds.client.Get(ds.apiBaseURL + "/lobby/tables")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var body struct {
		Tables []TableSnapshot `json:"tables"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body.Tables, nil
}

func (ds *DynamicScheduler) RecordHandPlayed(botID string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if ab, ok := ds.active[botID]; ok {
		ab.HandsPlayed++
	}
}

func (ds *DynamicScheduler) ActiveBots() map[string]*ActiveBot {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	out := make(map[string]*ActiveBot, len(ds.active))
	for k, v := range ds.active {
		cp := *v
		out[k] = &cp
	}
	return out
}

func (ds *DynamicScheduler) StandbyCount() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.standby)
}
