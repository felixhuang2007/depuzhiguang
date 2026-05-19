import 'package:flutter/material.dart';
import '../models/card.dart';

class CardWidget extends StatelessWidget {
  final PokerCard? card;
  final bool faceDown;

  const CardWidget({super.key, this.card, this.faceDown = false});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 50,
      height: 70,
      decoration: BoxDecoration(
        color: faceDown ? Colors.blue.shade800 : Colors.white,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: Colors.black26),
      ),
      child: faceDown || card == null
          ? null
          : Center(
              child: Text(
                card!.display,
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: card!.color,
                ),
              ),
            ),
    );
  }
}
