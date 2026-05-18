package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newCard(s Suit, r Rank) Card { return NewCard(s, r) }

func TestEvaluate_RoyalFlush(t *testing.T) {
	hole := []Card{newCard(Spades, Ace), newCard(Spades, King)}
	board := []Card{newCard(Spades, Queen), newCard(Spades, Jack), newCard(Spades, Ten), newCard(Hearts, Five), newCard(Diamonds, Nine)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, RoyalFlush, rank)
	assert.Equal(t, RoyalFlush, result.Category)
	assert.Equal(t, []Rank{Ace}, result.Kickers)
}

func TestEvaluate_StraightFlush(t *testing.T) {
	hole := []Card{newCard(Hearts, Nine), newCard(Hearts, Eight)}
	board := []Card{newCard(Hearts, Seven), newCard(Hearts, Six), newCard(Hearts, Five), newCard(Spades, Ace), newCard(Diamonds, King)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, StraightFlush, rank)
	assert.Equal(t, StraightFlush, result.Category)
	assert.Equal(t, []Rank{Nine}, result.Kickers)
}

func TestEvaluate_FourOfAKind(t *testing.T) {
	hole := []Card{newCard(Spades, Ace), newCard(Hearts, Ace)}
	board := []Card{newCard(Diamonds, Ace), newCard(Clubs, Ace), newCard(Spades, King), newCard(Hearts, Five), newCard(Diamonds, Nine)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, FourOfAKind, rank)
	assert.Equal(t, FourOfAKind, result.Category)
	assert.Equal(t, []Rank{Ace, King}, result.Kickers)
}

func TestEvaluate_FullHouse(t *testing.T) {
	hole := []Card{newCard(Spades, Ace), newCard(Hearts, Ace)}
	board := []Card{newCard(Diamonds, Ace), newCard(Clubs, King), newCard(Spades, King), newCard(Hearts, Five), newCard(Diamonds, Nine)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, FullHouse, rank)
	assert.Equal(t, FullHouse, result.Category)
	assert.Equal(t, []Rank{Ace, King}, result.Kickers)
}

func TestEvaluate_Flush(t *testing.T) {
	hole := []Card{newCard(Spades, Ace), newCard(Spades, King)}
	board := []Card{newCard(Spades, Nine), newCard(Spades, Seven), newCard(Spades, Two), newCard(Hearts, Five), newCard(Diamonds, Nine)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, Flush, rank)
	assert.Equal(t, Flush, result.Category)
	assert.Equal(t, []Rank{Ace, King, Nine, Seven, Two}, result.Kickers)
}

func TestEvaluate_Straight(t *testing.T) {
	hole := []Card{newCard(Spades, Nine), newCard(Hearts, Eight)}
	board := []Card{newCard(Diamonds, Seven), newCard(Clubs, Six), newCard(Spades, Five), newCard(Hearts, Ace), newCard(Diamonds, King)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, Straight, rank)
	assert.Equal(t, Straight, result.Category)
	assert.Equal(t, []Rank{Nine}, result.Kickers)
}

func TestEvaluate_Straight_AceLow(t *testing.T) {
	// Wheel: A-2-3-4-5
	hole := []Card{newCard(Spades, Ace), newCard(Hearts, Two)}
	board := []Card{newCard(Diamonds, Three), newCard(Clubs, Four), newCard(Spades, Five), newCard(Hearts, King), newCard(Diamonds, Queen)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, Straight, rank)
	assert.Equal(t, Straight, result.Category)
	assert.Equal(t, []Rank{Five}, result.Kickers)
}

func TestEvaluate_ThreeOfAKind(t *testing.T) {
	hole := []Card{newCard(Spades, Ace), newCard(Hearts, Ace)}
	board := []Card{newCard(Diamonds, Ace), newCard(Clubs, King), newCard(Spades, Queen), newCard(Hearts, Five), newCard(Diamonds, Nine)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, ThreeOfAKind, rank)
	assert.Equal(t, ThreeOfAKind, result.Category)
	assert.Equal(t, []Rank{Ace, King, Queen}, result.Kickers)
}

func TestEvaluate_TwoPair(t *testing.T) {
	hole := []Card{newCard(Spades, Ace), newCard(Hearts, Ace)}
	board := []Card{newCard(Diamonds, King), newCard(Clubs, King), newCard(Spades, Queen), newCard(Hearts, Five), newCard(Diamonds, Nine)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, TwoPair, rank)
	assert.Equal(t, TwoPair, result.Category)
	assert.Equal(t, []Rank{Ace, King, Queen}, result.Kickers)
}

func TestEvaluate_OnePair(t *testing.T) {
	hole := []Card{newCard(Spades, Ace), newCard(Hearts, Ace)}
	board := []Card{newCard(Diamonds, King), newCard(Clubs, Queen), newCard(Spades, Jack), newCard(Hearts, Five), newCard(Diamonds, Nine)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, OnePair, rank)
	assert.Equal(t, OnePair, result.Category)
	assert.Equal(t, []Rank{Ace, King, Queen, Jack}, result.Kickers)
}

func TestEvaluate_HighCard(t *testing.T) {
	hole := []Card{newCard(Spades, Ace), newCard(Hearts, King)}
	board := []Card{newCard(Diamonds, Queen), newCard(Clubs, Jack), newCard(Spades, Seven), newCard(Hearts, Five), newCard(Diamonds, Two)}

	rank, result := EvaluateBest(hole, board)
	assert.Equal(t, HighCard, rank)
	assert.Equal(t, HighCard, result.Category)
	assert.Equal(t, []Rank{Ace, King, Queen, Jack, Seven}, result.Kickers)
}

func TestEvaluate_CompareSameRank(t *testing.T) {
	// Two one-pair hands: pair of Aces with different kickers
	hole1 := []Card{newCard(Spades, Ace), newCard(Hearts, Ace)}
	board1 := []Card{newCard(Diamonds, King), newCard(Clubs, Queen), newCard(Spades, Jack), newCard(Hearts, Five), newCard(Diamonds, Nine)}

	hole2 := []Card{newCard(Spades, Ace), newCard(Hearts, Ace)}
	board2 := []Card{newCard(Diamonds, King), newCard(Clubs, Queen), newCard(Spades, Ten), newCard(Hearts, Five), newCard(Diamonds, Nine)}

	_, result1 := EvaluateBest(hole1, board1)
	_, result2 := EvaluateBest(hole2, board2)

	assert.Equal(t, OnePair, result1.Category)
	assert.Equal(t, OnePair, result2.Category)

	// result1 should be better because Jack > Ten
	assert.True(t, isBetterKicker(result1.Kickers, result2.Kickers))
}

func BenchmarkEvaluateBest(b *testing.B) {
	hole := []Card{newCard(Spades, Ace), newCard(Spades, King)}
	board := []Card{newCard(Spades, Queen), newCard(Spades, Seven), newCard(Spades, Two), newCard(Hearts, Five), newCard(Diamonds, Nine)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvaluateBest(hole, board)
	}
}

