package scheduler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynamicScheduler_Tick_AssignBots(t *testing.T) {
	// Return 2 tables with 2 and 1 players, target is 5
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/lobby/tables", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"tables":[{"id":"t1","name":"T1","max_seats":7,"occupied":2,"status":"waiting"},{"id":"t2","name":"T2","max_seats":7,"occupied":1,"status":"waiting"}]}`)
	}))
	defer srv.Close()

	botIDs := make([]string, 20)
	for i := 0; i < 20; i++ {
		botIDs[i] = fmt.Sprintf("bot-%d", i)
	}

	ds := NewDynamicScheduler(srv.URL, botIDs)
	actions := ds.Tick()

	// t1 needs 3, t2 needs 4 -> 7 assign actions
	assignCount := 0
	for _, a := range actions {
		if a.Type == "assign" {
			assignCount++
		}
	}
	assert.Equal(t, 7, assignCount)
	assert.Equal(t, 20-7, ds.StandbyCount())
}

func TestDynamicScheduler_RecordHandPlayed(t *testing.T) {
	botIDs := []string{"bot-0"}
	ds := NewDynamicScheduler("http://localhost", botIDs)
	ds.active["bot-0"] = &ActiveBot{BotID: "bot-0", TableID: "t1", MaxHands: 3, HandsPlayed: 0}
	delete(ds.standby, "bot-0")

	ds.RecordHandPlayed("bot-0")
	ds.RecordHandPlayed("bot-0")
	ab := ds.ActiveBots()["bot-0"]
	assert.Equal(t, 2, ab.HandsPlayed)

	// Third hand should trigger unassign on next Tick
	ds.RecordHandPlayed("bot-0")
	// Need a mock server for Tick to work
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"tables":[]}`)
	}))
	defer srv.Close()

	ds.apiBaseURL = srv.URL
	actions := ds.Tick()
	assert.Len(t, actions, 1)
	assert.Equal(t, "unassign", actions[0].Type)
	assert.Equal(t, "bot-0", actions[0].BotID)
	assert.Equal(t, 1, ds.StandbyCount())
}
