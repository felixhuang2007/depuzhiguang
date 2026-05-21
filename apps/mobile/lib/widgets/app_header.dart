import 'package:flutter/material.dart';
import '../theme.dart';

class AppHeader extends StatelessWidget implements PreferredSizeWidget {
  final String? title;
  const AppHeader({super.key, this.title});

  @override
  Size get preferredSize => const Size.fromHeight(48);

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        color: AppColors.header,
        border: Border(
          bottom: BorderSide(color: AppColors.goldBorder, width: 1),
        ),
      ),
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  const Text(
                    '♠',
                    style: TextStyle(
                      fontSize: 22,
                      color: AppColors.goldBright,
                      shadows: [
                        Shadow(color: AppColors.gold, blurRadius: 6),
                      ],
                    ),
                  ),
                  if (title != null) ...[
                    const SizedBox(width: 8),
                    Text(
                      title!,
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: AppColors.goldBright,
                      ),
                    ),
                  ],
                ],
              ),
              Row(
                children: [
                  const Text(
                    '💰 2,450',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: AppColors.goldBright,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Container(
                    width: 30,
                    height: 30,
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [AppColors.gold, AppColors.goldMuted],
                      ),
                      borderRadius: BorderRadius.circular(15),
                      border: Border.all(color: AppColors.goldBright, width: 2),
                    ),
                    child: const Icon(Icons.person, size: 16, color: AppColors.bg),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
