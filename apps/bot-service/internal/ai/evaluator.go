package ai

import "sort"

// Suit represents a card suit
type Suit uint8

const (
	InvalidSuit Suit = iota
	Spades
	Hearts
	Diamonds
	Clubs
)

// Rank represents a card rank (2-Ace)
type Rank uint8

const (
	InvalidRank Rank = iota
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

// Card is a playing card with suit and rank
type Card struct {
	Suit Suit
	Rank Rank
}

// HandRank represents the strength category of a 5-card poker hand
type HandRank uint8

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

// EvalResult contains the evaluated hand category and kickers for tie-breaking.
type EvalResult struct {
	Category HandRank
	Kickers  []Rank // highest first, for tie-breaking
}

// parseCard converts a string like "AS" or "10H" into a Card.
func parseCard(s string) Card {
	if len(s) < 2 {
		return Card{}
	}
	var rankStr string
	var suitChar byte
	if len(s) == 3 && (s[0] == '1' && s[1] == '0') {
		rankStr = "T"
		suitChar = s[2]
	} else {
		rankStr = string(s[0])
		suitChar = s[1]
	}

	var r Rank
	switch rankStr {
	case "2":
		r = Two
	case "3":
		r = Three
	case "4":
		r = Four
	case "5":
		r = Five
	case "6":
		r = Six
	case "7":
		r = Seven
	case "8":
		r = Eight
	case "9":
		r = Nine
	case "T", "10":
		r = Ten
	case "J":
		r = Jack
	case "Q":
		r = Queen
	case "K":
		r = King
	case "A":
		r = Ace
	}

	var su Suit
	switch suitChar {
	case 'S', 's':
		su = Spades
	case 'H', 'h':
		su = Hearts
	case 'D', 'd':
		su = Diamonds
	case 'C', 'c':
		su = Clubs
	}

	return Card{Suit: su, Rank: r}
}

// parseCards converts a slice of card strings to Cards.
func parseCards(cards []string) []Card {
	result := make([]Card, 0, len(cards))
	for _, c := range cards {
		if c != "" && c != "?" {
			result = append(result, parseCard(c))
		}
	}
	return result
}

// EvaluateBest evaluates the best 5-card hand from hole cards and board cards.
func EvaluateBest(hole []string, board []string) (HandRank, EvalResult) {
	all := parseCards(append(hole, board...))
	if len(all) < 5 {
		return HighCard, EvalResult{Category: HighCard, Kickers: []Rank{}}
	}

	var bestResult EvalResult
	var bestRank HandRank = HighCard

	n := len(all)
	for a := 0; a < n; a++ {
		for b := a + 1; b < n; b++ {
			for c := b + 1; c < n; c++ {
				for d := c + 1; d < n; d++ {
					for e := d + 1; e < n; e++ {
						combo := []Card{all[a], all[b], all[c], all[d], all[e]}
						rank, result := evaluate5(combo)
						if rank > bestRank || (rank == bestRank && isBetterKicker(result.Kickers, bestResult.Kickers)) {
							bestRank = rank
							bestResult = result
						}
					}
				}
			}
		}
	}

	return bestRank, bestResult
}

// evaluate5 evaluates a single 5-card hand and returns its rank and kickers.
func evaluate5(cards []Card) (HandRank, EvalResult) {
	ranks := make([]Rank, 5)
	suits := make([]Suit, 5)
	for i, c := range cards {
		ranks[i] = c.Rank
		suits[i] = c.Suit
	}
	sort.Slice(ranks, func(i, j int) bool { return ranks[i] > ranks[j] })

	isFlush := true
	for i := 1; i < 5; i++ {
		if suits[i] != suits[0] {
			isFlush = false
			break
		}
	}

	isStraight := true
	for i := 1; i < 5; i++ {
		if ranks[i] != ranks[i-1]-1 {
			isStraight = false
			break
		}
	}

	isWheel := ranks[0] == Ace && ranks[1] == Five && ranks[2] == Four && ranks[3] == Three && ranks[4] == Two

	if isStraight && isFlush {
		if ranks[0] == Ace && ranks[1] == King {
			return RoyalFlush, EvalResult{Category: RoyalFlush, Kickers: []Rank{Ace}}
		}
		return StraightFlush, EvalResult{Category: StraightFlush, Kickers: []Rank{ranks[0]}}
	}

	if isWheel && isFlush {
		return StraightFlush, EvalResult{Category: StraightFlush, Kickers: []Rank{Five}}
	}

	freq := make(map[Rank]int)
	for _, r := range ranks {
		freq[r]++
	}

	byCount := make(map[int][]Rank)
	for r, c := range freq {
		byCount[c] = append(byCount[c], r)
	}
	for _, group := range byCount {
		sort.Slice(group, func(i, j int) bool { return group[i] > group[j] })
	}

	if len(byCount[4]) > 0 {
		kicker := byCount[1][0]
		return FourOfAKind, EvalResult{Category: FourOfAKind, Kickers: []Rank{byCount[4][0], kicker}}
	}

	if len(byCount[3]) > 0 && len(byCount[2]) > 0 {
		return FullHouse, EvalResult{Category: FullHouse, Kickers: []Rank{byCount[3][0], byCount[2][0]}}
	}

	if isFlush {
		return Flush, EvalResult{Category: Flush, Kickers: append([]Rank(nil), ranks...)}
	}

	if isStraight {
		return Straight, EvalResult{Category: Straight, Kickers: []Rank{ranks[0]}}
	}

	if isWheel {
		return Straight, EvalResult{Category: Straight, Kickers: []Rank{Five}}
	}

	if len(byCount[3]) > 0 {
		kickers := make([]Rank, 0, 2)
		kickers = append(kickers, byCount[3][0])
		for _, r := range ranks {
			if freq[r] == 1 {
				kickers = append(kickers, r)
			}
		}
		return ThreeOfAKind, EvalResult{Category: ThreeOfAKind, Kickers: kickers}
	}

	if len(byCount[2]) == 2 {
		kickers := make([]Rank, 0, 3)
		kickers = append(kickers, byCount[2]...)
		for _, r := range ranks {
			if freq[r] == 1 {
				kickers = append(kickers, r)
				break
			}
		}
		return TwoPair, EvalResult{Category: TwoPair, Kickers: kickers}
	}

	if len(byCount[2]) == 1 {
		kickers := make([]Rank, 0, 4)
		kickers = append(kickers, byCount[2][0])
		for _, r := range ranks {
			if freq[r] == 1 {
				kickers = append(kickers, r)
			}
		}
		return OnePair, EvalResult{Category: OnePair, Kickers: kickers}
	}

	return HighCard, EvalResult{Category: HighCard, Kickers: append([]Rank(nil), ranks...)}
}

func isBetterKicker(k1, k2 []Rank) bool {
	for i := 0; i < len(k1) && i < len(k2); i++ {
		if k1[i] > k2[i] {
			return true
		}
		if k1[i] < k2[i] {
			return false
		}
	}
	return len(k1) > len(k2)
}

// handRankToStrength maps a hand rank to a base strength value (0.0 - 1.0).
func handRankToStrength(rank HandRank, kickers []Rank) float64 {
	base := map[HandRank]float64{
		HighCard:      0.00,
		OnePair:       0.20,
		TwoPair:       0.40,
		ThreeOfAKind:  0.55,
		Straight:      0.70,
		Flush:         0.75,
		FullHouse:     0.85,
		FourOfAKind:   0.93,
		StraightFlush: 0.97,
		RoyalFlush:    1.00,
	}[rank]

	// Add kicker bonus (0.0 - 0.05) for finer gradation within rank
	kickerBonus := 0.0
	if len(kickers) > 0 {
		kickerBonus = float64(kickers[0]) / float64(Ace) * 0.05
	}
	return base + kickerBonus
}

// hasFlushDraw returns true if the player has 4 cards of the same suit.
func hasFlushDraw(cards []Card) bool {
	suitCount := make(map[Suit]int)
	for _, c := range cards {
		suitCount[c.Suit]++
	}
	for _, count := range suitCount {
		if count >= 4 {
			return true
		}
	}
	return false
}

// hasStraightDraw returns true if the player has 4 cards in sequential order (open-ended or gutshot).
func hasStraightDraw(cards []Card) bool {
	if len(cards) < 4 {
		return false
	}
	ranks := make([]Rank, len(cards))
	for i, c := range cards {
		ranks[i] = c.Rank
	}
	sort.Slice(ranks, func(i, j int) bool { return ranks[i] > ranks[j] })

	// Remove duplicates
	unique := make([]Rank, 0, len(ranks))
	seen := make(map[Rank]bool)
	for _, r := range ranks {
		if !seen[r] {
			seen[r] = true
			unique = append(unique, r)
		}
	}

	if len(unique) < 4 {
		return false
	}

	for i := 0; i <= len(unique)-4; i++ {
		if unique[i]-unique[i+3] <= 4 {
			return true
		}
	}

	// Ace-low wheel draw: A,2,3,4 or A,2,3,5 or A,2,4,5 or A,3,4,5
	if seen[Ace] {
		lowCount := 0
		for r := Two; r <= Five; r++ {
			if seen[r] {
				lowCount++
			}
		}
		if lowCount >= 3 {
			return true
		}
	}

	return false
}

// isOpenEndedStraightDraw returns true if there are 4 sequential cards with gaps on both ends.
func isOpenEndedStraightDraw(cards []Card) bool {
	if len(cards) < 4 {
		return false
	}
	ranks := make([]Rank, len(cards))
	for i, c := range cards {
		ranks[i] = c.Rank
	}
	sort.Slice(ranks, func(i, j int) bool { return ranks[i] > ranks[j] })

	unique := make([]Rank, 0, len(ranks))
	seen := make(map[Rank]bool)
	for _, r := range ranks {
		if !seen[r] {
			seen[r] = true
			unique = append(unique, r)
		}
	}

	if len(unique) < 4 {
		return false
	}

	for i := 0; i <= len(unique)-4; i++ {
		if unique[i]-unique[i+3] == 3 {
			return true
		}
	}
	return false
}
