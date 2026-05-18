package engine

import "sort"

// EvalResult contains the evaluated hand category and kickers for tie-breaking.
type EvalResult struct {
	Category HandRank
	Kickers  []Rank // highest first, for tie-breaking
}

// EvaluateBest evaluates the best 5-card hand from hole cards and board cards.
// It generates all C(n,5) combinations and returns the highest-ranking hand.
func EvaluateBest(hole []Card, board []Card) (HandRank, EvalResult) {
	all := make([]Card, 0, len(hole)+len(board))
	all = append(all, hole...)
	all = append(all, board...)

	if len(all) < 5 {
		return HighCard, EvalResult{Category: HighCard, Kickers: []Rank{}}
	}

	var bestResult EvalResult
	var bestRank HandRank = HighCard

	// Generate all C(n,5) combinations using 5 nested indices
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
	// Extract and sort ranks descending
	ranks := make([]Rank, 5)
	suits := make([]Suit, 5)
	for i, c := range cards {
		ranks[i] = c.Rank()
		suits[i] = c.Suit()
	}
	sort.Slice(ranks, func(i, j int) bool { return ranks[i] > ranks[j] })

	// Check flush
	isFlush := true
	for i := 1; i < 5; i++ {
		if suits[i] != suits[0] {
			isFlush = false
			break
		}
	}

	// Check straight (including ace-low wheel)
	isStraight := true
	for i := 1; i < 5; i++ {
		if ranks[i] != ranks[i-1]-1 {
			isStraight = false
			break
		}
	}

	// Ace-low straight: A-2-3-4-5
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

	// Count rank frequencies
	freq := make(map[Rank]int)
	for _, r := range ranks {
		freq[r]++
	}

	// Build frequency groups: map from count -> list of ranks with that count
	byCount := make(map[int][]Rank)
	for r, c := range freq {
		byCount[c] = append(byCount[c], r)
	}
	// Sort each group descending
	for _, group := range byCount {
		sort.Slice(group, func(i, j int) bool { return group[i] > group[j] })
	}

	// Four of a Kind
	if len(byCount[4]) > 0 {
		kicker := byCount[1][0]
		return FourOfAKind, EvalResult{Category: FourOfAKind, Kickers: []Rank{byCount[4][0], kicker}}
	}

	// Full House
	if len(byCount[3]) > 0 && len(byCount[2]) > 0 {
		return FullHouse, EvalResult{Category: FullHouse, Kickers: []Rank{byCount[3][0], byCount[2][0]}}
	}

	// Flush
	if isFlush {
		return Flush, EvalResult{Category: Flush, Kickers: append([]Rank(nil), ranks...)}
	}

	// Straight (normal)
	if isStraight {
		return Straight, EvalResult{Category: Straight, Kickers: []Rank{ranks[0]}}
	}

	// Ace-low straight (wheel)
	if isWheel {
		return Straight, EvalResult{Category: Straight, Kickers: []Rank{Five}}
	}

	// Three of a Kind
	if len(byCount[3]) > 0 {
		kickers := make([]Rank, 0, 2)
		kickers = append(kickers, byCount[3][0])
		// Add remaining single cards in descending order
		for _, r := range ranks {
			if freq[r] == 1 {
				kickers = append(kickers, r)
			}
		}
		return ThreeOfAKind, EvalResult{Category: ThreeOfAKind, Kickers: kickers}
	}

	// Two Pair
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

	// One Pair
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

	// High Card
	return HighCard, EvalResult{Category: HighCard, Kickers: append([]Rank(nil), ranks...)}
}

// isBetterKicker returns true if k1 is lexicographically greater than k2.
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
