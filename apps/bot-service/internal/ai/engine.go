package ai

import (
	"math"
	"math/rand"
	"time"
)

// Difficulty level defines bot behavior
type Difficulty int

const (
	Fish Difficulty = iota
	Regular
	Shark
	Whale
)

// Decision represents a bot action
type Decision struct {
	Action string
	Amount int
	Delay  time.Duration
}

// Engine makes poker decisions
type Engine struct {
	Persona  Persona
	Position string
	// Track stats for adaptive behavior
	handsPlayed       int
	handsWon          int
	consecutiveLosses int
}

func NewEngineWithPersona(p Persona, pos string) *Engine {
	return &Engine{Persona: p, Position: pos}
}

func NewEngine(diff Difficulty, pos string) *Engine {
	return NewEngineWithPersona(GetPersona("regular"), pos)
}

func (e *Engine) Decide(hole []string, community []string, pot int, toCall int, stack int, minRaise int) Decision {
	delay := time.Duration(2+rand.Intn(6)) * time.Second

	strength := e.evaluateHandStrength(hole, community)

	// Apply tilt: consecutive losses make player more aggressive
	tiltAdjustment := float64(e.consecutiveLosses) * e.Persona.TiltFactor * 0.1
	effectiveAggression := e.Persona.Aggression + tiltAdjustment

	if stack <= 0 {
		return Decision{Action: "check", Delay: delay}
	}

	if toCall > 0 {
		potOdds := float64(toCall) / float64(pot+toCall)
		// Adjust call threshold by persona patience
		callThreshold := potOdds + 0.2 - (e.Persona.Patience * 0.15)

		if strength > callThreshold {
			if strength > 0.8+effectiveAggression*0.15 && stack > toCall*3 {
				return Decision{Action: "raise", Amount: minRaise * 2, Delay: delay}
			}
			return Decision{Action: "call", Amount: toCall, Delay: delay}
		}
		if strength < potOdds-0.1 {
			return Decision{Action: "fold", Delay: delay}
		}
		// Calling station rarely folds
		if e.Persona.Style == "calling_station" && rand.Float64() < 0.9 {
			return Decision{Action: "call", Amount: toCall, Delay: delay}
		}
		if rand.Float64() < e.Persona.Patience {
			return Decision{Action: "fold", Delay: delay}
		}
		return Decision{Action: "call", Amount: toCall, Delay: delay}
	}

	// No bet to call - decide whether to bet
	betThreshold := 0.6 - effectiveAggression*0.2
	if strength > betThreshold {
		betSize := pot / 2
		if betSize < minRaise {
			betSize = minRaise
		}
		return Decision{Action: "bet", Amount: betSize, Delay: delay}
	}
	return Decision{Action: "check", Delay: delay}
}

func (e *Engine) evaluateHandStrength(hole, community []string) float64 {
	score := 0.0
	highCards := map[string]float64{"A": 0.15, "K": 0.12, "Q": 0.10, "J": 0.08, "T": 0.06}
	for _, card := range hole {
		if len(card) > 0 {
			rank := string(card[0])
			if v, ok := highCards[rank]; ok {
				score += v
			}
		}
	}
	if len(hole) == 2 && len(hole[0]) > 0 && len(hole[1]) > 0 && hole[0][0] == hole[1][0] {
		score += 0.3
	}
	if len(hole) == 2 && len(hole[0]) > 1 && len(hole[1]) > 1 && hole[0][1] == hole[1][1] {
		score += 0.05
	}
	return math.Min(score, 1.0)
}
