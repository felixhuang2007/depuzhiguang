package engine

import (
	"testing"
)

func TestPot_SingleWinner(t *testing.T) {
	p := NewPot()
	p.AddBet("p1", 100)
	p.AddBet("p2", 100)
	p.AddBet("p3", 100)
	p.CloseBettingRound()

	winners := map[string]struct{}{"p1": {}}
	result := p.Award(winners)

	if result["p1"] != 300 {
		t.Errorf("expected p1 to win 300, got %d", result["p1"])
	}
	if p.Total() != 300 {
		t.Errorf("expected total 300, got %d", p.Total())
	}
}

func TestPot_SplitPot(t *testing.T) {
	p := NewPot()
	p.AddBet("p1", 100)
	p.AddBet("p2", 100)
	p.AddBet("p3", 100)
	p.CloseBettingRound()

	winners := map[string]struct{}{"p1": {}, "p2": {}}
	result := p.Award(winners)

	if result["p1"] != 150 {
		t.Errorf("expected p1 to win 150, got %d", result["p1"])
	}
	if result["p2"] != 150 {
		t.Errorf("expected p2 to win 150, got %d", result["p2"])
	}
}

func TestPot_SidePot_Basic(t *testing.T) {
	p := NewPot()
	p.AddBet("p1", 50)
	p.AddBet("p2", 100)
	p.AddBet("p3", 100)
	p.CloseBettingRound()

	// Main pot (150) split among all 3 winners
	// Side pot (100) split between p2 and p3
	winners := map[string]struct{}{"p1": {}, "p2": {}, "p3": {}}
	result := p.Award(winners)

	if result["p1"] != 50 {
		t.Errorf("expected p1 to win 50, got %d", result["p1"])
	}
	if result["p2"] != 100 {
		t.Errorf("expected p2 to win 100, got %d", result["p2"])
	}
	if result["p3"] != 100 {
		t.Errorf("expected p3 to win 100, got %d", result["p3"])
	}
	if p.Total() != 250 {
		t.Errorf("expected total 250, got %d", p.Total())
	}
}

func TestPot_SidePot_AllInWinner(t *testing.T) {
	p := NewPot()
	p.AddBet("p1", 50)
	p.AddBet("p2", 100)
	p.AddBet("p3", 100)
	p.CloseBettingRound()

	// p1 wins main pot only, p2 and p3 are not winners
	winners := map[string]struct{}{"p1": {}}
	result := p.Award(winners)

	if result["p1"] != 150 {
		t.Errorf("expected p1 to win 150 (main pot), got %d", result["p1"])
	}
	if p.Total() != 250 {
		t.Errorf("expected total 250, got %d", p.Total())
	}
}

func TestPot_MultiRound(t *testing.T) {
	p := NewPot()
	// Preflop
	p.AddBet("p1", 100)
	p.AddBet("p2", 100)
	p.AddBet("p3", 100)
	p.CloseBettingRound()

	// Flop
	p.AddBet("p1", 50)
	p.AddBet("p2", 50)
	p.AddBet("p3", 50)
	p.CloseBettingRound()

	winners := map[string]struct{}{"p1": {}}
	result := p.Award(winners)

	if result["p1"] != 450 {
		t.Errorf("expected p1 to win 450, got %d", result["p1"])
	}
	if p.Total() != 450 {
		t.Errorf("expected total 450, got %d", p.Total())
	}
}

func TestPot_RemainderDistribution(t *testing.T) {
	p := NewPot()
	p.AddBet("p1", 100)
	p.AddBet("p2", 100)
	p.AddBet("p3", 100)
	p.CloseBettingRound()

	// 300 / 2 = 150 each, no remainder
	// Let's make a remainder: 3 players bet 100 = 300 total
	// 2 winners -> 150 each, remainder 0. Not good.
	// Let's do 4 players, 3 winners: 400 / 3 = 133 rem 1
	p2 := NewPot()
	p2.AddBet("p1", 100)
	p2.AddBet("p2", 100)
	p2.AddBet("p3", 100)
	p2.AddBet("p4", 100)
	p2.CloseBettingRound()

	winners := map[string]struct{}{"p1": {}, "p2": {}, "p3": {}}
	result := p2.Award(winners)

	// 400 / 3 = 133 rem 1 -> p1 gets 134, p2 gets 133, p3 gets 133
	if result["p1"] != 134 {
		t.Errorf("expected p1 to win 134, got %d", result["p1"])
	}
	if result["p2"] != 133 {
		t.Errorf("expected p2 to win 133, got %d", result["p2"])
	}
	if result["p3"] != 133 {
		t.Errorf("expected p3 to win 133, got %d", result["p3"])
	}
}
