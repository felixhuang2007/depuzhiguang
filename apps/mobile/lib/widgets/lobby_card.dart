import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import '../theme.dart';
import '../models/table_info.dart';

class LobbyCard extends StatelessWidget {
  final TableInfo table;
  final VoidCallback onTap;

  const LobbyCard({super.key, required this.table, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isFull = table.isFull;
    final isAlmostFull = !isFull && table.currentPlayers >= table.maxPlayers - 1;

    Color borderColor;
    Color dotColor;
    Color nameColor;
    Color playerColor;
    if (isFull) {
      borderColor = AppColors.full.withOpacity(0.5);
      dotColor = AppColors.full;
      nameColor = AppColors.textMuted;
      playerColor = AppColors.full;
    } else if (isAlmostFull) {
      borderColor = AppColors.gold.withOpacity(0.4);
      dotColor = AppColors.textMuted;
      nameColor = AppColors.goldBright;
      playerColor = AppColors.textMuted;
    } else {
      borderColor = AppColors.goldBorder;
      dotColor = AppColors.gold;
      nameColor = AppColors.goldBright;
      playerColor = AppColors.goldBright;
    }

    return GestureDetector(
      onTap: isFull ? null : onTap,
      child: Opacity(
        opacity: isFull ? 0.6 : 1.0,
        child: Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: AppColors.surface,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: borderColor),
            boxShadow: [
              BoxShadow(
                color: AppColors.gold.withOpacity(0.05),
                blurRadius: 8,
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    table.name,
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.bold,
                      color: nameColor,
                    ),
                  ),
                  Row(
                    children: [
                      Container(
                        width: 7,
                        height: 7,
                        decoration: BoxDecoration(
                          color: dotColor,
                          borderRadius: BorderRadius.circular(3.5),
                          boxShadow: dotColor == AppColors.gold
                              ? [BoxShadow(color: dotColor.withOpacity(0.5), blurRadius: 4)]
                              : null,
                        ),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        isFull ? l10n.full : l10n.online,
                        style: TextStyle(fontSize: 9, color: dotColor),
                      ),
                    ],
                  ),
                ],
              ),
              const SizedBox(height: 6),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      Text(
                        table.stakes,
                        style: const TextStyle(fontSize: 11, color: AppColors.textMuted),
                      ),
                      const SizedBox(width: 6),
                      Text('·', style: TextStyle(fontSize: 11, color: AppColors.textMuted.withOpacity(0.5))),
                      const SizedBox(width: 6),
                      Text(
                        '${l10n.limit} ${table.limit}',
                        style: const TextStyle(fontSize: 11, color: AppColors.textMuted),
                      ),
                    ],
                  ),
                  Text(
                    '${table.currentPlayers}/${table.maxPlayers} ${l10n.players}',
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                      color: playerColor,
                    ),
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
