package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
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
	MsgYourTurn      MessageType = "your_turn"
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
	State       int           `json:"state"`
	Community   []CardJSON    `json:"community"`
	Players     []PlayerState `json:"players"`
	Pot         int           `json:"pot"`
	CurrentTurn int           `json:"current_turn"`
	Button      int           `json:"button"`
	MinRaise    int           `json:"min_raise"`
	BigBlind    int           `json:"big_blind"`
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
	Status    int        `json:"status"`
	Bet       int        `json:"bet"`
	HoleCards []CardJSON `json:"hole_cards,omitempty"`
}

// GameClient connects a bot to the game server via WebSocket.
type GameClient struct {
	wsURL       string
	playerID    string
	tableID     string
	engine      *ai.Engine
	conn        *websocket.Conn
	stopCh      chan struct{}
	actionMu    sync.RWMutex
	onAction    func(phase, action string, amount, pot, stack int)
	handsPlayed int
	maxHands    int

	// Reconnection state
	mu           sync.RWMutex
	connected    bool
	reconnecting bool
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
	conn, err := c.dial()
	if err != nil {
		return err
	}
	c.conn = conn
	c.setConnected(true)

	// Send join table message
	joinPayload, _ := json.Marshal(map[string]string{
		"table_id": c.tableID,
	})
	if err := c.conn.WriteJSON(Message{Type: MsgJoinTable, Payload: joinPayload}); err != nil {
		conn.Close()
		c.setConnected(false)
		return fmt.Errorf("join failed: %w", err)
	}

	log.Printf("[%s] Connected and joined table %s", c.playerID, c.tableID)
	return nil
}

func (c *GameClient) dial() (*websocket.Conn, error) {
	u, err := url.Parse(c.wsURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("player_id", c.playerID)
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}
	return conn, nil
}

// Run starts the read loop, processing game messages.
func (c *GameClient) Run() {
	defer func() {
		if c.conn != nil {
			c.conn.Close()
		}
	}()

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
			c.setConnected(false)

			// Attempt reconnection
			if c.shouldReconnect() {
				if err := c.reconnect(); err != nil {
					log.Printf("[%s] Reconnect failed: %v", c.playerID, err)
					return
				}
				continue
			}
			return
		}

		switch msg.Type {
		case MsgStateSnapshot:
			c.handleStateSnapshot(msg.Payload)
		case MsgStateDelta:
			// TODO: handle delta updates for efficiency
			log.Printf("[%s] Received state_delta (ignored)", c.playerID)
		case MsgHandResult:
			c.handsPlayed++
			log.Printf("[%s] Hand result received (hands: %d/%d)", c.playerID, c.handsPlayed, c.maxHands)
		case MsgYourTurn:
			c.handleYourTurn(msg.Payload)
		case MsgError:
			log.Printf("[%s] Error: %s", c.playerID, string(msg.Payload))
		case MsgPong:
			// ignore
		default:
			log.Printf("[%s] Received %s", c.playerID, msg.Type)
		}
	}
}

func (c *GameClient) reconnect() error {
	c.mu.Lock()
	if c.reconnecting {
		c.mu.Unlock()
		return fmt.Errorf("already reconnecting")
	}
	c.reconnecting = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.reconnecting = false
		c.mu.Unlock()
	}()

	// Wait a bit before reconnecting
	time.Sleep(2 * time.Second)

	log.Printf("[%s] Attempting to reconnect...", c.playerID)

	conn, err := c.dial()
	if err != nil {
		return err
	}

	c.conn = conn
	c.setConnected(true)

	// Re-join table after reconnection
	joinPayload, _ := json.Marshal(map[string]string{
		"table_id": c.tableID,
	})
	if err := c.conn.WriteJSON(Message{Type: MsgJoinTable, Payload: joinPayload}); err != nil {
		conn.Close()
		c.setConnected(false)
		return fmt.Errorf("re-join failed: %w", err)
	}

	log.Printf("[%s] Reconnected and rejoined table %s", c.playerID, c.tableID)
	return nil
}

func (c *GameClient) shouldReconnect() bool {
	select {
	case <-c.stopCh:
		return false
	default:
		return true
	}
}

func (c *GameClient) setConnected(v bool) {
	c.mu.Lock()
	c.connected = v
	c.mu.Unlock()
}

