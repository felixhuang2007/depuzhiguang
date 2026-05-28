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
	var style string
	switch diff {
	case Fish:
		style = "loose_passive"
	case Shark:
		style = "tight_aggressive"
	case Whale:
		style = "rock"
	default:
		style = "adaptive"
	}
	return NewEngineWithPersona(GetPersona(style), pos)
}

func (e *Engine) RecordResult(won bool) {
	e.handsPlayed++
	if won {
		e.handsWon++
		e.consecutiveLosses = 0
	} else {
		e.consecutiveLosses++
	}
}

func (e *Engine) Decide(hole []string, community []string, pot int, toCall int, stack int, minRaise int) Decision {
	delay := time.Duration(2+rand.Intn(6)) * time.Second

	if stack <= 0 {
		return Decision{Action: "check", Delay: delay}
	}

	// Determine if preflop or postflop
	isPreflop := len(community) == 0

	var strength float64
	if isPreflop {
		strength = e.evaluatePreflopStrength(hole)
	} else {
		strength = e.evaluateHandStrength(hole, community)
	}

	// Apply tilt adjustment
	tiltAdjustment := float64(e.consecutiveLosses) * e.Persona.TiltFactor * 0.05
	effectiveAggression := math.Min(1.0, e.Persona.Aggression+tiltAdjustment)

	// Position factor: late position allows more aggressive/loose play
	positionFactor := e.positionFactor()

	if isPreflop {
		return e.decidePreflop(strength, toCall, stack, minRaise, effectiveAggression, positionFactor, delay)
	}
	return e.decidePostflop(strength, toCall, stack, minRaise, effectiveAggression, positionFactor, pot, delay)
}

func (e *Engine) positionFactor() float64 {
	switch e.Position {
	case "BTN":
		return 1.0 // Button (dealer) - best position
	case "CO":
		return 0.8 // Cutoff
	case "MP":
		return 0.5 // Middle position
	case "UTG":
		return 0.2 // Under the gun - worst position
	case "BB", "SB":
		return 0.6 // Blinds
	default:
		return 0.5
	}
}

