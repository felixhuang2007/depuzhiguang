package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages WebSocket connections for all tables
type Hub struct {
	// connections maps playerID to their WebSocket connection
	connections map[string]*websocket.Conn
	// tablePlayers maps tableID to set of connected playerIDs
	tablePlayers map[string]map[string]struct{}
	mu           sync.RWMutex
}

// NewHub creates a new connection hub
func NewHub() *Hub {
	return &Hub{
		connections:  make(map[string]*websocket.Conn),
		tablePlayers: make(map[string]map[string]struct{}),
	}
}

// Register adds a player connection
func (h *Hub) Register(playerID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[playerID] = conn
}

// Unregister removes a player connection
func (h *Hub) Unregister(playerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, playerID)
	// Remove from all tables
	for tableID, players := range h.tablePlayers {
		delete(players, playerID)
		if len(players) == 0 {
			delete(h.tablePlayers, tableID)
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

// SendToPlayer sends a message to a specific player
func (h *Hub) SendToPlayer(playerID string, msg Message) error {
	h.mu.RLock()
	conn := h.connections[playerID]
	h.mu.RUnlock()
	if conn == nil {
		return nil // player not connected
	}
	return conn.WriteJSON(msg)
}

// BroadcastToTable sends a message to all players at a table
func (h *Hub) BroadcastToTable(tableID string, msg Message) {
	h.mu.RLock()
	players := h.tablePlayers[tableID]
	conns := make(map[string]*websocket.Conn, len(players))
	for pid := range players {
		conns[pid] = h.connections[pid]
	}
	h.mu.RUnlock()

	for _, conn := range conns {
		if conn != nil {
			_ = conn.WriteJSON(msg)
		}
	}
}

// BroadcastToTableExcept sends to all except one player
func (h *Hub) BroadcastToTableExcept(tableID, excludePlayerID string, msg Message) {
	h.mu.RLock()
	players := h.tablePlayers[tableID]
	conns := make(map[string]*websocket.Conn, len(players))
	for pid := range players {
		if pid != excludePlayerID {
			conns[pid] = h.connections[pid]
		}
	}
	h.mu.RUnlock()

	for _, conn := range conns {
		if conn != nil {
			_ = conn.WriteJSON(msg)
		}
	}
}
