package ai

type Persona struct {
	Style      string
	VPIPTarget float64
	PFRTarget  float64
	Aggression float64
	BluffRate  float64
	TiltFactor float64
	Patience   float64
}

func GetPersona(style string) Persona {
	switch style {
	case "tight_aggressive":
		return Persona{Style: "tight_aggressive", VPIPTarget: 0.18, PFRTarget: 0.15, Aggression: 0.75, BluffRate: 0.25, TiltFactor: 0.3, Patience: 0.8}
	case "loose_aggressive":
		return Persona{Style: "loose_aggressive", VPIPTarget: 0.32, PFRTarget: 0.25, Aggression: 0.85, BluffRate: 0.40, TiltFactor: 0.4, Patience: 0.4}
	case "nit":
		return Persona{Style: "nit", VPIPTarget: 0.10, PFRTarget: 0.08, Aggression: 0.60, BluffRate: 0.10, TiltFactor: 0.1, Patience: 0.95}
	case "loose_passive":
		return Persona{Style: "loose_passive", VPIPTarget: 0.35, PFRTarget: 0.07, Aggression: 0.20, BluffRate: 0.15, TiltFactor: 0.2, Patience: 0.5}
	case "maniac":
		return Persona{Style: "maniac", VPIPTarget: 0.55, PFRTarget: 0.45, Aggression: 0.95, BluffRate: 0.55, TiltFactor: 0.7, Patience: 0.1}
	case "rock":
		return Persona{Style: "rock", VPIPTarget: 0.12, PFRTarget: 0.10, Aggression: 0.50, BluffRate: 0.05, TiltFactor: 0.1, Patience: 0.9}
	case "calling_station":
		return Persona{Style: "calling_station", VPIPTarget: 0.45, PFRTarget: 0.03, Aggression: 0.10, BluffRate: 0.05, TiltFactor: 0.1, Patience: 0.3}
	case "adaptive":
		return Persona{Style: "adaptive", VPIPTarget: 0.25, PFRTarget: 0.18, Aggression: 0.70, BluffRate: 0.30, TiltFactor: 0.3, Patience: 0.6}
	default:
		return Persona{Style: "regular", VPIPTarget: 0.22, PFRTarget: 0.16, Aggression: 0.60, BluffRate: 0.20, TiltFactor: 0.2, Patience: 0.5}
	}
}

func AllPersonas() []string {
	return []string{
		"tight_aggressive", "loose_aggressive", "nit", "loose_passive",
		"maniac", "rock", "calling_station", "adaptive",
	}
}
