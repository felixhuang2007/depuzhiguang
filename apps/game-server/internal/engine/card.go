package engine

// No imports needed

// Suit represents a card suit
type Suit uint8

const (
	InvalidSuit Suit = iota
	Spades
	Hearts
	Diamonds
	Clubs
)

func (s Suit) String() string {
	switch s {
	case Spades:
		return "♠"
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	default:
		return "?"
	}
}

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

func (r Rank) String() string {
	switch r {
	case Two:
		return "2"
	case Three:
		return "3"
	case Four:
		return "4"
	case Five:
		return "5"
	case Six:
		return "6"
	case Seven:
		return "7"
	case Eight:
		return "8"
	case Nine:
		return "9"
	case Ten:
		return "T"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	case Ace:
		return "A"
	default:
		return "?"
	}
}

// Card is a playing card with suit and rank
type Card struct {
	suit Suit
	rank Rank
}

// NewCard creates a new card. Validates inputs.
func NewCard(suit Suit, rank Rank) Card {
	if suit < Spades || suit > Clubs {
		suit = InvalidSuit
	}
	if rank < Two || rank > Ace {
		rank = InvalidRank
	}
	return Card{suit: suit, rank: rank}
}

func (c Card) Suit() Suit   { return c.suit }
func (c Card) Rank() Rank   { return c.rank }
func (c Card) String() string { return c.rank.String() + c.suit.String() }

// ToByte returns a compact byte representation for wire protocol
func (c Card) ToByte() byte { return byte(c.suit)<<4 | byte(c.rank) }

// CardFromByte reconstructs a Card from a compact byte representation
func CardFromByte(b byte) Card { return NewCard(Suit(b>>4), Rank(b&0x0F)) }
