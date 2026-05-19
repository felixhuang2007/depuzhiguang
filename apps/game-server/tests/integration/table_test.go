package integration

import (
	"testing"

	"github.com/depuzhiguang/game-server/internal/engine"
	"github.com/depuzhiguang/game-server/internal/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullHand_ThreePlayers(t *testing.T) {
	// Setup table
	tbl := table.NewTable(table.TableConfig{
		ID:         "test-table",
		Name:       "Integration Test",
		MaxSeats:   6,
		SmallBlind: 5,
		BigBlind:   10,
		MinBuyIn:   20,
		MaxBuyIn:   100,
	})

	p0 := table.NewPlayer("alice", 0, 1000)
	p1 := table.NewPlayer("bob", 1, 1000)
	p2 := table.NewPlayer("charlie", 2, 1000)

	require.NoError(t, tbl.Join(p0))
	require.NoError(t, tbl.Join(p1))
	require.NoError(t, tbl.Join(p2))

	// Start hand
	game := table.NewGame(tbl)
	err := game.Start()
	require.NoError(t, err)
	assert.Equal(t, table.StatePreflop, game.State)

	// Verify initial state
	assert.Equal(t, 46, game.Deck.Len()) // 52 - 6 hole cards
	assert.Equal(t, 15, game.Pot.Total())
	assert.Equal(t, 2, game.CurrentTurn)

	// Preflop: Charlie calls 10, Alice calls 5, Bob checks
	err = game.ProcessAction(table.Action{PlayerID: "charlie", Type: table.ActionCall})
	require.NoError(t, err)
	err = game.ProcessAction(table.Action{PlayerID: "alice", Type: table.ActionCall})
	require.NoError(t, err)
	err = game.ProcessAction(table.Action{PlayerID: "bob", Type: table.ActionCheck})
	require.NoError(t, err)

	// Verify flop
	assert.Equal(t, table.StateFlop, game.State)
	assert.Equal(t, 3, len(game.Community))
	assert.Equal(t, 30, game.Pot.Total())

	// Flop: all check
	err = game.ProcessAction(table.Action{PlayerID: "alice", Type: table.ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(table.Action{PlayerID: "bob", Type: table.ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(table.Action{PlayerID: "charlie", Type: table.ActionCheck})
	require.NoError(t, err)

	// Verify turn
	assert.Equal(t, table.StateTurn, game.State)
	assert.Equal(t, 4, len(game.Community))

	// Turn: all check
	err = game.ProcessAction(table.Action{PlayerID: "alice", Type: table.ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(table.Action{PlayerID: "bob", Type: table.ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(table.Action{PlayerID: "charlie", Type: table.ActionCheck})
	require.NoError(t, err)

	// Verify river
	assert.Equal(t, table.StateRiver, game.State)
	assert.Equal(t, 5, len(game.Community))

	// River: all check
	err = game.ProcessAction(table.Action{PlayerID: "alice", Type: table.ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(table.Action{PlayerID: "bob", Type: table.ActionCheck})
	require.NoError(t, err)
	err = game.ProcessAction(table.Action{PlayerID: "charlie", Type: table.ActionCheck})
	require.NoError(t, err)

	// Verify showdown
	assert.Equal(t, table.StateShowdown, game.State)

	// Run showdown
	results := table.RunShowdown(game)
	require.NotNil(t, results)
	assert.Equal(t, table.StateComplete, game.State)

	// Verify results
	assert.Equal(t, 3, len(results))

	// Total pot should be distributed
	totalWon := 0
	for _, r := range results {
		totalWon += r.WonAmount
	}
	assert.Equal(t, 30, totalWon)

	// Winner should have best hand
	var winner table.ShowdownResult
	for _, r := range results {
		if r.WonAmount > winner.WonAmount {
			winner = r
		}
	}
	assert.NotEmpty(t, winner.PlayerID)
	assert.True(t, winner.HandRank >= engine.HighCard)
}

func TestHand_EarlyFold(t *testing.T) {
	tbl := table.NewTable(table.TableConfig{
		MaxSeats: 6, SmallBlind: 5, BigBlind: 10, MinBuyIn: 20, MaxBuyIn: 100,
	})

	p0 := table.NewPlayer("alice", 0, 1000)
	p1 := table.NewPlayer("bob", 1, 1000)
	p2 := table.NewPlayer("charlie", 2, 1000)

	tbl.Join(p0)
	tbl.Join(p1)
	tbl.Join(p2)

	game := table.NewGame(tbl)
	game.Start()

	// Charlie folds
	game.ProcessAction(table.Action{PlayerID: "charlie", Type: table.ActionFold})
	// Alice folds
	game.ProcessAction(table.Action{PlayerID: "alice", Type: table.ActionFold})

	// Bob wins by default
	assert.Equal(t, table.StateComplete, game.State)
	assert.Equal(t, 15, game.Pot.Total())
}
