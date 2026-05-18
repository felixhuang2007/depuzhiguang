package table

import "fmt"

// TableConfig defines table parameters
type TableConfig struct {
	ID        string
	Name      string
	MaxSeats  int   // 2, 6, or 9
	SmallBlind int
	BigBlind   int
	MinBuyIn   int   // in bb
	MaxBuyIn   int   // in bb
	RakePercent float64 // e.g., 0.05 for 5%
	RakeCap     int     // max rake in bb
}

// Table manages players, seats, and configuration
type Table struct {
	Config  TableConfig
	Seats   map[int]*Player  // seat number -> player (nil if empty)
}

// NewTable creates a new table with given configuration
func NewTable(config TableConfig) *Table {
	seats := make(map[int]*Player, config.MaxSeats)
	for i := 0; i < config.MaxSeats; i++ {
		seats[i] = nil
	}
	return &Table{
		Config: config,
		Seats:  seats,
	}
}

// Join seats a player at the table
func (t *Table) Join(player *Player) error {
	if player == nil {
		return fmt.Errorf("player is nil")
	}
	if t.Seats[player.Seat] != nil {
		return fmt.Errorf("seat %d is occupied", player.Seat)
	}
	if player.Seat < 0 || player.Seat >= t.Config.MaxSeats {
		return fmt.Errorf("invalid seat %d", player.Seat)
	}
	if player.Stack < t.Config.MinBuyIn*t.Config.BigBlind {
		return fmt.Errorf("stack %d below min buy-in %d", player.Stack, t.Config.MinBuyIn*t.Config.BigBlind)
	}
	if player.Stack > t.Config.MaxBuyIn*t.Config.BigBlind {
		return fmt.Errorf("stack %d above max buy-in %d", player.Stack, t.Config.MaxBuyIn*t.Config.BigBlind)
	}
	t.Seats[player.Seat] = player
	return nil
}

// Leave removes a player from the table
func (t *Table) Leave(seat int) (*Player, error) {
	if seat < 0 || seat >= t.Config.MaxSeats {
		return nil, fmt.Errorf("invalid seat %d", seat)
	}
	p := t.Seats[seat]
	if p == nil {
		return nil, fmt.Errorf("seat %d is empty", seat)
	}
	t.Seats[seat] = nil
	return p, nil
}

// PlayerAt returns the player at a given seat
func (t *Table) PlayerAt(seat int) *Player {
	if seat < 0 || seat >= t.Config.MaxSeats {
		return nil
	}
	return t.Seats[seat]
}

// OccupiedSeats returns a list of seats that have players
func (t *Table) OccupiedSeats() []int {
	var seats []int
	for i := 0; i < t.Config.MaxSeats; i++ {
		if t.Seats[i] != nil {
			seats = append(seats, i)
		}
	}
	return seats
}

// PlayerCount returns number of seated players
func (t *Table) PlayerCount() int {
	return len(t.OccupiedSeats())
}

// NextAvailableSeat returns the next empty seat, or -1 if full
func (t *Table) NextAvailableSeat() int {
	for i := 0; i < t.Config.MaxSeats; i++ {
		if t.Seats[i] == nil {
			return i
		}
	}
	return -1
}
