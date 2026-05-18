package table

import (
	"testing"

	"github.com/depuzhiguang/game-server/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunShowdown_SingleWinner(t *testing.T) {
	table := NewTable(TableConfig{
		MaxSeats: 6, SmallBlind: 5, BigBlind: 10, MinBuyIn: 20, MaxBuyIn: 100,
	})

	p0 := NewPlayer("alice", 0, 1000)
	p1 := NewPlayer("bob", 1, 1000)
	p2 := NewPlayer("charlie", 2, 1000)

	table.Join(p0)
	table.Join(p1)
	table.Join(p2)

	game := NewGame(table)
	game.Start()

	// Everyone checks through to showdown
	game.ProcessAction(Action{PlayerID: "charlie", Type: ActionCall})
	game.ProcessAction(Action{PlayerID: "alice", Type: ActionCall})
	game.ProcessAction(Action{PlayerID: "bob", Type: ActionCheck})

	// Flop
	game.ProcessAction(Action{PlayerID: "alice", Type: ActionCheck})
	game.ProcessAction(Action{PlayerID: "bob", Type: ActionCheck})
	game.ProcessAction(Action{PlayerID: "charlie", Type: ActionCheck})

	// Turn
	game.ProcessAction(Action{PlayerID: "alice", Type: ActionCheck})
	game.ProcessAction(Action{PlayerID: "bob", Type: ActionCheck})
	game.ProcessAction(Action{PlayerID: "charlie", Type: ActionCheck})

	// River
	game.ProcessAction(Action{PlayerID: "alice", Type: ActionCheck})
	game.ProcessAction(Action{PlayerID: "bob", Type: ActionCheck})
	game.ProcessAction(Action{PlayerID: "charlie", Type: ActionCheck})

	require.Equal(t, StateShowdown, game.State)

	results := RunShowdown(game)
	require.NotNil(t, results)
	require.Equal(t, 3, len(results))

	// Find winner
	var winner string
	for _, r := range results {
		if r.WonAmount > 0 {
			winner = r.PlayerID
		}
	}
	assert.NotEmpty(t, winner)
	assert.Equal(t, StateComplete, game.State)
}

func TestRunShowdown_AllFolded(t *testing.T) {
	table := NewTable(TableConfig{
		MaxSeats: 6, SmallBlind: 5, BigBlind: 10, MinBuyIn: 20, MaxBuyIn: 100,
	})

	p0 := NewPlayer("alice", 0, 1000)
	p1 := NewPlayer("bob", 1, 1000)

	table.Join(p0)
	table.Join(p1)

	game := NewGame(table)
	game.Start()

	// Alice folds
	game.ProcessAction(Action{PlayerID: "alice", Type: ActionFold})

	// Bob wins by default
	require.Equal(t, StateComplete, game.State)

	// No showdown needed when all fold
	results := RunShowdown(game)
	assert.Nil(t, results)
}

func TestBetterKickers(t *testing.T) {
	assert.True(t, betterKickers([]engine.Rank{engine.Ace, engine.King}, []engine.Rank{engine.Ace, engine.Queen}))
	assert.False(t, betterKickers([]engine.Rank{engine.Ace, engine.Queen}, []engine.Rank{engine.Ace, engine.King}))
	assert.True(t, betterKickers([]engine.Rank{engine.Ace}, []engine.Rank{engine.King}))
}
