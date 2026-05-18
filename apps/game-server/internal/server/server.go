package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

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
	server *http.Server
}

// NewServer creates a new game server
func NewServer(addr string) *Server {
	hub := NewHub()
	mux := http.NewServeMux()

	s := &Server{
		hub: hub,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ws", s.wsHandler)

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

		// TODO: Route message to game logic
		log.Printf("Received %s from %s", msg.Type, playerID)
	}

	log.Printf("Player %s disconnected", playerID)
}
