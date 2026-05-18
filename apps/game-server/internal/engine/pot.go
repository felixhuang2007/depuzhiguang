package engine

import "sort"

type subPot struct {
	amount   int
	eligible []string
}

// Pot tracks bets across betting rounds and manages main pot / side pots.
type Pot struct {
	bets       map[string]int
	closedPots []subPot
}

// NewPot creates a new Pot.
func NewPot() *Pot {
	return &Pot{
		bets:       make(map[string]int),
		closedPots: make([]subPot, 0),
	}
}

// AddBet adds amount to the player's total bet for the current round.
// If amount is 0, the player is removed (folded).
func (p *Pot) AddBet(playerID string, amount int) {
	if amount == 0 {
		delete(p.bets, playerID)
		return
	}
	p.bets[playerID] += amount
}

// CloseBettingRound finalizes current bets into sub-pots using the short-stack algorithm.
func (p *Pot) CloseBettingRound() {
	if len(p.bets) == 0 {
		return
	}

	// Build a sorted list of (playerID, contribution) ascending.
	type playerBet struct {
		id     string
		amount int
	}

	bets := make([]playerBet, 0, len(p.bets))
	for id, amount := range p.bets {
		bets = append(bets, playerBet{id: id, amount: amount})
	}
	sort.Slice(bets, func(i, j int) bool {
		return bets[i].amount < bets[j].amount
	})

	prev := 0
	for i := 0; i < len(bets); i++ {
		if bets[i].amount == prev {
			continue
		}
		level := bets[i].amount - prev
		eligible := make([]string, 0, len(bets)-i)
		for j := i; j < len(bets); j++ {
			eligible = append(eligible, bets[j].id)
		}
		p.closedPots = append(p.closedPots, subPot{
			amount:   level * len(eligible),
			eligible: eligible,
		})
		prev = bets[i].amount
	}

	// Clear current bets for next round.
	p.bets = make(map[string]int)
}

// Award distributes all closed pots to winners.
// Each sub-pot is split among eligible winners. Remainder chips are distributed
// one-by-one to winners in order of appearance in eligible list.
func (p *Pot) Award(winners map[string]struct{}) map[string]int {
	result := make(map[string]int)

	for _, sp := range p.closedPots {
		// Find eligible winners (intersection).
		eligibleWinners := make([]string, 0)
		for _, pid := range sp.eligible {
			if _, ok := winners[pid]; ok {
				eligibleWinners = append(eligibleWinners, pid)
			}
		}

		if len(eligibleWinners) == 0 {
			continue
		}

		share := sp.amount / len(eligibleWinners)
		remainder := sp.amount % len(eligibleWinners)

		for i, pid := range eligibleWinners {
			result[pid] += share
			if i < remainder {
				result[pid]++
			}
		}
	}

	return result
}

// Total returns the sum of all closed pot amounts.
func (p *Pot) Total() int {
	total := 0
	for _, sp := range p.closedPots {
		total += sp.amount
	}
	return total
}
