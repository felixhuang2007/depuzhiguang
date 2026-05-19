package registrar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateProfile(t *testing.T) {
	p := GenerateProfile(0, "tight_aggressive")
	assert.NotEmpty(t, p.Username)
	assert.NotEmpty(t, p.Password)
	assert.Equal(t, "tight_aggressive", p.Style)
	assert.Equal(t, 10000, p.InitialGold)
}

func TestGenerateProfile_UniqueUsername(t *testing.T) {
	p1 := GenerateProfile(0, "nit")
	p2 := GenerateProfile(1, "nit")
	assert.NotEqual(t, p1.Username, p2.Username)
}
