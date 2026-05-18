package table

import (
	"testing"

	"github.com/depuzhiguang/game-server/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGame_Start(t *testing.T) {
	table := NewTable(TableConfig{
		ID:         "test-table",
		Name:       "Test Table",
		MaxSeats:   6,
		SmallBlind: 5,
		BigBlind:   10,
		MinBuyIn:   20,
		MaxBuyIn:   100,
	})

	p0 := NewPlayer("alice", 0, 1000)
	p1 := NewPlayer("bob", 1, 1000)
	p2 := NewPlayer("charlie", 2, 1000)

	require.NoError(t, table.Join(p0))
	require.NoError(t, table.Join(p1))
	require.NoError(t, table.Join(p2))

	game := NewGame(table)
	require.NotNil(t, game)
	assert.Equal(t, StateWaiting, game.State)

	err := game.Start()
	require.NoError(t, err)

	// Verify state advanced to Preflop
	assert.Equal(t, StatePreflop, game.State)

	// Verify deck was created and shuffled
	assert.NotNil(t, game.Deck)
	assert.Equal(t, 52-6, game.Deck.Len()) // 6 hole cards dealt

	// Verify hole cards dealt to each player
	assert.NotEqual(t, engine.Card{}, table.Seats[0].Holes[0])
	assert.NotEqual(t, engine.Card{}, table.Seats[0].Holes[1])
	assert.NotEqual(t, engine.Card{}, table.Seats[1].Holes[0])
	assert.NotEqual(t, engine.Card{}, table.Seats[1].Holes[1])
	assert.NotEqual(t, engine.Card{}, table.Seats[2].Holes[0])
	assert.NotEqual(t, engine.Card{}, table.Seats[2].Holes[1])

	// Verify blinds posted
	assert.Equal(t, 5, table.Seats[0].Bet)
	assert.Equal(t, 5, table.Seats[0].TotalBet)
	assert.Equal(t, 10, table.Seats[1].Bet)
	assert.Equal(t, 10, table.Seats[1].TotalBet)
	assert.Equal(t, 0, table.Seats[2].Bet)

	// Verify stacks reduced
	assert.Equal(t, 995, table.Seats[0].Stack)
	assert.Equal(t, 990, table.Seats[1].Stack)
	assert.Equal(t, 1000, table.Seats[2].Stack)

	// Verify pot created with blinds
	assert.NotNil(t, game.Pot)
	assert.Equal(t, 15, game.Pot.Total())

	// Verify current turn is UTG (seat 2, after big blind)
	assert.Equal(t, 2, game.CurrentTurn)
}

func TestGame_PreflopFold(t *testing.T) {
	table := NewTable(TableConfig{
		ID:         "test-table",
		Name:       "Test Table",
		MaxSeats:   6,
		SmallBlind: 5,
		BigBlind:   10,
		MinBuyIn:   20,
		MaxBuyIn:   100,
	})

	p0 := NewPlayer("alice", 0, 1000)
	p1 := NewPlayer("bob", 1, 1000)
	p2 := NewPlayer("charlie", 2, 1000)

	require.NoError(t, table.Join(p0))
	require.NoError(t, table.Join(p1))
	require.NoError(t, table.Join(p2))

	game := NewGame(table)
	err := game.Start()
	require.NoError(t, err)

	// Charlie (UTG) folds
	err = game.ProcessAction(Action{PlayerID: "charlie", Type: ActionFold})
	require.NoError(t, err)
	assert.Equal(t, Folded, table.Seats[2].Status)

	// Alice (SB) folds
	err = game.ProcessAction(Action{PlayerID: "alice", Type: ActionFold})
	require.NoError(t, err)
	assert.Equal(t, Folded, table.Seats[0].Status)

	// Bob (BB) wins the hand
	// After all others fold, hand should be complete
	assert.Equal(t, StateComplete, game.State)
}

