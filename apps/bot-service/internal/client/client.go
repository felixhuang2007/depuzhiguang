package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/depuzhiguang/bot-service/internal/ai"
	"github.com/gorilla/websocket"
)

// MessageType represents WebSocket message types from game server.
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

// Message is the envelope for WebSocket communications.
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ActionPayload sends a player action.
type ActionPayload struct {
	TableID  string `json:"table_id"`
	PlayerID string `json:"player_id"`
	Action   string `json:"action"`
	Amount   int    `json:"amount,omitempty"`
}

// StateSnapshotPayload represents full table state.
type StateSnapshotPayload struct {
	TableID     string        `json:"table_id"`
	State       string        `json:"state"`
	Community   []CardJSON    `json:"community"`
	Players     []PlayerState `json:"players"`
	Pot         int           `json:"pot"`
	CurrentTurn int           `json:"current_turn"`
	Button      int           `json:"button"`
}

// CardJSON represents a card as received from server.
type CardJSON struct {
	Suit int `json:"suit"`
	Rank int `json:"rank"`
}

// PlayerState represents a player's visible state.
type PlayerState struct {
	ID        string     `json:"id"`
	Seat      int        `json:"seat"`
	Stack     int        `json:"stack"`
	Status    string     `json:"status"`
	Bet       int        `json:"bet"`
	HoleCards []CardJSON `json:"hole_cards,omitempty"`
}

// GameClient connects a bot to the game server via WebSocket.
type GameClient struct {
	wsURL    string
	playerID string
	tableID  string
	engine   *ai.Engine
	conn     *websocket.Conn
	stopCh   chan struct{}
}

// NewGameClient creates a new game client for a bot.
func NewGameClient(wsURL, playerID, tableID string, engine *ai.Engine) *GameClient {
	return &GameClient{
		wsURL:    wsURL,
		playerID: playerID,
		tableID:  tableID,
		engine:   engine,
		stopCh:   make(chan struct{}),
	}
}

// Connect establishes the WebSocket connection and joins the table.
func (c *GameClient) Connect() error {
	u, err := url.Parse(c.wsURL)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("player_id", c.playerID)
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	c.conn = conn

	// Send join table message
	joinPayload, _ := json.Marshal(map[string]string{
		"table_id": c.tableID,
	})
	if err := c.conn.WriteJSON(Message{Type: MsgJoinTable, Payload: joinPayload}); err != nil {
		conn.Close()
		return fmt.Errorf("join failed: %w", err)
	}

	log.Printf("[%s] Connected and joined table %s", c.playerID, c.tableID)
	return nil
}

// Run starts the read loop, processing game messages.
func (c *GameClient) Run() {
	defer c.conn.Close()

	// Send periodic pings
	go c.pingLoop()

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[%s] WebSocket error: %v", c.playerID, err)
			}
			return
		}

		switch msg.Type {
		case MsgStateSnapshot:
			c.handleStateSnapshot(msg.Payload)
		case MsgHandResult:
			log.Printf("[%s] Hand result received", c.playerID)
		case MsgError:
			log.Printf("[%s] Error: %s", c.playerID, string(msg.Payload))
		case MsgPong:
			// ignore
		default:
			log.Printf("[%s] Received %s", c.playerID, msg.Type)
		}
	}
}

// Stop disconnects the client.
func (c *GameClient) Stop() {
	close(c.stopCh)
	if c.conn != nil {
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
	}
}

func (c *GameClient) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			if err := c.conn.WriteJSON(Message{Type: MsgPing}); err != nil {
				return
			}
		}
	}
}

func (c *GameClient) handleStateSnapshot(raw json.RawMessage) {
	var state StateSnapshotPayload
	if err := json.Unmarshal(raw, &state); err != nil {
		log.Printf("[%s] Failed to unmarshal state: %v", c.playerID, err)
		return
	}

	// Check if it's our turn
	var mySeat, myStack, toCall, currentBet, minRaise int
	var myHole []string
	var community []string
	for _, p := range state.Players {
		if p.ID == c.playerID {
			mySeat = p.Seat
			myStack = p.Stack
			currentBet = p.Bet
			for _, card := range p.HoleCards {
				myHole = append(myHole, cardToString(card))
			}
		}
	}
	if mySeat != state.CurrentTurn {
		return // not our turn
	}

	for _, card := range state.Community {
		community = append(community, cardToString(card))
	}

	// Calculate to-call
	maxBet := 0
	for _, p := range state.Players {
		if p.Bet > maxBet {
			maxBet = p.Bet
		}
	}
	toCall = maxBet - currentBet
	minRaise = 10 // TODO: get from table config

	decision := c.engine.Decide(myHole, community, state.Pot, toCall, myStack, minRaise)

	actionPayload, _ := json.Marshal(ActionPayload{
		TableID:  c.tableID,
		PlayerID: c.playerID,
		Action:   decision.Action,
		Amount:   decision.Amount,
	})

	log.Printf("[%s] Decision: %s %d (delay %v)", c.playerID, decision.Action, decision.Amount, decision.Delay)

	time.AfterFunc(decision.Delay, func() {
		if err := c.conn.WriteJSON(Message{Type: MsgAction, Payload: actionPayload}); err != nil {
			log.Printf("[%s] Failed to send action: %v", c.playerID, err)
		}
	})
}

func cardToString(c CardJSON) string {
	ranks := []string{"", "A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
	suits := []string{"C", "D", "H", "S"}
	if c.Rank >= 1 && c.Rank <= 14 && c.Suit >= 0 && c.Suit <= 3 {
		return ranks[c.Rank] + suits[c.Suit]
	}
	return "?"
}