// decidePreflop makes preflop decisions based on hand tier and persona.
func (e *Engine) decidePreflop(strength float64, toCall int, stack int, minRaise int, aggression float64, positionFactor float64, delay time.Duration) Decision {
	// Adjust playable threshold by persona and position
	playableThreshold := 1.0 - e.Persona.VPIPTarget - positionFactor*0.15
	playableThreshold = math.Max(0.1, math.Min(0.9, playableThreshold))

	// Raise threshold
	raiseThreshold := 1.0 - e.Persona.PFRTarget - positionFactor*0.1
	raiseThreshold = math.Max(0.15, math.Min(0.95, raiseThreshold))

	// Special persona overrides
	switch e.Persona.Style {
	case "calling_station":
		// Calling station: rarely folds preflop if there's any bet, rarely raises
		if toCall > 0 {
			if toCall <= stack/10 || strength > 0.3 {
				return Decision{Action: "call", Amount: toCall, Delay: delay}
			}
		}
		if strength > 0.7 {
			return Decision{Action: "raise", Amount: minRaise * 2, Delay: delay}
		}
		return Decision{Action: "fold", Delay: delay}

	case "maniac":
		// Maniac: plays almost anything, frequently raises
		if strength > 0.15 || rand.Float64() < 0.4 {
			if toCall > 0 {
				if rand.Float64() < aggression {
					return Decision{Action: "raise", Amount: minRaise * (2 + rand.Intn(3)), Delay: delay}
				}
				return Decision{Action: "call", Amount: toCall, Delay: delay}
			}
			return Decision{Action: "raise", Amount: minRaise * 2, Delay: delay}
		}
		return Decision{Action: "fold", Delay: delay}

	case "rock", "nit":
		// Very tight: only premium hands
		if strength > 0.75 {
			if toCall > 0 {
				return Decision{Action: "raise", Amount: minRaise * 3, Delay: delay}
			}
			return Decision{Action: "raise", Amount: minRaise * 3, Delay: delay}
		}
		if strength > 0.55 && toCall <= minRaise {
			return Decision{Action: "call", Amount: toCall, Delay: delay}
		}
		return Decision{Action: "fold", Delay: delay}

	case "loose_passive":
		// Plays many hands but rarely raises
		if strength > playableThreshold {
			if toCall > 0 {
				return Decision{Action: "call", Amount: toCall, Delay: delay}
			}
			if strength > 0.65 {
				return Decision{Action: "raise", Amount: minRaise, Delay: delay}
			}
			return Decision{Action: "check", Delay: delay}
		}
		if toCall == 0 {
			return Decision{Action: "check", Delay: delay}
		}
		return Decision{Action: "fold", Delay: delay}

	case "loose_aggressive":
		// Plays many hands and frequently raises
		if strength > playableThreshold || rand.Float64() < e.Persona.BluffRate {
			if toCall > 0 {
				if strength > raiseThreshold || rand.Float64() < aggression {
					return Decision{Action: "raise", Amount: minRaise * (2 + rand.Intn(2)), Delay: delay}
				}
				return Decision{Action: "call", Amount: toCall, Delay: delay}
			}
			return Decision{Action: "raise", Amount: minRaise * 2, Delay: delay}
		}
		if toCall == 0 {
			return Decision{Action: "check", Delay: delay}
		}
		return Decision{Action: "fold", Delay: delay}
	}

	// Default / tight_aggressive / adaptive logic
	if strength > raiseThreshold {
		if toCall > 0 {
			raiseSize := minRaise * 2
			if strength > 0.85 {
				raiseSize = minRaise * 3
			}
			if raiseSize > stack {
				raiseSize = stack
			}
			return Decision{Action: "raise", Amount: raiseSize, Delay: delay}
		}
		return Decision{Action: "raise", Amount: minRaise * 2, Delay: delay}
	}
	if strength > playableThreshold {
		if toCall > 0 {
			if toCall > stack/3 && strength < 0.6 {
				return Decision{Action: "fold", Delay: delay}
			}
			return Decision{Action: "call", Amount: toCall, Delay: delay}
		}
		return Decision{Action: "check", Delay: delay}
	}
	if toCall == 0 {
		return Decision{Action: "check", Delay: delay}
	}
	return Decision{Action: "fold", Delay: delay}
}

