package server

import (
	"fmt"
	"log"
	"sync"

	"github.com/depuzhiguang/game-server/internal/table"
)

// TableManager manages all active poker tables and routes WebSocket messages to games.
type TableManager struct {
	hub    *Hub
	tables map[string]*tableManagerEntry
	mu     sync.RWMutex
}

type tableManagerEntry struct {
	Table *table.Table
	Game  *table.Game
}

// NewTableManager creates a new table manager.
func NewTableManager(hub *Hub) *TableManager {
	return &TableManager{
		hub:    hub,
		tables: make(map[string]*tableManagerEntry),
	}
}

// CreateTable creates a new table with the given configuration.
func (tm *TableManager) CreateTable(config table.TableConfig) (*table.Table, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tables[config.ID]; exists {
		return nil, fmt.Errorf("table %s already exists", config.ID)
	}

	t := table.NewTable(config)
	tm.tables[config.ID] = &tableManagerEntry{Table: t}
	log.Printf("Table %s created (%d seats, %d/%d blinds)", config.ID, config.MaxSeats, config.SmallBlind, config.BigBlind)
	return t, nil
}

// GetTable returns a table by ID.
func (tm *TableManager) GetTable(tableID string) (*table.Table, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	entry, ok := tm.tables[tableID]
	if !ok {
		return nil, false
	}
	return entry.Table, true
}

// HandleJoin processes a player joining a table.
func (tm *TableManager) HandleJoin(payload JoinTablePayload) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	entry, ok := tm.tables[payload.TableID]
	if !ok {
		return fmt.Errorf("table not found")
	}

	t := entry.Table
	seat := t.NextAvailableSeat()
	if seat < 0 {
		return fmt.Errorf("table is full")
	}

	// Default buy-in: 100 big blinds
	buyIn := 100 * t.Config.BigBlind
	player := table.NewPlayer(payload.PlayerID, seat, buyIn)
	if err := t.Join(player); err != nil {
		return err
	}

	tm.hub.JoinTable(payload.PlayerID, payload.TableID)

	// Broadcast player joined
	tm.hub.BroadcastToTable(payload.TableID, Message{
		Type: MsgPlayerJoined,
		Payload: PlayerEventPayload{
			TableID:  payload.TableID,
			PlayerID: payload.PlayerID,
			Seat:     seat,
		},
	})

	// Send state snapshot to joining player
	tm.sendStateSnapshot(payload.PlayerID, entry)

	// Auto-start if enough players and no active game
	if entry.Game == nil && t.PlayerCount() >= 2 {
		entry.Game = table.NewGame(t)
		if err := entry.Game.Start(); err != nil {
			log.Printf("Failed to start game on table %s: %v", payload.TableID, err)
			entry.Game = nil
		} else {
			log.Printf("Game started on table %s", payload.TableID)
			tm.broadcastGameState(payload.TableID, entry)
		}
	}

	return nil
}

// HandleLeave processes a player leaving a table.
func (tm *TableManager) HandleLeave(payload LeaveTablePayload) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	entry, ok := tm.tables[payload.TableID]
	if !ok {
		return fmt.Errorf("table not found")
	}

	t := entry.Table
	// Find player's seat
	for seat, p := range t.Seats {
		if p != nil && p.ID == payload.PlayerID {
			t.Leave(seat)
			break
		}
	}

	tm.hub.LeaveTable(payload.PlayerID, payload.TableID)

	tm.hub.BroadcastToTable(payload.TableID, Message{
		Type: MsgPlayerLeft,
		Payload: PlayerEventPayload{
			TableID:  payload.TableID,
			PlayerID: payload.PlayerID,
		},
	})

	return nil
}

// HandleAction processes a player action.
func (tm *TableManager) HandleAction(payload ActionPayload) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	entry, ok := tm.tables[payload.TableID]
	if !ok {
		return fmt.Errorf("table not found")
	}

	if entry.Game == nil {
		return fmt.Errorf("no active game")
	}

	action := table.Action{
		PlayerID: payload.PlayerID,
		Type:     payload.Action,
		Amount:   payload.Amount,
	}

	if err := entry.Game.ProcessAction(action); err != nil {
		return err
	}

	tm.broadcastGameState(payload.TableID, entry)

	// Check if hand is complete
	if entry.Game.State == table.StateComplete {
		// Broadcast hand result
		tm.broadcastHandResult(payload.TableID, entry)
		// Reset for next hand after a delay
		go tm.scheduleNextHand(payload.TableID)
	}

	return nil
}

func (tm *TableManager) sendStateSnapshot(playerID string, entry *tableManagerEntry) {
	t := entry.Table
	g := entry.Game

	var players []PlayerState
	for _, seat := range t.OccupiedSeats() {
		p := t.Seats[seat]
		ps := PlayerState{
			ID:     p.ID,
			Seat:   p.Seat,
			Stack:  p.Stack,
			Status: p.Status,
			Bet:    p.Bet,
		}
		// Only send hole cards to the player themselves
		if p.ID == playerID {
			ps.HoleCards = p.Holes[:]
		}
		players = append(players, ps)
	}

	payload := StateSnapshotPayload{
		TableID:   t.Config.ID,
		Players:   players,
		Pot:       0,
		CurrentTurn: -1,
	}

	if g != nil {
		payload.State = g.State
		payload.Community = g.Community
		payload.CurrentTurn = g.CurrentTurn
		payload.Button = g.Button
		if g.Pot != nil {
			payload.Pot = g.Pot.Total()
		}
	}

	tm.hub.SendToPlayer(playerID, Message{Type: MsgStateSnapshot, Payload: payload})
}

func (tm *TableManager) broadcastGameState(tableID string, entry *tableManagerEntry) {
	// Send personalized snapshot to each player
	for _, seat := range entry.Table.OccupiedSeats() {
		p := entry.Table.Seats[seat]
		tm.sendStateSnapshot(p.ID, entry)
	}
}

func (tm *TableManager) broadcastHandResult(tableID string, entry *tableManagerEntry) {
	winners := []string{}
	// Simple: find last player in hand
	for _, seat := range entry.Table.OccupiedSeats() {
		p := entry.Table.Seats[seat]
		if p.IsInHand() {
			winners = append(winners, p.ID)
		}
	}

	payload := HandResultPayload{
		TableID:   tableID,
		Winners:   winners,
		Community: entry.Game.Community,
		Pot:       entry.Game.Pot.Total(),
	}

	tm.hub.BroadcastToTable(tableID, Message{Type: MsgHandResult, Payload: payload})
}

func (tm *TableManager) scheduleNextHand(tableID string) {
	// TODO: Add configurable delay before next hand
	// For now, reset immediately when a new player joins or action is taken
	tm.mu.Lock()
	entry, ok := tm.tables[tableID]
	if !ok || entry.Table.PlayerCount() < 2 {
		tm.mu.Unlock()
		return
	}
	entry.Game = table.NewGame(entry.Table)
	if err := entry.Game.Start(); err != nil {
		log.Printf("Failed to restart game on table %s: %v", tableID, err)
		entry.Game = nil
		tm.mu.Unlock()
		return
	}
	tm.mu.Unlock()

	tm.broadcastGameState(tableID, entry)
}
