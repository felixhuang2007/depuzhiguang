package server

import (
	"context"
	"fmt"
	"log/slog"
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
	lobby  *LobbyManager
	server *http.Server
	logg   *slog.Logger
}

// NewServer creates a new game server
func NewServer(addr string, apiBaseURL string, logg *slog.Logger) *Server {
	hub := NewHub(logg)
	tm := NewTableManager(hub, logg)
	lobby := NewLobbyManager(tm, logg)
	tm.SetLobby(lobby)
	mux := http.NewServeMux()

	s := &Server{
		hub:   hub,
		tm:    tm,
		lobby: lobby,
		logg:  logg,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ws", s.wsHandler)
	mux.HandleFunc("/lobby/tables", s.lobby.TablesHandler)
	mux.HandleFunc("/ws/lobby", s.lobby.WSHandler)

	// Create default test table
	if _, err := tm.CreateTable(table.TableConfig{
		ID:          "default-6max",
		Name:        "Default 6-Max",
		MaxSeats:    6,
		SmallBlind:  5,
		BigBlind:    10,
		MinBuyIn:    50,
		MaxBuyIn:    200,
		RakePercent: 0.05,
		RakeCap:     3,
	}); err != nil {
		s.logg.Error("failed to create table", "error", err)
	}

	// Create t1 table (matches Flutter app mock lobby)
	if _, err := tm.CreateTable(table.TableConfig{
		ID:          "t1",
		Name:        "经典六人桌",
		MaxSeats:    6,
		SmallBlind:  5,
		BigBlind:    10,
		MinBuyIn:    500,
		MaxBuyIn:    5000,
		RakePercent: 0.05,
		RakeCap:     3,
	}); err != nil {
		s.logg.Error("failed to create table", "error", err)
	}

	// Create simulation tables (match bot-service scheduler)
	simTableNames := []string{"sim-table-0", "sim-table-1", "sim-table-2"}
	for _, name := range simTableNames {
		if _, err := tm.CreateTable(table.TableConfig{
			ID:          name,
			Name:        name,
			MaxSeats:    7,
			SmallBlind:  5,
			BigBlind:    10,
			MinBuyIn:    50,
			MaxBuyIn:    200,
			RakePercent: 0.05,
			RakeCap:     3,
		}); err != nil {
			s.logg.Error("failed to create table", "error", err)
		}
	}

	return s
}

// Start begins listening for connections
func (s *Server) Start() error {
	s.logg.Info("Server starting", slog.String("addr", s.server.Addr))
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
		s.logg.Error("WebSocket upgrade failed", slog.String("error", err.Error()))
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

	s.logg.Info("Player connected", slog.String("player_id", playerID))

	// Send welcome message
	_ = conn.WriteJSON(Message{Type: MsgPong, Payload: map[string]string{"message": "connected"}})

	// Read loop
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logg.Error("WebSocket error", slog.String("player_id", playerID), slog.String("error", err.Error()))
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
			s.logg.Info("Action received", slog.String("player_id", playerID), slog.String("action", getString(payload, "action")), slog.Int("amount", action.Amount))
			if err := s.tm.HandleAction(action); err != nil {
				s.logg.Error("Action failed", slog.String("player_id", playerID), slog.String("error", err.Error()))
				_ = conn.WriteJSON(Message{Type: MsgError, Payload: ErrorPayload{Code: "action_failed", Message: err.Error()}})
			}

		default:
			s.logg.Info("Received message", slog.String("type", string(msg.Type)), slog.String("player_id", playerID))
		}
	}

	s.logg.Info("Player disconnected", slog.String("player_id", playerID))
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
