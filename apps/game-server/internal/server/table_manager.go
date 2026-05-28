package server

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/depuzhiguang/game-server/internal/table"
)

// TableManager manages all active poker tables and routes WebSocket messages to games.
type TableManager struct {
	hub    *Hub
	lobby  *LobbyManager
	tables map[string]*tableManagerEntry
	mu     sync.RWMutex
	logg   *slog.Logger
}

type tableManagerEntry struct {
	Table       *table.Table
	Game        *table.Game
	actionTimer *time.Timer
}

// NewTableManager creates a new table manager.
func NewTableManager(hub *Hub, logg *slog.Logger) *TableManager {
	return &TableManager{
		hub:    hub,
		tables: make(map[string]*tableManagerEntry),
		logg:   logg,
	}
}

// SetLobby sets the lobby manager for broadcasting table updates.
func (tm *TableManager) SetLobby(lobby *LobbyManager) {
	tm.lobby = lobby
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
	tm.logg.Info("Table created",
		slog.String("table_id", config.ID),
		slog.Int("max_seats", config.MaxSeats),
		slog.Int("small_blind", config.SmallBlind),
		slog.Int("big_blind", config.BigBlind))
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

// GetTableList returns a list of all tables with their public metadata.
func (tm *TableManager) GetTableList() []TableInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]TableInfo, 0, len(tm.tables))
	for id, entry := range tm.tables {
		t := entry.Table
		status := "waiting"
		if entry.Game != nil {
			status = "playing"
		}
		result = append(result, TableInfo{
			ID:         id,
			Name:       t.Config.Name,
			MaxSeats:   t.Config.MaxSeats,
			Occupied:   t.PlayerCount(),
			SmallBlind: t.Config.SmallBlind,
			BigBlind:   t.Config.BigBlind,
			Status:     status,
		})
	}
	return result
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
	tm.logg.Info("HandleJoin",
		slog.String("player_id", payload.PlayerID),
		slog.String("table_id", payload.TableID),
		slog.Int("seat", seat),
		slog.Int("occupied", t.PlayerCount()),
		slog.Int("max_seats", t.Config.MaxSeats))
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

	go func() {
		// Broadcast player joined (async to avoid blocking)
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
	}()

	// Auto-start if enough players and no active game
	if entry.Game == nil && t.PlayerCount() >= 2 {
		entry.Game = table.NewGame(t)
		if err := entry.Game.Start(); err != nil {
			tm.logg.Error("Failed to start game",
				slog.String("table_id", payload.TableID),
				slog.String("error", err.Error()))
			entry.Game = nil
		} else {
			tm.logg.Info("Game started", slog.String("table_id", payload.TableID))
			go tm.broadcastGameState(payload.TableID, entry)
		}
	}

	if tm.lobby != nil {
		go tm.lobby.BroadcastTablesUpdate()
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

	// Clear game if not enough players to continue
	if t.PlayerCount() < 2 {
		entry.Game = nil
	}

	if tm.lobby != nil {
		go tm.lobby.BroadcastTablesUpdate()
	}

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

	// Cancel any existing action timer since player is acting
	if entry.actionTimer != nil {
		entry.actionTimer.Stop()
		entry.actionTimer = nil
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

	// Check if hand is complete or reached showdown
	if entry.Game.State == table.StateComplete || entry.Game.State == table.StateShowdown {
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
		TableID:     t.Config.ID,
		Players:     players,
		Pot:         0,
		CurrentTurn: -1,
	}

	if g != nil {
		payload.State = g.State
		payload.Community = g.Community
		payload.CurrentTurn = g.CurrentTurn
		payload.Button = g.Button
		payload.BigBlind = t.Config.BigBlind
		if g.Pot != nil {
			payload.Pot = g.Pot.Total()
		}
		// Minimum raise is at least the big blind or the last raise size
		if g.LastRaise > 0 {
			payload.MinRaise = g.LastRaise
		} else {
			payload.MinRaise = t.Config.BigBlind
		}
	}

	tm.hub.SendToPlayer(playerID, Message{Type: MsgStateSnapshot, Payload: payload})
}

func (tm *TableManager) startActionTimer(tableID string, entry *tableManagerEntry) {
	// Cancel existing timer
	if entry.actionTimer != nil {
		entry.actionTimer.Stop()
		entry.actionTimer = nil
	}

	// Only start timer if game is active and waiting for player action
	if entry.Game == nil || entry.Game.State == table.StateComplete || entry.Game.State == table.StateShowdown {
		return
	}

	currentSeat := entry.Game.CurrentTurn
	if currentSeat < 0 || entry.Table.Seats[currentSeat] == nil {
		return
	}

	player := entry.Table.Seats[currentSeat]

	entry.actionTimer = time.AfterFunc(30*time.Second, func() {
		tm.mu.Lock()
		defer tm.mu.Unlock()

		e, ok := tm.tables[tableID]
		if !ok || e.Game == nil || e.Game != entry.Game {
			return
		}

		// Verify it's still this player's turn
		if e.Game.CurrentTurn != currentSeat {
			return
		}

		if e.Table.Seats[currentSeat] == nil || e.Table.Seats[currentSeat].ID != player.ID {
			return
		}

		tm.logg.Info("Action timeout, auto-folding",
			slog.String("player_id", player.ID),
			slog.String("table_id", tableID))

		action := table.Action{
			PlayerID: player.ID,
			Type:     table.ActionFold,
		}
		if err := e.Game.ProcessAction(action); err == nil {
			tm.broadcastGameState(tableID, e)
			if e.Game.State == table.StateComplete || e.Game.State == table.StateShowdown {
				tm.broadcastHandResult(tableID, e)
				go tm.scheduleNextHand(tableID)
			}
		}
	})
}

func (tm *TableManager) broadcastGameState(tableID string, entry *tableManagerEntry) {
	// Send personalized snapshot to each player
	for _, seat := range entry.Table.OccupiedSeats() {
		p := entry.Table.Seats[seat]
		tm.sendStateSnapshot(p.ID, entry)
	}
	// Start action timer for current player
	tm.startActionTimer(tableID, entry)
}

func (tm *TableManager) broadcastHandResult(tableID string, entry *tableManagerEntry) {
	g := entry.Game

	// Calculate rake from total pot
	totalPot := g.Pot.Total()
	rake := 0
	if totalPot > 0 && entry.Table.Config.RakePercent > 0 {
		rake = int(float64(totalPot) * entry.Table.Config.RakePercent)
		rakeCap := entry.Table.Config.RakeCap * entry.Table.Config.BigBlind
		if rake > rakeCap {
			rake = rakeCap
		}
	}

	var winnerIDs []string
	winnings := make(map[string]int)

	if g.State == table.StateShowdown {
		// Run full showdown with hand evaluation and pot distribution
		results := table.RunShowdown(g)
		if results == nil {
			return
		}
		for _, r := range results {
			if r.WonAmount > 0 {
				winnerIDs = append(winnerIDs, r.PlayerID)
				winnings[r.PlayerID] = r.WonAmount
			}
		}
	} else {
		// Single remaining player wins without showdown
		for _, seat := range entry.Table.OccupiedSeats() {
			p := entry.Table.Seats[seat]
			if p.IsInHand() {
				winnerIDs = append(winnerIDs, p.ID)
				winnings[p.ID] = totalPot
				break
			}
		}
	}

	// Deduct rake proportionally from winners
	if rake > 0 && len(winnerIDs) > 0 {
		totalWon := 0
		for _, amount := range winnings {
			totalWon += amount
		}
		if totalWon > 0 {
			deducted := 0
			for i, pid := range winnerIDs {
				share := winnings[pid]
				deduction := int(float64(share) * float64(rake) / float64(totalWon))
				if i == len(winnerIDs)-1 {
					// Last winner absorbs rounding remainder
					deduction = rake - deducted
				}
				winnings[pid] = share - deduction
				if winnings[pid] < 0 {
					winnings[pid] = 0
				}
				deducted += deduction
			}
		}
	}

	// Add winnings to player stacks
	for pid, amount := range winnings {
		if amount > 0 {
			for _, seat := range entry.Table.OccupiedSeats() {
				p := entry.Table.Seats[seat]
				if p.ID == pid {
					p.Stack += amount
					break
				}
			}
		}
	}

	payload := HandResultPayload{
		TableID:   tableID,
		Winners:   winnerIDs,
		Community: g.Community,
		Pot:       totalPot,
		Rake:      rake,
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

	// Cancel any pending action timer from previous hand
	if entry.actionTimer != nil {
		entry.actionTimer.Stop()
		entry.actionTimer = nil
	}

	entry.Game = table.NewGame(entry.Table)
	if err := entry.Game.Start(); err != nil {
		tm.logg.Error("Failed to restart game",
			slog.String("table_id", tableID),
			slog.String("error", err.Error()))
		entry.Game = nil
		tm.mu.Unlock()
		return
	}
	tm.mu.Unlock()

	tm.broadcastGameState(tableID, entry)
}
