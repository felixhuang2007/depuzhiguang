import 'package:flutter/material.dart';
import '../theme.dart';

class PlayerAvatar extends StatelessWidget {
  final String? emoji;
  final bool isActive;
  final bool isDealer;
  final String? timerText;
  final double size;

  const PlayerAvatar({
    super.key,
    this.emoji = '👤',
    this.isActive = false,
    this.isDealer = false,
    this.timerText,
    this.size = 32,
  });

  @override
  Widget build(BuildContext context) {
    final borderWidth = isActive ? 2.0 : 1.0;
    final borderColor = isActive ? AppColors.goldBright : AppColors.gold.withOpacity(0.4);

    return Stack(
      clipBehavior: Clip.none,
      children: [
        Container(
          padding: const EdgeInsets.all(2),
          decoration: BoxDecoration(
            color: AppColors.surface.withOpacity(0.7),
            borderRadius: BorderRadius.circular(6),
            border: Border.all(color: borderColor, width: borderWidth),
            boxShadow: isActive
                ? [
                    BoxShadow(
                      color: AppColors.goldBright.withOpacity(0.3),
                      blurRadius: 8,
                    ),
                  ]
                : null,
          ),
          child: Container(
            width: size,
            height: size,
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [AppColors.gold, AppColors.goldMuted],
              ),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Center(
              child: Text(
                emoji ?? '👤',
                style: TextStyle(fontSize: size * 0.5),
              ),
            ),
          ),
        ),
        if (timerText != null)
          Positioned(
            top: -10,
            left: 0,
            right: 0,
            child: Center(
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                decoration: BoxDecoration(
                  color: AppColors.surface.withOpacity(0.9),
                  borderRadius: BorderRadius.circular(6),
                  border: Border.all(color: AppColors.goldBorder),
                ),
                child: Text(
                  timerText!,
                  style: const TextStyle(fontSize: 7, color: AppColors.goldBright),
                ),
              ),
            ),
          ),
        if (isDealer)
          Positioned(
            bottom: -6,
            right: -4,
            child: Container(
              width: 14,
              height: 14,
              decoration: BoxDecoration(
                color: AppColors.goldBright,
                borderRadius: BorderRadius.circular(7),
                border: Border.all(color: AppColors.goldBorder),
              ),
              child: const Center(
                child: Text(
                  'D',
                  style: TextStyle(
                    fontSize: 8,
                    color: AppColors.bg,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }
}
