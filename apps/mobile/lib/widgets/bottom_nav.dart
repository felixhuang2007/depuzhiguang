import 'package:flutter/material.dart';
import '../theme.dart';

class BottomNav extends StatelessWidget {
  final int currentIndex;
  final ValueChanged<int> onTap;

  const BottomNav({super.key, required this.currentIndex, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final items = [
      _NavItem(icon: Icons.home, label: '大厅'),
      _NavItem(icon: Icons.chat_bubble_outline, label: '社交'),
      _NavItem(icon: Icons.emoji_events, label: '排行'),
      _NavItem(icon: Icons.person_outline, label: '我的'),
    ];

    return Container(
      decoration: const BoxDecoration(
        color: AppColors.header,
        border: Border(
          top: BorderSide(color: AppColors.goldBorder, width: 1),
        ),
      ),
      child: SafeArea(
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: items.asMap().entries.map((e) {
            final i = e.key;
            final item = e.value;
            final active = i == currentIndex;
            return GestureDetector(
              onTap: () => onTap(i),
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 6),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      item.icon,
                      size: 22,
                      color: active ? AppColors.goldBright : AppColors.textMuted,
                    ),
                    const SizedBox(height: 2),
                    Text(
                      item.label,
                      style: TextStyle(
                        fontSize: 9,
                        color: active ? AppColors.goldBright : AppColors.textMuted,
                        fontWeight: active ? FontWeight.bold : FontWeight.normal,
                      ),
                    ),
                  ],
                ),
              ),
            );
          }).toList(),
        ),
      ),
    );
  }
}

class _NavItem {
  final IconData icon;
  final String label;
  _NavItem({required this.icon, required this.label});
}
