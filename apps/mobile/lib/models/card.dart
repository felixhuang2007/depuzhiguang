import 'package:flutter/material.dart';

class PokerCard {
  final int suit;
  final int rank;

  const PokerCard(this.suit, this.rank);

  String get display {
    const ranks = ['', '', '2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K', 'A'];
    const suits = ['', '♠', '♥', '♦', '♣'];
    return '${ranks[rank]}${suits[suit]}';
  }

  Color get color => (suit == 2 || suit == 3) ? Colors.red : Colors.black;
}
