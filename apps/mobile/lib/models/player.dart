import 'card.dart';

class Player {
  final String id;
  final String name;
  final double chips;
  final int? seat;
  final bool isDealer;
  final bool isActive;
  final bool hasFolded;
  final bool isAllIn;
  final bool isAway;
  final String? statusTag;
  final List<PokerCard>? holeCards;
  final String? avatar;
  final bool isHero;

  const Player({
    required this.id,
    required this.name,
    required this.chips,
    this.seat,
    this.isDealer = false,
    this.isActive = false,
    this.hasFolded = false,
    this.isAllIn = false,
    this.isAway = false,
    this.statusTag,
    this.holeCards,
    this.avatar,
    this.isHero = false,
  });
}