// decidePostflop makes postflop decisions based on evaluated hand strength.
func (e *Engine) decidePostflop(strength float64, toCall int, stack int, minRaise int, aggression float64, positionFactor float64, pot int, delay time.Duration) Decision {
	potOdds := 0.0
	if pot+toCall > 0 {
		potOdds = float64(toCall) / float64(pot+toCall)
	}

	// Bluff chance increases in late position
	bluffChance := e.Persona.BluffRate + positionFactor*0.1
	bluffChance = math.Min(1.0, bluffChance)

	// Special personas
	switch e.Persona.Style {
	case "calling_station":
		// Almost never folds postflop, never raises
		if toCall > 0 {
			if toCall <= stack/5 || strength > 0.15 {
				return Decision{Action: "call", Amount: toCall, Delay: delay}
			}
		}
		if toCall == 0 {
			return Decision{Action: "check", Delay: delay}
		}
		return Decision{Action: "fold", Delay: delay}

	case "maniac":
		// Frequently bets/raises regardless of strength
		if toCall > 0 {
			if rand.Float64() < aggression || strength > 0.3 {
				return Decision{Action: "raise", Amount: minRaise * (2 + rand.Intn(3)), Delay: delay}
			}
			return Decision{Action: "call", Amount: toCall, Delay: delay}
		}
		return Decision{Action: "raise", Amount: pot/2 + minRaise, Delay: delay}

	case "rock", "nit":
		// Only bets strong hands, folds easily
		if strength > 0.8 {
			if toCall > 0 {
				return Decision{Action: "raise", Amount: minRaise * 3, Delay: delay}
			}
			return Decision{Action: "bet", Amount: pot*2/3 + minRaise, Delay: delay}
		}
		if strength > potOdds+0.15 {
			if toCall > 0 {
				return Decision{Action: "call", Amount: toCall, Delay: delay}
			}
			return Decision{Action: "check", Delay: delay}
		}
		if toCall == 0 {
			return Decision{Action: "check", Delay: delay}
		}
		return Decision{Action: "fold", Delay: delay}

	case "loose_passive":
		// Calls with medium+ hands, rarely bets
		if strength > 0.5 || (strength > potOdds && toCall <= stack/5) {
			if toCall > 0 {
				return Decision{Action: "call", Amount: toCall, Delay: delay}
			}
			if strength > 0.65 {
				return Decision{Action: "bet", Amount: pot/2, Delay: delay}
			}
			return Decision{Action: "check", Delay: delay}
		}
		if toCall == 0 {
			return Decision{Action: "check", Delay: delay}
		}
		return Decision{Action: "fold", Delay: delay}

	case "loose_aggressive":
		// Bets/raises frequently, floats with draws
		if strength > 0.45 || rand.Float64() < bluffChance {
			if toCall > 0 {
				if strength > 0.55 || rand.Float64() < aggression {
					return Decision{Action: "raise", Amount: minRaise * (2 + rand.Intn(2)), Delay: delay}
				}
				return Decision{Action: "call", Amount: toCall, Delay: delay}
			}
			return Decision{Action: "bet", Amount: pot*2/3 + minRaise, Delay: delay}
		}
		if strength > potOdds {
			if toCall > 0 {
				return Decision{Action: "call", Amount: toCall, Delay: delay}
			}
			return Decision{Action: "check", Delay: delay}
		}
		if toCall == 0 {
			return Decision{Action: "check", Delay: delay}
		}
		return Decision{Action: "fold", Delay: delay}
	}

	// Default / tight_aggressive / adaptive
	// Value betting: strong hands
	if strength > 0.75 {
		if toCall > 0 {
			raiseSize := minRaise * 2
			if strength > 0.9 {
				raiseSize = pot
			}
			if raiseSize > stack {
				raiseSize = stack
			}
			return Decision{Action: "raise", Amount: raiseSize, Delay: delay}
		}
		betSize := pot * 2 / 3
		if betSize < minRaise {
			betSize = minRaise
		}
		if betSize > stack {
			betSize = stack
		}
		return Decision{Action: "bet", Amount: betSize, Delay: delay}
	}

	// Semi-bluff / draw: medium strength with aggression
	if strength > 0.45 {
		if toCall > 0 {
			if strength > potOdds+0.15 {
				if strength > 0.6 && rand.Float64() < aggression {
					return Decision{Action: "raise", Amount: minRaise * 2, Delay: delay}
				}
				return Decision{Action: "call", Amount: toCall, Delay: delay}
			}
			return Decision{Action: "fold", Delay: delay}
		}
		if rand.Float64() < aggression*0.7 {
			betSize := pot / 2
			if betSize < minRaise {
				betSize = minRaise
			}
			return Decision{Action: "bet", Amount: betSize, Delay: delay}
		}
		return Decision{Action: "check", Delay: delay}
	}

	// Bluff occasionally based on persona and position
	if rand.Float64() < bluffChance && toCall == 0 {
		betSize := pot / 2
		if betSize < minRaise {
			betSize = minRaise
		}
		return Decision{Action: "bet", Amount: betSize, Delay: delay}
	}

	// Weak hand: fold to bets, check if possible
	if toCall > 0 {
		if strength > potOdds-0.05 {
			return Decision{Action: "call", Amount: toCall, Delay: delay}
		}
		return Decision{Action: "fold", Delay: delay}
	}
	return Decision{Action: "check", Delay: delay}
}

