package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngine_Decide_FoldWeakHand(t *testing.T) {
	engine := NewEngine(Fish, "UTG")
	decision := engine.Decide([]string{"72o"}, nil, 100, 20, 1000, 40)
	assert.Equal(t, "fold", decision.Action)
}

func TestEngine_Decide_CallWithStrongHand(t *testing.T) {
	engine := NewEngine(Regular, "BTN")
	// AA facing a bet should always call or raise
	decision := engine.Decide([]string{"AH", "AD"}, nil, 100, 20, 1000, 40)
	assert.Contains(t, []string{"call", "raise"}, decision.Action)
}

func TestEngine_Decide_BetStrongHand(t *testing.T) {
	engine := NewEngine(Shark, "CO")
	// AA preflop with no bet should raise (premium hand)
	decision := engine.Decide([]string{"AH", "AD"}, nil, 100, 0, 1000, 20)
	assert.Contains(t, []string{"raise", "bet"}, decision.Action)
}

func TestEngine_PersonaDrivenDecision(t *testing.T) {
	// Maniac should almost never fold preflop with premium hand
	maniac := NewEngineWithPersona(GetPersona("maniac"), "BTN")
	hole := []string{"AH", "KD"}
	community := []string{}
	decision := maniac.Decide(hole, community, 100, 10, 1000, 20)
	assert.NotEqual(t, "fold", decision.Action, "maniac should not fold premium hand")

	// Nit should fold weak hands
	nit := NewEngineWithPersona(GetPersona("nit"), "UTG")
	holeWeak := []string{"7H", "2D"}
	decision = nit.Decide(holeWeak, community, 100, 10, 1000, 20)
	assert.Equal(t, "fold", decision.Action, "nit should fold weak hand")
}
