package table

import (
	"github.com/depuzhiguang/game-server/internal/engine"
)

// PlayerStatus represents a player's current state at the table
type PlayerStatus int

const (
	Waiting PlayerStatus = iota
	Active
	Folded
	AllIn
	SittingOut
)

func (s PlayerStatus) String() string {
	switch s {
	case Waiting:
		return "waiting"
	case Active:
		return "active"
	case Folded:
		return "folded"
	case AllIn:
		return "all-in"
	case SittingOut:
		return "sitting-out"
	default:
		return "unknown"
	}
}

// Player represents a player seated at a table
type Player struct {
	ID       string
	Seat     int
	Stack    int
	Holes    [2]engine.Card
	Status   PlayerStatus
	Bet      int  // amount bet in current round
	TotalBet int  // total amount bet across all rounds this hand
}

// NewPlayer creates a new player with initial stack
func NewPlayer(id string, seat int, stack int) *Player {
	return &Player{
		ID:     id,
		Seat:   seat,
		Stack:  stack,
		Status: Waiting,
	}
}

// IsActive returns true if player can still act this hand
func (p *Player) IsActive() bool {
	return p.Status == Active || p.Status == Waiting
}

// IsInHand returns true if player hasn't folded
func (p *Player) IsInHand() bool {
	return p.Status == Active || p.Status == AllIn
}

// ResetForNewHand clears hand-specific state
func (p *Player) ResetForNewHand() {
	p.Holes = [2]engine.Card{}
	if p.Status == Folded || p.Status == AllIn {
		p.Status = Waiting
	}
	p.Bet = 0
	p.TotalBet = 0
}

// CanAct returns true if player can make a decision right now
func (p *Player) CanAct() bool {
	return p.Status == Active
}
