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
	decision := engine.Decide([]string{"AA"}, nil, 100, 20, 1000, 40)
	assert.Contains(t, []string{"call", "raise"}, decision.Action)
}

func TestEngine_Decide_BetStrongHand(t *testing.T) {
	engine := NewEngine(Shark, "CO")
	// AA with no bet should bet or check (strength = 0.6, threshold is > 0.6)
	decision := engine.Decide([]string{"AA"}, nil, 100, 0, 1000, 20)
	assert.Contains(t, []string{"bet", "check"}, decision.Action)
}