// SetActionCallback registers a callback that will be invoked after each AI decision.
func (c *GameClient) SetActionCallback(cb func(phase, action string, amount, pot, stack int)) {
	c.actionMu.Lock()
	defer c.actionMu.Unlock()
	c.onAction = cb
}

// Stop disconnects the client.
func (c *GameClient) Stop() {
	close(c.stopCh)
	if c.conn != nil {
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
	}
}

// SetMaxHands sets the maximum number of hands this bot will play.
func (c *GameClient) SetMaxHands(n int) {
	c.maxHands = n
}

// HandsPlayed returns the number of hands played so far.
func (c *GameClient) HandsPlayed() int {
	return c.handsPlayed
}

// MaxHands returns the maximum number of hands.
func (c *GameClient) MaxHands() int {
	return c.maxHands
}

// Leave sends a leave_table message and disconnects.
func (c *GameClient) Leave() error {
	payload, _ := json.Marshal(map[string]string{
		"table_id":  c.tableID,
		"player_id": c.playerID,
	})
	if err := c.conn.WriteJSON(Message{Type: MsgLeaveTable, Payload: payload}); err != nil {
		return err
	}
	c.Stop()
	return nil
}

func (c *GameClient) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.mu.RLock()
			connected := c.connected
			c.mu.RUnlock()
			if connected && c.conn != nil {
				if err := c.conn.WriteJSON(Message{Type: MsgPing}); err != nil {
					log.Printf("[%s] Ping failed: %v", c.playerID, err)
				}
			}
		}
	}
}

func (c *GameClient) handleYourTurn(raw json.RawMessage) {
	// If server sends explicit your_turn message, handle it
	var payload struct {
		TableID string `json:"table_id"`
		Seat    int    `json:"seat"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}
	log.Printf("[%s] It's my turn! (seat %d)", c.playerID, payload.Seat)
}

func (c *GameClient) handleStateSnapshot(raw json.RawMessage) {
	var state StateSnapshotPayload
	if err := json.Unmarshal(raw, &state); err != nil {
		log.Printf("[%s] Failed to unmarshal state: %v", c.playerID, err)
		return
	}

	// Check if it's our turn and we can act
	var mySeat, myStack, toCall, currentBet, minRaise int
	var myHole []string
	var community []string
	for _, p := range state.Players {
		if p.ID == c.playerID {
			mySeat = p.Seat
			myStack = p.Stack
			currentBet = p.Bet
			// Only act if status is Active (1)
			if p.Status != 1 {
				return
			}
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
	minRaise = state.MinRaise
	if minRaise <= 0 {
		minRaise = state.BigBlind
	}
	if minRaise <= 0 {
		minRaise = 10
	}

	decision := c.engine.Decide(myHole, community, state.Pot, toCall, myStack, minRaise)

	if c.onAction != nil {
		c.actionMu.RLock()
		cb := c.onAction
		c.actionMu.RUnlock()
		if cb != nil {
			cb(getPhase(state.State), decision.Action, decision.Amount, state.Pot, myStack)
		}
	}

	actionPayload, _ := json.Marshal(ActionPayload{
		TableID:  c.tableID,
		PlayerID: c.playerID,
		Action:   decision.Action,
		Amount:   decision.Amount,
	})

	log.Printf("[%s] Decision: %s %d (delay %v)", c.playerID, decision.Action, decision.Amount, decision.Delay)

	time.AfterFunc(decision.Delay, func() {
		c.mu.RLock()
		connected := c.connected
		c.mu.RUnlock()
		if !connected {
			log.Printf("[%s] Cannot send action: not connected", c.playerID)
			return
		}
		if err := c.conn.WriteJSON(Message{Type: MsgAction, Payload: actionPayload}); err != nil {
			log.Printf("[%s] Failed to send action: %v", c.playerID, err)
		}
	})
}

func getPhase(state int) string {
	switch state {
	case 2:
		return "preflop"
	case 3:
		return "flop"
	case 4:
		return "turn"
	case 5:
		return "river"
	default:
		return "unknown"
	}
}

func cardToString(c CardJSON) string {
	ranks := []string{"", "A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
	suits := []string{"C", "D", "H", "S"}
	if c.Rank >= 1 && c.Rank <= 14 && c.Suit >= 0 && c.Suit <= 3 {
		return ranks[c.Rank] + suits[c.Suit]
	}
	return "?"
}