func TestGame_FullHand(t *testing.T) {
	table := NewTable(TableConfig{
		ID:         "test-table",
		Name:       "Test Table",
		MaxSeats:   6,
		SmallBlind: 5,
		BigBlind:   10,
		MinBuyIn:   20,
		MaxBuyIn:   100,
	})

	p0 := NewPlayer("alice", 0, 1000)
	p1 := NewPlayer("bob", 1, 1000)
	p2 := NewPlayer("charlie", 2, 1000)

	require.NoError(t, table.Join(p0))
	require.NoError(t, table.Join(p1))
	require.NoError(t, table.Join(p2))

	game := NewGame(table)
	err := game.Start()
	require.NoError(t, err)

	// Preflop: everyone calls/checks to match BB
	// Charlie calls 10
	err = game.ProcessAction(Action{PlayerID: "charlie", Type: ActionCall})
	require.NoError(t, err)

	// Alice calls 5 more (already has 5 in)
	err = game.ProcessAction(Action{PlayerID: "alice", Type: ActionCall})
	require.NoError(t, err)

	// Bob checks
	err = game.ProcessAction(Action{PlayerID: "bob", Type: ActionCheck})
	require.NoError(t, err)

	// Should now be on Flop
	assert.Equal(t, StateFlop, game.State)
	assert.Equal(t, 3, len(game.Community))

	// Flop: everyone checks
	err = game.ProcessAction(Action{PlayerID: "alice", Type: ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(Action{PlayerID: "bob", Type: ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(Action{PlayerID: "charlie", Type: ActionCheck})
	require.NoError(t, err)

	// Should now be on Turn
	assert.Equal(t, StateTurn, game.State)
	assert.Equal(t, 4, len(game.Community))

	// Turn: everyone checks
	err = game.ProcessAction(Action{PlayerID: "alice", Type: ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(Action{PlayerID: "bob", Type: ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(Action{PlayerID: "charlie", Type: ActionCheck})
	require.NoError(t, err)

	// Should now be on River
	assert.Equal(t, StateRiver, game.State)
	assert.Equal(t, 5, len(game.Community))

	// River: everyone checks
	err = game.ProcessAction(Action{PlayerID: "alice", Type: ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(Action{PlayerID: "bob", Type: ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(Action{PlayerID: "charlie", Type: ActionCheck})
	require.NoError(t, err)

	// Should now be Showdown
	assert.Equal(t, StateShowdown, game.State)
}

func TestGame_BettingRoundComplete(t *testing.T) {
	table := NewTable(TableConfig{
		ID:         "test-table",
		Name:       "Test Table",
		MaxSeats:   6,
		SmallBlind: 5,
		BigBlind:   10,
		MinBuyIn:   20,
		MaxBuyIn:   100,
	})

	p0 := NewPlayer("alice", 0, 1000)
	p1 := NewPlayer("bob", 1, 1000)
	p2 := NewPlayer("charlie", 2, 1000)

	require.NoError(t, table.Join(p0))
	require.NoError(t, table.Join(p1))
	require.NoError(t, table.Join(p2))

	game := NewGame(table)
	err := game.Start()
	require.NoError(t, err)

	// After start, betting round is NOT complete because Charlie hasn't acted yet
	assert.False(t, game.isBettingRoundComplete())

	// Charlie calls
	err = game.ProcessAction(Action{PlayerID: "charlie", Type: ActionCall})
	require.NoError(t, err)
	assert.False(t, game.isBettingRoundComplete())

	// Alice calls
	err = game.ProcessAction(Action{PlayerID: "alice", Type: ActionCall})
	require.NoError(t, err)
	assert.False(t, game.isBettingRoundComplete())

	// Before Bob checks, the round is not complete
	assert.False(t, game.isBettingRoundComplete())

	// Bob checks
	err = game.ProcessAction(Action{PlayerID: "bob", Type: ActionCheck})
	require.NoError(t, err)
	// After Bob checks, preflop round is complete and game advances to Flop
	assert.Equal(t, StateFlop, game.State)
}
