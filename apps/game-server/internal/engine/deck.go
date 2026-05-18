package engine

import (
	"crypto/rand"
	"math/big"
)

// Deck represents a standard 52-card playing deck.
type Deck struct {
	cards []Card
	next  int
}

// NewDeck creates a new unshuffled 52-card deck.
// Cards are ordered by suit (Spades, Hearts, Diamonds, Clubs)
// and within each suit by rank (Two through Ace).
func NewDeck() *Deck {
	cards := make([]Card, 0, 52)
	for s := Spades; s <= Clubs; s++ {
		for r := Two; r <= Ace; r++ {
			cards = append(cards, NewCard(s, r))
		}
	}
	return &Deck{cards: cards}
}

// Len returns the number of remaining cards in the deck.
func (d *Deck) Len() int {
	return len(d.cards) - d.next
}

// Deal removes and returns the top card from the deck.
// The second return value is false if the deck is empty.
func (d *Deck) Deal() (Card, bool) {
	if d.next >= len(d.cards) {
		return Card{}, false
	}
	c := d.cards[d.next]
	d.next++
	return c, true
}

// Shuffle randomizes the order of the remaining cards in the deck
// using a Fisher-Yates shuffle with crypto-secure randomness.
func (d *Deck) Shuffle() {
	n := len(d.cards) - d.next
	if n <= 1 {
		return
	}
	for i := n - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			panic(err)
		}
		j := int(jBig.Int64())
		d.cards[d.next+i], d.cards[d.next+j] = d.cards[d.next+j], d.cards[d.next+i]
	}
}

// Reset restores the deck to its initial unshuffled state.
func (d *Deck) Reset() {
	d.next = 0
}
