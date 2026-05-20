package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages WebSocket connections for all tables
type Hub struct {
	// connections maps playerID to their WebSocket connection
	connections map[string]*websocket.Conn
	// writeMu protects concurrent writes to each connection
	writeMu map[string]*sync.Mutex
	// tablePlayers maps tableID to set of connected playerIDs
	tablePlayers map[string]map[string]struct{}
	// observers maps tableID to set of observing playerIDs
	observers map[string]map[string]struct{}
	mu        sync.RWMutex
}

// NewHub creates a new connection hub
func NewHub() *Hub {
	return &Hub{
		connections:  make(map[string]*websocket.Conn),
		writeMu:      make(map[string]*sync.Mutex),
		tablePlayers: make(map[string]map[string]struct{}),
		observers:    make(map[string]map[string]struct{}),
	}
}

// Register adds a player connection
func (h *Hub) Register(playerID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[playerID] = conn
	h.writeMu[playerID] = &sync.Mutex{}
}

// Unregister removes a player connection
func (h *Hub) Unregister(playerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, playerID)
	delete(h.writeMu, playerID)
	// Remove from all tables
	for tableID, players := range h.tablePlayers {
		delete(players, playerID)
		if len(players) == 0 {
			delete(h.tablePlayers, tableID)
		}
	}
	// Remove from all observers
	for tableID, obs := range h.observers {
		delete(obs, playerID)
		if len(obs) == 0 {
			delete(h.observers, tableID)
		}
	}
}

// JoinTable registers a player to a table's broadcast group
func (h *Hub) JoinTable(playerID, tableID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.tablePlayers[tableID] == nil {
		h.tablePlayers[tableID] = make(map[string]struct{})
	}
	h.tablePlayers[tableID][playerID] = struct{}{}
}

// LeaveTable removes a player from a table's broadcast group
func (h *Hub) LeaveTable(playerID, tableID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if players := h.tablePlayers[tableID]; players != nil {
		delete(players, playerID)
		if len(players) == 0 {
			delete(h.tablePlayers, tableID)
		}
	}
}

// JoinAsObserver registers a player as an observer of a table
func (h *Hub) JoinAsObserver(playerID, tableID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.observers[tableID] == nil {
		h.observers[tableID] = make(map[string]struct{})
	}
	h.observers[tableID][playerID] = struct{}{}
}

// LeaveAsObserver removes a player from observers
func (h *Hub) LeaveAsObserver(playerID, tableID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if obs := h.observers[tableID]; obs != nil {
		delete(obs, playerID)
		if len(obs) == 0 {
			delete(h.observers, tableID)
		}
	}
}

// BroadcastToObservers sends a message to all observers of a table
func (h *Hub) BroadcastToObservers(tableID string, msg Message) {
	h.mu.RLock()
	obs := h.observers[tableID]
	conns := make(map[string]*websocket.Conn, len(obs))
	muMap := make(map[string]*sync.Mutex, len(obs))
	for pid := range obs {
		conns[pid] = h.connections[pid]
		muMap[pid] = h.writeMu[pid]
	}
	h.mu.RUnlock()

	for pid, conn := range conns {
		if conn != nil {
			if mu := muMap[pid]; mu != nil {
				mu.Lock()
				_ = conn.WriteJSON(msg)
				mu.Unlock()
			}
		}
	}
}

// SendToPlayer sends a message to a specific player
func (h *Hub) SendToPlayer(playerID string, msg Message) error {
	h.mu.RLock()
	conn := h.connections[playerID]
	mu := h.writeMu[playerID]
	h.mu.RUnlock()
	if conn == nil {
		return nil // player not connected
	}
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}
	return conn.WriteJSON(msg)
}

// BroadcastToTable sends a message to all players at a table
func (h *Hub) BroadcastToTable(tableID string, msg Message) {
	h.mu.RLock()
	players := h.tablePlayers[tableID]
	conns := make(map[string]*websocket.Conn, len(players))
	muMap := make(map[string]*sync.Mutex, len(players))
	for pid := range players {
		conns[pid] = h.connections[pid]
		muMap[pid] = h.writeMu[pid]
	}
	h.mu.RUnlock()

	for pid, conn := range conns {
		if conn != nil {
			if mu := muMap[pid]; mu != nil {
				mu.Lock()
				_ = conn.WriteJSON(msg)
				mu.Unlock()
			}
		}
	}
}

// BroadcastToTableExcept sends to all except one player
func (h *Hub) BroadcastToTableExcept(tableID, excludePlayerID string, msg Message) {
	h.mu.RLock()
	players := h.tablePlayers[tableID]
	conns := make(map[string]*websocket.Conn, len(players))
	muMap := make(map[string]*sync.Mutex, len(players))
	for pid := range players {
		if pid != excludePlayerID {
			conns[pid] = h.connections[pid]
			muMap[pid] = h.writeMu[pid]
		}
	}
	h.mu.RUnlock()

	for pid, conn := range conns {
		if conn != nil {
			if mu := muMap[pid]; mu != nil {
				mu.Lock()
				_ = conn.WriteJSON(msg)
				mu.Unlock()
			}
		}
	}
}
