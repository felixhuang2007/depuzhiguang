package table

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTable(t *testing.T) {
	config := TableConfig{MaxSeats: 6, SmallBlind: 10, BigBlind: 20, MinBuyIn: 20, MaxBuyIn: 200}
	table := NewTable(config)
	assert.Equal(t, 6, table.Config.MaxSeats)
	assert.Equal(t, 0, table.PlayerCount())
}

func TestTable_Join(t *testing.T) {
	config := TableConfig{MaxSeats: 6, BigBlind: 20, MinBuyIn: 20, MaxBuyIn: 200}
	table := NewTable(config)

	p := NewPlayer("p1", 0, 1000)
	err := table.Join(p)
	assert.NoError(t, err)
	assert.Equal(t, 1, table.PlayerCount())
	assert.Equal(t, p, table.PlayerAt(0))
}

func TestTable_Join_OccupiedSeat(t *testing.T) {
	config := TableConfig{MaxSeats: 6, BigBlind: 20, MinBuyIn: 20, MaxBuyIn: 200}
	table := NewTable(config)

	p1 := NewPlayer("p1", 0, 1000)
	p2 := NewPlayer("p2", 0, 1000)
	table.Join(p1)
	err := table.Join(p2)
	assert.Error(t, err)
}

func TestTable_Join_BelowMinBuyIn(t *testing.T) {
	config := TableConfig{MaxSeats: 6, BigBlind: 20, MinBuyIn: 20, MaxBuyIn: 200}
	table := NewTable(config)

	p := NewPlayer("p1", 0, 100) // below 400 min
	err := table.Join(p)
	assert.Error(t, err)
}

func TestTable_Leave(t *testing.T) {
	config := TableConfig{MaxSeats: 6, BigBlind: 20, MinBuyIn: 20, MaxBuyIn: 200}
	table := NewTable(config)

	p := NewPlayer("p1", 2, 1000)
	table.Join(p)

	left, err := table.Leave(2)
	assert.NoError(t, err)
	assert.Equal(t, p, left)
	assert.Nil(t, table.PlayerAt(2))
	assert.Equal(t, 0, table.PlayerCount())
}

func TestTable_OccupiedSeats(t *testing.T) {
	config := TableConfig{MaxSeats: 9, BigBlind: 20, MinBuyIn: 20, MaxBuyIn: 200}
	table := NewTable(config)

	table.Join(NewPlayer("p1", 0, 1000))
	table.Join(NewPlayer("p2", 3, 1000))
	table.Join(NewPlayer("p3", 7, 1000))

	seats := table.OccupiedSeats()
	assert.Equal(t, []int{0, 3, 7}, seats)
}

func TestTable_NextAvailableSeat(t *testing.T) {
	config := TableConfig{MaxSeats: 6, BigBlind: 20, MinBuyIn: 20, MaxBuyIn: 200}
	table := NewTable(config)

	table.Join(NewPlayer("p1", 0, 1000))
	table.Join(NewPlayer("p2", 1, 1000))

	assert.Equal(t, 2, table.NextAvailableSeat())
}
