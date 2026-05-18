package server

import (
	"github.com/depuzhiguang/game-server/internal/engine"
	"github.com/depuzhiguang/game-server/internal/table"
)

// MessageType identifies the kind of WebSocket message
type MessageType string

const (
	MsgJoinTable     MessageType = "join_table"
	MsgLeaveTable    MessageType = "leave_table"
	MsgAction        MessageType = "action"
	MsgStateSnapshot MessageType = "state_snapshot"
	MsgStateDelta    MessageType = "state_delta"
	MsgPlayerJoined  MessageType = "player_joined"
	MsgPlayerLeft    MessageType = "player_left"
	MsgHandResult    MessageType = "hand_result"
	MsgError         MessageType = "error"
	MsgPing          MessageType = "ping"
	MsgPong          MessageType = "pong"
)

// Message is the envelope for all WebSocket communications
type Message struct {
	Type MessageType `json:"type"`
	// Payload is type-specific data
	Payload interface{} `json:"payload,omitempty"`
}

// JoinTablePayload requests to join a table
type JoinTablePayload struct {
	TableID  string `json:"table_id"`
	PlayerID string `json:"player_id"`
	Token    string `json:"token"`
}

// LeaveTablePayload requests to leave a table
type LeaveTablePayload struct {
	TableID  string `json:"table_id"`
	PlayerID string `json:"player_id"`
}

// ActionPayload sends a player action
type ActionPayload struct {
	TableID  string           `json:"table_id"`
	PlayerID string           `json:"player_id"`
	Action   table.ActionType `json:"action"`
	Amount   int              `json:"amount,omitempty"`
}

// StateSnapshotPayload sends full table state
type StateSnapshotPayload struct {
	TableID     string              `json:"table_id"`
	State       table.GameState     `json:"state"`
	Community   []engine.Card       `json:"community"`
	Players     []PlayerState       `json:"players"`
	Pot         int                 `json:"pot"`
	CurrentTurn int                 `json:"current_turn"`
	Button      int                 `json:"button"`
}

// PlayerState represents a player's visible state
type PlayerState struct {
	ID       string           `json:"id"`
	Seat     int              `json:"seat"`
	Stack    int              `json:"stack"`
	Status   table.PlayerStatus `json:"status"`
	Bet      int              `json:"bet"`
	HoleCards []engine.Card   `json:"hole_cards,omitempty"` // only for self
}

// StateDeltaPayload sends incremental state update
type StateDeltaPayload struct {
	TableID string                 `json:"table_id"`
	Changes map[string]interface{} `json:"changes"`
}

// PlayerEventPayload notifies player join/leave
type PlayerEventPayload struct {
	TableID  string `json:"table_id"`
	PlayerID string `json:"player_id"`
	Seat     int    `json:"seat"`
}

// HandResultPayload sends hand outcome
type HandResultPayload struct {
	TableID   string                  `json:"table_id"`
	Winners   []string                `json:"winners"`
	Community []engine.Card           `json:"community"`
	Pot       int                     `json:"pot"`
}

// ErrorPayload sends an error message
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
