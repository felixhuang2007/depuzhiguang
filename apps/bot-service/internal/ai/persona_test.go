package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPersona_TAG(t *testing.T) {
	p := GetPersona("tight_aggressive")
	assert.Equal(t, "tight_aggressive", p.Style)
	assert.True(t, p.VPIPTarget >= 0.15 && p.VPIPTarget <= 0.22)
	assert.True(t, p.PFRTarget >= 0.12 && p.PFRTarget <= 0.18)
}

func TestGetPersona_Maniac(t *testing.T) {
	p := GetPersona("maniac")
	assert.Equal(t, "maniac", p.Style)
	assert.True(t, p.VPIPTarget >= 0.45)
	assert.True(t, p.BluffRate >= 0.5)
}

func TestGetPersona_Invalid(t *testing.T) {
	p := GetPersona("unknown")
	assert.Equal(t, "regular", p.Style)
}
