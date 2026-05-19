package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPersona_TAG(t *testing.T) {
	p := GetPersona("tight_aggressive")
	assert.Equal(t, "tight_aggressive", p.Style)
	assert.Equal(t, 0.18, p.VPIPTarget)
	assert.Equal(t, 0.15, p.PFRTarget)
	assert.Equal(t, 0.75, p.Aggression)
	assert.Equal(t, 0.25, p.BluffRate)
	assert.Equal(t, 0.3, p.TiltFactor)
	assert.Equal(t, 0.8, p.Patience)
}

func TestGetPersona_Maniac(t *testing.T) {
	p := GetPersona("maniac")
	assert.Equal(t, "maniac", p.Style)
	assert.Equal(t, 0.55, p.VPIPTarget)
	assert.Equal(t, 0.45, p.PFRTarget)
	assert.Equal(t, 0.95, p.Aggression)
	assert.Equal(t, 0.55, p.BluffRate)
	assert.Equal(t, 0.7, p.TiltFactor)
	assert.Equal(t, 0.1, p.Patience)
}

func TestGetPersona_Invalid(t *testing.T) {
	p := GetPersona("unknown")
	assert.Equal(t, "regular", p.Style)
	assert.Equal(t, 0.22, p.VPIPTarget)
	assert.Equal(t, 0.16, p.PFRTarget)
}

func TestGetPersona_AllStyles(t *testing.T) {
	for style, expected := range personas {
		p := GetPersona(style)
		assert.Equal(t, expected.Style, p.Style, "style mismatch for %s", style)
		assert.Equal(t, expected.VPIPTarget, p.VPIPTarget, "VPIP mismatch for %s", style)
		assert.Equal(t, expected.PFRTarget, p.PFRTarget, "PFR mismatch for %s", style)
	}
}

func TestAllPersonas(t *testing.T) {
	names := AllPersonas()
	assert.Len(t, names, 8)
	for _, name := range names {
		_, ok := personas[name]
		assert.True(t, ok, "%s should be a valid persona", name)
	}
}