// evaluatePreflopStrength returns a strength score (0.0 - 1.0) for hole cards preflop.
func (e *Engine) evaluatePreflopStrength(hole []string) float64 {
	if len(hole) != 2 {
		return 0.0
	}

	c1 := parseCard(hole[0])
	c2 := parseCard(hole[1])
	if c1.Rank == InvalidRank || c2.Rank == InvalidRank {
		return 0.0
	}

	// Ensure r1 >= r2
	r1, r2 := c1.Rank, c2.Rank
	if r2 > r1 {
		r1, r2 = r2, r1
	}

	suited := c1.Suit == c2.Suit && c1.Suit != InvalidSuit
	paired := r1 == r2

	// Base score from rank
	score := 0.0

	if paired {
		// Pocket pairs
		switch r1 {
		case Ace:
			score = 0.95
		case King:
			score = 0.88
		case Queen:
			score = 0.82
		case Jack:
			score = 0.76
		case Ten:
			score = 0.70
		case Nine:
			score = 0.63
		case Eight:
			score = 0.57
		case Seven:
			score = 0.51
		case Six:
			score = 0.45
		case Five:
			score = 0.40
		case Four:
			score = 0.35
		case Three:
			score = 0.30
		case Two:
			score = 0.25
		}
	} else {
		// Unpaired hands
		gap := int(r1) - int(r2)

		// Base from high card
		highScore := map[Rank]float64{
			Ace:   0.55,
			King:  0.48,
			Queen: 0.43,
			Jack:  0.38,
			Ten:   0.33,
			Nine:  0.28,
			Eight: 0.24,
			Seven: 0.20,
			Six:   0.17,
			Five:  0.14,
			Four:  0.12,
			Three: 0.10,
			Two:   0.08,
		}[r1]

		// Adjust for kicker
		kickerBonus := float64(r2) / float64(Ace) * 0.08

		// Connectedness bonus
		connectedBonus := 0.0
		if gap == 1 {
			connectedBonus = 0.06
		} else if gap == 2 {
			connectedBonus = 0.03
		} else if gap == 3 {
			connectedBonus = 0.01
		}

		score = highScore + kickerBonus + connectedBonus
	}

	// Suited bonus
	if suited {
		score += 0.04
	}

	return math.Min(score, 1.0)
}

// evaluateHandStrength evaluates hand strength using the proper evaluator plus draw potential.
func (e *Engine) evaluateHandStrength(hole, community []string) float64 {
	rank, result := EvaluateBest(hole, community)
	strength := handRankToStrength(rank, result.Kickers)

	// Add draw potential (only for non-made hands)
	if rank <= OnePair {
		all := parseCards(append(hole, community...))
		if hasFlushDraw(all) {
			if isOpenEndedStraightDraw(all) {
				strength += 0.18 // combo draw
			} else {
				strength += 0.12 // flush draw
			}
		} else if isOpenEndedStraightDraw(all) {
			strength += 0.10 // open-ended straight draw
		} else if hasStraightDraw(all) {
			strength += 0.05 // gutshot
		}
	}

	return math.Min(strength, 1.0)
}

func getPhase(state int) string {
	switch state {
	case 2:
		return "preflop"
	case 3:
		return "flop"
	case 4:
		return "turn"
	case 5:
		return "river"
	default:
		return "unknown"
	}
}

func cardToString(c CardJSON) string {
	ranks := []string{"", "A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
	suits := []string{"C", "D", "H", "S"}
	if c.Rank >= 1 && c.Rank <= 14 && c.Suit >= 0 && c.Suit <= 3 {
		return ranks[c.Rank] + suits[c.Suit]
	}
	return "?"
}

type CardJSON struct {
	Suit int `json:"suit"`
	Rank int `json:"rank"`
}

func (e *Engine) SetPosition(pos string) {
	e.Position = pos
}

func (e *Engine) GetPersonaStyle() string {
	return e.Persona.Style
}
