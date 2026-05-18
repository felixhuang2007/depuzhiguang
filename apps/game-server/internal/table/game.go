package table

import (
	"fmt"

	"github.com/depuzhiguang/game-server/internal/engine"
)

// GameState represents the current state of a poker hand
type GameState int

const (
	StateWaiting GameState = iota
	StateDealing
	StatePreflop
	StateFlop
	StateTurn
	StateRiver
	StateShowdown
	StateComplete
)

// ActionType represents a player action
type ActionType int

const (
	ActionFold ActionType = iota
	ActionCheck
	ActionCall
	ActionBet
	ActionRaise
	ActionAllIn
)

// Action represents a player decision
type Action struct {
	PlayerID string
	Type     ActionType
	Amount   int
}

// Game manages a single poker hand
type Game struct {
	Table         *Table
	State         GameState
	Deck          *engine.Deck
	Pot           *engine.Pot
	Button        int
	Community     []engine.Card
	CurrentTurn   int
	LastRaise     int
	LastAggressor int
	highestBet    int
	acted         map[string]bool
}

// NewGame creates a new game for the given table
func NewGame(table *Table) *Game {
	return &Game{
		Table:  table,
		State:  StateWaiting,
		Button: 0,
		acted:  make(map[string]bool),
	}
}

// Start begins a new hand
func (g *Game) Start() error {
	if g.Table.PlayerCount() < 2 {
		return fmt.Errorf("need at least 2 players")
	}

	g.State = StateDealing
	g.Deck = engine.NewDeck()
	g.Deck.Shuffle()
	g.Pot = engine.NewPot()
	g.Community = nil
	g.acted = make(map[string]bool)
	g.highestBet = 0
	g.LastRaise = g.Table.Config.BigBlind

	// Deal hole cards
	for _, seat := range g.Table.OccupiedSeats() {
		player := g.Table.Seats[seat]
		player.ResetForNewHand()
		c1, _ := g.Deck.Deal()
		c2, _ := g.Deck.Deal()
		player.Holes = [2]engine.Card{c1, c2}
		player.Status = Active
	}

	// Post blinds
	occupied := g.Table.OccupiedSeats()
	if len(occupied) < 2 {
		return fmt.Errorf("need at least 2 players")
	}

	// Button is at last occupied seat (so first player is SB)
	g.Button = occupied[len(occupied)-1]

	// Small blind: next player after button
	sbSeat := g.nextActiveSeat(g.Button)
	bbSeat := g.nextActiveSeat(sbSeat)

	sb := g.Table.Seats[sbSeat]
	bb := g.Table.Seats[bbSeat]

	sbBet := g.Table.Config.SmallBlind
	if sb.Stack < sbBet {
		sbBet = sb.Stack
	}
	sb.Stack -= sbBet
	sb.Bet = sbBet
	sb.TotalBet = sbBet
	g.Pot.AddBet(sb.ID, sbBet)

	bbBet := g.Table.Config.BigBlind
	if bb.Stack < bbBet {
		bbBet = bb.Stack
	}
	bb.Stack -= bbBet
	bb.Bet = bbBet
	bb.TotalBet = bbBet
	g.Pot.AddBet(bb.ID, bbBet)

	g.highestBet = bbBet

	// First to act: next after big blind (UTG)
	g.CurrentTurn = g.nextActiveSeat(bbSeat)
	g.LastAggressor = bbSeat

	g.State = StatePreflop
	return nil
}

