package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// LobbyManager manages lobby HTTP API and WebSocket connections for table list updates.
type LobbyManager struct {
	tm     *TableManager
	logg   *slog.Logger
	conns  map[string]*websocket.Conn
	writeMu map[string]*sync.Mutex
	mu     sync.RWMutex
}

// TableInfo represents a table's public metadata for the lobby.
type TableInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	MaxSeats   int    `json:"max_seats"`
	Occupied   int    `json:"occupied"`
	SmallBlind int    `json:"small_blind"`
	BigBlind   int    `json:"big_blind"`
	Status     string `json:"status"`
}

// NewLobbyManager creates a new lobby manager.
func NewLobbyManager(tm *TableManager, logg *slog.Logger) *LobbyManager {
	return &LobbyManager{
		tm:      tm,
		logg:    logg,
		conns:   make(map[string]*websocket.Conn),
		writeMu: make(map[string]*sync.Mutex),
	}
}

// TablesHandler returns the list of tables as JSON.
func (lm *LobbyManager) TablesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tables := lm.tm.GetTableList()
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"tables": tables}); err != nil {
		lm.logg.Error("failed to encode tables", slog.String("error", err.Error()))
	}
}

// WSHandler handles WebSocket connections for lobby updates.
func (lm *LobbyManager) WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		lm.logg.Error("lobby ws upgrade failed", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		playerID = "anon_" + r.RemoteAddr
	}

	lm.mu.Lock()
	if oldConn, exists := lm.conns[playerID]; exists {
		oldConn.Close()
	}
	lm.conns[playerID] = conn
	lm.writeMu[playerID] = &sync.Mutex{}
	lm.mu.Unlock()

	// Send initial snapshot
	lm.sendTablesSnapshot(conn)

	// Listen for messages (keep connection alive)
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		if msg.Type == MsgJoinTable {
			// Client requests to join a specific table
			// Just acknowledge; actual join happens via /ws
			_ = conn.WriteJSON(Message{Type: MsgPong})
		}
	}

	lm.mu.Lock()
	delete(lm.conns, playerID)
	delete(lm.writeMu, playerID)
	lm.mu.Unlock()
}

// BroadcastTablesUpdate sends current table list to all lobby connections.
func (lm *LobbyManager) BroadcastTablesUpdate() {
	lm.mu.RLock()
	conns := make(map[string]*websocket.Conn, len(lm.conns))
	muMap := make(map[string]*sync.Mutex, len(lm.conns))
	for pid, c := range lm.conns {
		conns[pid] = c
		muMap[pid] = lm.writeMu[pid]
	}
	lm.mu.RUnlock()

	if len(conns) == 0 {
		return
	}

	tables := lm.tm.GetTableList()
	payload := map[string]interface{}{"tables": tables}
	msg := Message{Type: "tables_update", Payload: payload}

	for pid, c := range conns {
		if mu := muMap[pid]; mu != nil {
			mu.Lock()
			if err := c.WriteJSON(msg); err != nil {
				lm.logg.Error("lobby broadcast failed", slog.String("player_id", pid), slog.String("error", err.Error()))
			}
			mu.Unlock()
		}
	}
}

func (lm *LobbyManager) sendTablesSnapshot(conn *websocket.Conn) {
	tables := lm.tm.GetTableList()
	payload := map[string]interface{}{"tables": tables}
	if err := conn.WriteJSON(Message{Type: "tables_update", Payload: payload}); err != nil {
		lm.logg.Error("lobby snapshot failed", slog.String("error", err.Error()))
	}
}
