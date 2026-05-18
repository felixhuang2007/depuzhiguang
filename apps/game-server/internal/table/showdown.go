package table

import (
	"github.com/depuzhiguang/game-server/internal/engine"
)

// ShowdownResult contains the outcome for a single player
type ShowdownResult struct {
	PlayerID  string
	HandRank  engine.HandRank
	Result    engine.EvalResult
	WonAmount int
}

// RunShowdown evaluates all hands and distributes the pot
func RunShowdown(game *Game) []ShowdownResult {
	if game.State != StateShowdown {
		return nil
	}

	// Close final betting round
	game.Pot.CloseBettingRound()

	// Evaluate each player's best hand
	type playerHand struct {
		player *Player
		rank   engine.HandRank
		result engine.EvalResult
	}

	var hands []playerHand
	for _, seat := range game.Table.OccupiedSeats() {
		p := game.Table.Seats[seat]
		if !p.IsInHand() {
			continue
		}
		rank, result := engine.EvaluateBest(p.Holes[:], game.Community)
		hands = append(hands, playerHand{player: p, rank: rank, result: result})
	}

	if len(hands) == 0 {
		return nil
	}

	// Find best hand rank
	bestRank := hands[0].rank
	for _, h := range hands {
		if h.rank > bestRank {
			bestRank = h.rank
		}
	}

	// Collect all players with the best hand
	winners := make(map[string]struct{})
	var bestKickers []engine.Rank
	for _, h := range hands {
		if h.rank == bestRank {
			if bestKickers == nil || betterKickers(h.result.Kickers, bestKickers) {
				bestKickers = h.result.Kickers
				winners = make(map[string]struct{})
				winners[h.player.ID] = struct{}{}
			} else if kickerEqual(h.result.Kickers, bestKickers) {
				winners[h.player.ID] = struct{}{}
			}
		}
	}

	// Award pot
	awards := game.Pot.Award(winners)

	// Build results
	var results []ShowdownResult
	for _, h := range hands {
		results = append(results, ShowdownResult{
			PlayerID:  h.player.ID,
			HandRank:  h.rank,
			Result:    h.result,
			WonAmount: awards[h.player.ID],
		})
	}

	game.State = StateComplete
	return results
}

func betterKickers(a, b []engine.Rank) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return len(a) > len(b)
}

func kickerEqual(a, b []engine.Rank) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