// ProcessAction handles a player action
func (g *Game) ProcessAction(action Action) error {
	player := g.Table.Seats[g.CurrentTurn]
	if player == nil || player.ID != action.PlayerID {
		return fmt.Errorf("not player's turn")
	}
	if !player.CanAct() {
		return fmt.Errorf("player cannot act")
	}

	switch action.Type {
	case ActionFold:
		player.Status = Folded
		g.Pot.AddBet(player.ID, 0)

	case ActionCheck:
		if player.Bet < g.highestBet {
			return fmt.Errorf("cannot check, must call %d", g.highestBet-player.Bet)
		}

	case ActionCall:
		callAmount := g.highestBet - player.Bet
		if callAmount <= 0 {
			return fmt.Errorf("nothing to call")
		}
		if callAmount > player.Stack {
			callAmount = player.Stack
		}
		player.Stack -= callAmount
		player.Bet += callAmount
		player.TotalBet += callAmount
		g.Pot.AddBet(player.ID, player.Bet)

	case ActionBet, ActionRaise:
		if action.Amount <= 0 {
			return fmt.Errorf("bet/raise amount must be positive")
		}

		// Calculate minimum allowed bet/raise
		minAmount := g.highestBet
		if g.highestBet == 0 {
			// No previous bet this round — minimum bet is 1 big blind
			minAmount = g.Table.Config.BigBlind
		} else {
			// Raise must be at least last raise size more than current highest bet
			minAmount = g.highestBet + g.LastRaise
		}

		if action.Amount < minAmount {
			return fmt.Errorf("bet/raise must be at least %d", minAmount)
		}
		if action.Amount > player.Stack+player.Bet {
			return fmt.Errorf("insufficient stack for bet/raise")
		}

		// If betting all their stack, it's an all-in
		if action.Amount >= player.Stack+player.Bet {
			allInAmount := player.Stack
			player.Stack = 0
			player.Bet += allInAmount
			player.TotalBet += allInAmount
			g.Pot.AddBet(player.ID, player.Bet)
			if player.Bet > g.highestBet {
				raiseSize := player.Bet - g.highestBet
				g.LastRaise = raiseSize
				g.highestBet = player.Bet
				g.LastAggressor = g.CurrentTurn
			}
			player.Status = AllIn
		} else {
			additional := action.Amount - player.Bet
			player.Stack -= additional
			player.Bet = action.Amount
			player.TotalBet += additional
			g.Pot.AddBet(player.ID, player.Bet)
			raiseSize := action.Amount - g.highestBet
			g.LastRaise = raiseSize
			g.highestBet = action.Amount
			g.LastAggressor = g.CurrentTurn
		}

	case ActionAllIn:
		allInAmount := player.Stack
		player.Stack = 0
		player.Bet += allInAmount
		player.TotalBet += allInAmount
		g.Pot.AddBet(player.ID, player.Bet)
		if player.Bet > g.highestBet {
			g.highestBet = player.Bet
			g.LastAggressor = g.CurrentTurn
		}
		player.Status = AllIn
	}

	g.acted[player.ID] = true

	// Check if hand is over (all but one folded)
	activeInHand := g.activePlayerCount()
	if activeInHand <= 1 {
		g.advanceState()
		return nil
	}

	// Move to next player
	g.CurrentTurn = g.nextActivePlayer(g.CurrentTurn)

	// Check if betting round is complete
	if g.isBettingRoundComplete() {
		g.advanceState()
	}

	return nil
}

// advanceState moves to the next betting round or ends the hand
func (g *Game) advanceState() {
	// Close current betting round
	g.Pot.CloseBettingRound()

	// Reset bets for next round
	for _, seat := range g.Table.OccupiedSeats() {
		p := g.Table.Seats[seat]
		p.Bet = 0
	}
	g.highestBet = 0
	g.acted = make(map[string]bool)
	g.LastRaise = g.Table.Config.BigBlind

	switch g.State {
	case StatePreflop:
		g.dealCommunity(3)
		g.State = StateFlop
	case StateFlop:
		g.dealCommunity(1)
		g.State = StateTurn
	case StateTurn:
		g.dealCommunity(1)
		g.State = StateRiver
	case StateRiver:
		g.State = StateShowdown
	case StateShowdown, StateComplete:
		g.State = StateComplete
		return
	}

	// If only one player left in hand, hand is complete
	if g.activePlayerCount() <= 1 {
		g.State = StateComplete
		return
	}

	// Set first actor for next round (first active player after button)
	if g.State != StateComplete {
		g.CurrentTurn = g.nextActiveSeat(g.Button)
	}
}

// dealCommunity deals community cards
func (g *Game) dealCommunity(count int) {
	for i := 0; i < count; i++ {
		c, ok := g.Deck.Deal()
		if ok {
			g.Community = append(g.Community, c)
		}
	}
}

// isBettingRoundComplete checks if all active players have matched the highest bet
func (g *Game) isBettingRoundComplete() bool {
	activePlayers := 0
	allActed := true
	allMatched := true

	for _, seat := range g.Table.OccupiedSeats() {
		p := g.Table.Seats[seat]
		if !p.IsInHand() {
			continue
		}
		activePlayers++

		if !g.acted[p.ID] {
			allActed = false
		}

		// All-in players are always "matched" for round completion
		if p.Status != AllIn && p.Bet < g.highestBet {
			allMatched = false
		}
	}

	if activePlayers <= 1 {
		return true
	}

	return allActed && allMatched
}

// nextActiveSeat returns the next occupied seat
func (g *Game) nextActiveSeat(from int) int {
	seat := (from + 1) % g.Table.Config.MaxSeats
	for i := 0; i < g.Table.Config.MaxSeats; i++ {
		if g.Table.Seats[seat] != nil {
			return seat
		}
		seat = (seat + 1) % g.Table.Config.MaxSeats
	}
	return -1
}

// nextActivePlayer returns the next player who can act (not folded, not all-in)
func (g *Game) nextActivePlayer(from int) int {
	seat := (from + 1) % g.Table.Config.MaxSeats
	for i := 0; i < g.Table.Config.MaxSeats; i++ {
		p := g.Table.Seats[seat]
		if p != nil && p.CanAct() {
			return seat
		}
		seat = (seat + 1) % g.Table.Config.MaxSeats
	}
	return -1
}

// activePlayerCount returns number of players still in hand
func (g *Game) activePlayerCount() int {
	count := 0
	for _, seat := range g.Table.OccupiedSeats() {
		if g.Table.Seats[seat].IsInHand() {
			count++
		}
	}
	return count
}
