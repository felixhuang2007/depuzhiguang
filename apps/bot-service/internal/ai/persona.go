package ai

// Persona defines a poker playing style with statistical targets.
type Persona struct {
	// Style is the persona identifier (e.g., "tight_aggressive").
	Style string
	// VPIPTarget is the target Voluntarily Put Money In Pot percentage (0.0–1.0).
	VPIPTarget float64
	// PFRTarget is the target Pre-Flop Raise percentage (0.0–1.0).
	PFRTarget float64
	// Aggression measures how aggressively the player bets/raises (0.0–1.0).
	Aggression float64
	// BluffRate is the frequency of bluffing (0.0–1.0).
	BluffRate float64
	// TiltFactor increases aggression after consecutive losses (0.0–1.0).
	TiltFactor float64
	// Patience affects willingness to wait for strong hands (0.0–1.0).
	Patience float64
}

var personas = map[string]Persona{
	"tight_aggressive": {Style: "tight_aggressive", VPIPTarget: 0.18, PFRTarget: 0.15, Aggression: 0.75, BluffRate: 0.25, TiltFactor: 0.3, Patience: 0.8},
	"loose_aggressive": {Style: "loose_aggressive", VPIPTarget: 0.32, PFRTarget: 0.25, Aggression: 0.85, BluffRate: 0.40, TiltFactor: 0.4, Patience: 0.4},
	"nit":              {Style: "nit", VPIPTarget: 0.10, PFRTarget: 0.08, Aggression: 0.60, BluffRate: 0.10, TiltFactor: 0.1, Patience: 0.95},
	"loose_passive":    {Style: "loose_passive", VPIPTarget: 0.35, PFRTarget: 0.07, Aggression: 0.20, BluffRate: 0.15, TiltFactor: 0.2, Patience: 0.5},
	"maniac":           {Style: "maniac", VPIPTarget: 0.55, PFRTarget: 0.45, Aggression: 0.95, BluffRate: 0.55, TiltFactor: 0.7, Patience: 0.1},
	"rock":             {Style: "rock", VPIPTarget: 0.12, PFRTarget: 0.10, Aggression: 0.50, BluffRate: 0.05, TiltFactor: 0.1, Patience: 0.9},
	"calling_station":  {Style: "calling_station", VPIPTarget: 0.45, PFRTarget: 0.03, Aggression: 0.10, BluffRate: 0.05, TiltFactor: 0.1, Patience: 0.3},
	"adaptive":         {Style: "adaptive", VPIPTarget: 0.25, PFRTarget: 0.18, Aggression: 0.70, BluffRate: 0.30, TiltFactor: 0.3, Patience: 0.6},
}

var defaultPersona = Persona{Style: "regular", VPIPTarget: 0.22, PFRTarget: 0.16, Aggression: 0.60, BluffRate: 0.20, TiltFactor: 0.2, Patience: 0.5}

// GetPersona returns the persona for the given style, or "regular" if unknown.
func GetPersona(style string) Persona {
	if p, ok := personas[style]; ok {
		return p
	}
	return defaultPersona
}

// AllPersonas returns all available persona style names.
func AllPersonas() []string {
	names := make([]string, 0, len(personas))
	for name := range personas {
		names = append(names, name)
	}
	return names
}
