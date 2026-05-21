import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import '../theme.dart';

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        // Avatar + stats
        Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                AppColors.gold.withOpacity(0.08),
                Colors.transparent,
              ],
            ),
          ),
          child: Column(
            children: [
              Container(
                width: 72,
                height: 72,
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [AppColors.gold, AppColors.goldMuted],
                  ),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: AppColors.goldBright, width: 2),
                  boxShadow: [
                    BoxShadow(
                      color: AppColors.gold.withOpacity(0.2),
                      blurRadius: 16,
                    ),
                  ],
                ),
                child: const Icon(Icons.person, size: 36, color: AppColors.bg),
              ),
              const SizedBox(height: 12),
              const Text(
                'hch2003',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppColors.goldBright,
                ),
              ),
              const SizedBox(height: 4),
              const Text(
                'ID: 8839201',
                style: TextStyle(fontSize: 11, color: AppColors.textMuted),
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  _Stat(label: l10n.gold, value: '2,450'),
                  const SizedBox(width: 24),
                  _Stat(label: l10n.hands, value: '128'),
                  const SizedBox(width: 24),
                  _Stat(label: l10n.winRate, value: '62%'),
                ],
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        // Menu items
        _MenuItem(icon: Icons.account_balance_wallet, label: l10n.recharge, onTap: () {}),
        _MenuItem(icon: Icons.history, label: l10n.handHistory, onTap: () {}),
        _MenuItem(icon: Icons.settings, label: l10n.settings, onTap: () {}),
        _MenuItem(icon: Icons.help_outline, label: l10n.help, onTap: () {}),
        _MenuItem(
          icon: Icons.logout,
          label: l10n.logout,
          color: AppColors.foldRed,
          onTap: () {},
        ),
      ],
    );
  }
}

class _Stat extends StatelessWidget {
  final String label;
  final String value;
  const _Stat({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(
          value,
          style: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
            color: AppColors.goldBright,
          ),
        ),
        const SizedBox(height: 2),
        Text(
          label,
          style: const TextStyle(fontSize: 10, color: AppColors.textMuted),
        ),
      ],
    );
  }
}

class _MenuItem extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color? color;
  final VoidCallback onTap;

  const _MenuItem({
    required this.icon,
    required this.label,
    this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final iconColor = color ?? AppColors.gold;
    final textColor = color ?? AppColors.goldBright;
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: GestureDetector(
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
          decoration: BoxDecoration(
            color: AppColors.surface,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: AppColors.gold.withOpacity(0.3)),
          ),
          child: Row(
            children: [
              Icon(icon, size: 18, color: iconColor),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  label,
                  style: TextStyle(fontSize: 13, color: textColor),
                ),
              ),
              Icon(Icons.chevron_right, size: 18, color: AppColors.gold.withOpacity(0.6)),
            ],
          ),
        ),
      ),
    );
  }
}
