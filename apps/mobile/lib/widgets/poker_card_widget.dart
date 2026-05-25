import 'package:flutter/material.dart';
import '../models/card.dart';
import '../theme.dart';

class PokerCardWidget extends StatelessWidget {
  final PokerCard? card;
  final bool faceDown;
  final double width;
  final double height;

  const PokerCardWidget({
    super.key,
    this.card,
    this.faceDown = false,
    this.width = 24,
    this.height = 34,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        gradient: faceDown
            ? const LinearGradient(
                colors: [Color(0xFF8B1A1A), Color(0xFF5a0f0f)],
              )
            : const LinearGradient(
                colors: [Colors.white, Color(0xFFF5F5F5)],
              ),
        borderRadius: BorderRadius.circular(4),
        border: Border.all(
          color: faceDown ? AppColors.gold : AppColors.goldBorder,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.3),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: faceDown || card == null
          ? Center(
              child: Container(
                width: width * 0.65,
                height: height * 0.7,
                decoration: BoxDecoration(
                  border: Border.all(
                    color: AppColors.gold.withOpacity(0.3),
                    style: BorderStyle.solid,
                  ),
                  borderRadius: BorderRadius.circular(3),
                ),
                child: Center(
                  child: Text(
                    '♠',
                    style: TextStyle(
                      fontSize: width * 0.35,
                      color: AppColors.gold.withOpacity(0.4),
                    ),
                  ),
                ),
              ),
            )
          : Stack(
              children: [
                // Top-left rank + suit
                Positioned(
                  top: 2,
                  left: 2,
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        _rankLabel(card!.rank),
                        style: TextStyle(
                          fontSize: width * 0.35,
                          fontWeight: FontWeight.bold,
                          color: card!.color,
                          height: 1,
                        ),
                      ),
                      Text(
                        _suitSymbol(card!.suit),
                        style: TextStyle(
                          fontSize: width * 0.28,
                          color: card!.color,
                          height: 1,
                        ),
                      ),
                    ],
                  ),
                ),
                // Center large suit
                Positioned(
                  top: height * 0.55,
                  left: width * 0.5,
                  child: Transform.translate(
                    offset: const Offset(-0.5, -0.5),
                    child: Text(
                      _suitSymbol(card!.suit),
                      style: TextStyle(
                        fontSize: width * 0.6,
                        color: card!.color,
                      ),
                    ),
                  ),
                ),
              ],
            ),
    );
  }

  String _rankLabel(int rank) {
    const ranks = ['', '', '2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K', 'A'];
    return ranks[rank];
  }

  String _suitSymbol(int suit) {
    const suits = ['', '♠', '♥', '♦', '♣'];
    return suits[suit];
  }
}
