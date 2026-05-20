package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/depuzhiguang/game-server/internal/table"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Server wraps the HTTP server and WebSocket hub
type Server struct {
	hub    *Hub
	tm     *TableManager
	server *http.Server
}

// NewServer creates a new game server
func NewServer(addr string) *Server {
	hub := NewHub()
	tm := NewTableManager(hub)
	mux := http.NewServeMux()

	s := &Server{
		hub: hub,
		tm:  tm,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ws", s.wsHandler)

	// Create default test table
	_, _ = tm.CreateTable(table.TableConfig{
		ID:          "default-6max",
		Name:        "Default 6-Max",
		MaxSeats:    6,
		SmallBlind:  5,
		BigBlind:    10,
		MinBuyIn:    50,
		MaxBuyIn:    200,
		RakePercent: 0.05,
		RakeCap:     3,
	})

	// Create simulation tables (match bot-service scheduler)
	simTableNames := []string{"sim-table-0", "sim-table-1", "sim-table-2"}
	for _, name := range simTableNames {
		_, _ = tm.CreateTable(table.TableConfig{
			ID:          name,
			Name:        name,
			MaxSeats:    7,
			SmallBlind:  5,
			BigBlind:    10,
			MinBuyIn:    50,
			MaxBuyIn:    200,
			RakePercent: 0.05,
			RakeCap:     3,
		})
	}

	return s
}

// Start begins listening for connections
func (s *Server) Start() error {
	log.Printf("Server starting on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// TODO: Authenticate player from query params
	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		playerID = fmt.Sprintf("anon_%d", time.Now().UnixNano())
	}

	s.hub.Register(playerID, conn)
	defer s.hub.Unregister(playerID)

	// Track joined tables for cleanup on disconnect
	joinedTables := make(map[string]struct{})
	defer func() {
		for tableID := range joinedTables {
			_ = s.tm.HandleLeave(LeaveTablePayload{
				TableID:  tableID,
				PlayerID: playerID,
			})
		}
	}()

	log.Printf("Player %s connected", playerID)

	// Send welcome message
	_ = conn.WriteJSON(Message{Type: MsgPong, Payload: map[string]string{"message": "connected"}})

	// Read loop
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for %s: %v", playerID, err)
			}
			break
		}

		// Handle ping
		if msg.Type == MsgPing {
			_ = conn.WriteJSON(Message{Type: MsgPong})
			continue
		}

		// Route message to game logic
		switch msg.Type {
		case MsgJoinTable:
			payload, ok := msg.Payload.(map[string]interface{})
			if !ok {
				_ = conn.WriteJSON(Message{Type: MsgError, Payload: ErrorPayload{Code: "bad_request", Message: "invalid join payload"}})
				continue
			}
			join := JoinTablePayload{
				TableID:  getString(payload, "table_id"),
				PlayerID: playerID,
				Token:    getString(payload, "token"),
			}
			if err := s.tm.HandleJoin(join); err != nil {
				_ = conn.WriteJSON(Message{Type: MsgError, Payload: ErrorPayload{Code: "join_failed", Message: err.Error()}})
			} else {
				joinedTables[join.TableID] = struct{}{}
			}

		case MsgLeaveTable:
			payload, ok := msg.Payload.(map[string]interface{})
			if !ok {
				continue
			}
			leave := LeaveTablePayload{
				TableID:  getString(payload, "table_id"),
				PlayerID: playerID,
			}
			if err := s.tm.HandleLeave(leave); err == nil {
				delete(joinedTables, leave.TableID)
			}

		case MsgAction:
			payload, ok := msg.Payload.(map[string]interface{})
			if !ok {
				_ = conn.WriteJSON(Message{Type: MsgError, Payload: ErrorPayload{Code: "bad_request", Message: "invalid action payload"}})
				continue
			}
			action := ActionPayload{
				TableID:  getString(payload, "table_id"),
				PlayerID: playerID,
				Amount:   getInt(payload, "amount"),
			}
			// Parse action type from string
			action.Action = parseActionType(getString(payload, "action"))
			log.Printf("Action from %s: %s %d", playerID, getString(payload, "action"), action.Amount)
			if err := s.tm.HandleAction(action); err != nil {
				log.Printf("Action failed for %s: %v", playerID, err)
				_ = conn.WriteJSON(Message{Type: MsgError, Payload: ErrorPayload{Code: "action_failed", Message: err.Error()}})
			}

		default:
			log.Printf("Received %s from %s", msg.Type, playerID)
		}
	}

	log.Printf("Player %s disconnected", playerID)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case int64:
			return int(n)
		}
	}
	return 0
}

func parseActionType(s string) table.ActionType {
	switch s {
	case "fold":
		return table.ActionFold
	case "check":
		return table.ActionCheck
	case "call":
		return table.ActionCall
	case "bet":
		return table.ActionBet
	case "raise":
		return table.ActionRaise
	case "all_in":
		return table.ActionAllIn
	default:
		return table.ActionFold
	}
}
