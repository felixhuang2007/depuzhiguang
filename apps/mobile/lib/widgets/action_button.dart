import 'package:flutter/material.dart';
import '../theme.dart';

class ActionButton extends StatelessWidget {
  final String label;
  final IconData? icon;
  final String? text;
  final Color bgColor;
  final double size;
  final VoidCallback onTap;

  const ActionButton({
    super.key,
    required this.label,
    this.icon,
    this.text,
    required this.bgColor,
    required this.size,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: size,
            height: size,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [
                  bgColor.withOpacity(0.9),
                  bgColor.withOpacity(0.6),
                ],
              ),
              shape: BoxShape.circle,
              border: Border.all(color: AppColors.gold, width: 2),
              boxShadow: [
                BoxShadow(
                  color: bgColor.withOpacity(0.3),
                  blurRadius: 8,
                ),
              ],
            ),
            child: Center(
              child: text != null
                  ? Text(
                      text!,
                      style: const TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                        color: AppColors.goldBright,
                      ),
                    )
                  : Icon(icon, color: Colors.white, size: size * 0.4),
            ),
          ),
          const SizedBox(height: 2),
          Text(
            label,
            style: const TextStyle(fontSize: 8, color: AppColors.goldBright),
          ),
        ],
      ),
    );
  }
}
