import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import '../theme.dart';

class RankingScreen extends StatefulWidget {
  const RankingScreen({super.key});

  @override
  State<RankingScreen> createState() => _RankingScreenState();
}

class _RankingScreenState extends State<RankingScreen> {
  int _tab = 0;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final tabs = [l10n.wealthRank, l10n.winRateRank, l10n.streakRank];

    return Column(
      children: [
        // Tabs
        Container(
          decoration: const BoxDecoration(
            border: Border(
              bottom: BorderSide(color: AppColors.goldBorder, width: 0.5),
            ),
          ),
          child: Row(
            children: tabs.asMap().entries.map((e) {
              final i = e.key;
              final label = e.value;
              final active = i == _tab;
              return Expanded(
                child: GestureDetector(
                  onTap: () => setState(() => _tab = i),
                  child: Container(
                    padding: const EdgeInsets.symmetric(vertical: 10),
                    decoration: BoxDecoration(
                      border: Border(
                        bottom: BorderSide(
                          color: active ? AppColors.gold : Colors.transparent,
                          width: 2,
                        ),
                      ),
                    ),
                    child: Text(
                      label,
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: 11,
                        color: active ? AppColors.goldBright : AppColors.textMuted,
                        fontWeight: active ? FontWeight.bold : FontWeight.normal,
                      ),
                    ),
                  ),
                ),
              );
            }).toList(),
          ),
        ),
        // Podium
        const Padding(
          padding: EdgeInsets.all(16),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              _PodiumItem(rank: 2, emoji: '🦇', name: '静牌', value: '88.5万', color: Color(0xFFC0C0C0)),
              SizedBox(width: 16),
              _PodiumItem(rank: 1, emoji: '🧔', name: '柒少', value: '156.2万', color: AppColors.goldBright, isFirst: true),
              SizedBox(width: 16),
              _PodiumItem(rank: 3, emoji: '🦈', name: '超哥', value: '72.1万', color: Color(0xFFCD7F32)),
            ],
          ),
        ),
        // List
        Expanded(
          child: ListView.separated(
            padding: const EdgeInsets.symmetric(horizontal: 8),
            itemCount: 4,
            separatorBuilder: (_, __) => const SizedBox(height: 4),
            itemBuilder: (context, i) {
              final items = [
                const _RankRow(rank: 4, emoji: '👩‍🎤', name: '脆皮五华', value: '65.8万', isMe: false),
                const _RankRow(rank: 5, emoji: '🧔', name: '见南山', value: '58.3万', isMe: false),
                const _RankRow(rank: 8, emoji: '👩', name: '静牌', value: '35.2万', isMe: false),
                const _RankRow(rank: 12, emoji: '👤', name: 'hch2003 (我)', value: '12.4万', isMe: true),
              ];
              return items[i];
            },
          ),
        ),
      ],
    );
  }
}

class _PodiumItem extends StatelessWidget {
  final int rank;
  final String emoji;
  final String name;
  final String value;
  final Color color;
  final bool isFirst;

  const _PodiumItem({
    required this.rank,
    required this.emoji,
    required this.name,
    required this.value,
    required this.color,
    this.isFirst = false,
  });

  @override
  Widget build(BuildContext context) {
    final size = isFirst ? 48.0 : 40.0;
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: size,
          height: size,
          decoration: BoxDecoration(
            gradient: LinearGradient(colors: [color, color.withOpacity(0.7)]),
            borderRadius: BorderRadius.circular(size / 2),
            border: Border.all(color: AppColors.gold, width: 2),
            boxShadow: isFirst
                ? [BoxShadow(color: AppColors.gold.withOpacity(0.3), blurRadius: 12)]
                : null,
          ),
          child: Center(
            child: Text(emoji, style: TextStyle(fontSize: isFirst ? 22 : 18)),
          ),
        ),
        const SizedBox(height: 4),
        Text(name, style: TextStyle(fontSize: isFirst ? 10 : 9, fontWeight: FontWeight.bold, color: AppColors.goldBright)),
        Text(value, style: const TextStyle(fontSize: 8, color: AppColors.textMuted)),
      ],
    );
  }
}

class _RankRow extends StatelessWidget {
  final int rank;
  final String emoji;
  final String name;
  final String value;
  final bool isMe;

  const _RankRow({
    required this.rank,
    required this.emoji,
    required this.name,
    required this.value,
    required this.isMe,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
      decoration: BoxDecoration(
        color: isMe ? AppColors.surface.withOpacity(0.8) : AppColors.surface.withOpacity(0.4),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: isMe ? AppColors.gold : AppColors.gold.withOpacity(0.2)),
      ),
      child: Row(
        children: [
          SizedBox(
            width: 20,
            child: Text(
              '$rank',
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 10,
                fontWeight: FontWeight.bold,
                color: isMe ? AppColors.goldBright : AppColors.textMuted,
              ),
            ),
          ),
          const SizedBox(width: 6),
          Container(
            width: 26,
            height: 26,
            decoration: BoxDecoration(
              gradient: const LinearGradient(colors: [AppColors.gold, AppColors.goldMuted]),
              borderRadius: BorderRadius.circular(4),
              border: Border.all(color: AppColors.goldBright),
            ),
            child: Center(child: Text(emoji, style: const TextStyle(fontSize: 12))),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              name,
              style: TextStyle(
                fontSize: 10,
                color: isMe ? AppColors.goldBright : AppColors.goldBright.withOpacity(0.8),
              ),
            ),
          ),
          Text(
            value,
            style: TextStyle(
              fontSize: 9,
              fontWeight: isMe ? FontWeight.bold : FontWeight.normal,
              color: isMe ? AppColors.goldBright : AppColors.textMuted,
            ),
          ),
        ],
      ),
    );
  }
}
