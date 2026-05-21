import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import '../theme.dart';

class SocialScreen extends StatefulWidget {
  const SocialScreen({super.key});

  @override
  State<SocialScreen> createState() => _SocialScreenState();
}

class _SocialScreenState extends State<SocialScreen> {
  int _tab = 0;

  static const _messages = [
    _MessageItem(
      emoji: '🧔',
      name: '柒少',
      message: '在吗？今晚一起开桌？',
      time: '10:23',
      unread: 2,
    ),
    _MessageItem(
      emoji: '👩‍🎤',
      name: '脆皮五华',
      message: '哈哈，那把牌太刺激了',
      time: '昨天',
    ),
    _MessageItem(
      emoji: '🦈',
      name: '超哥',
      message: '明天下午有局，来吗？',
      time: '周一',
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final tabs = [l10n.messages, l10n.friends, l10n.inbox, l10n.chatroom];

    return Column(
      children: [
        // Sub tabs
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
        // Message list
        Expanded(
          child: ListView.separated(
            padding: const EdgeInsets.all(8),
            itemCount: _messages.length,
            separatorBuilder: (_, __) => const SizedBox(height: 6),
            itemBuilder: (context, i) => _messages[i],
          ),
        ),
      ],
    );
  }
}

class _MessageItem extends StatelessWidget {
  final String emoji;
  final String name;
  final String message;
  final String time;
  final int? unread;

  const _MessageItem({
    required this.emoji,
    required this.name,
    required this.message,
    required this.time,
    this.unread,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppColors.gold.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          Container(
            width: 36,
            height: 36,
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [AppColors.gold, AppColors.goldMuted],
              ),
              borderRadius: BorderRadius.circular(6),
              border: Border.all(color: AppColors.goldBright),
            ),
            child: Center(child: Text(emoji, style: const TextStyle(fontSize: 16))),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(name, style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: AppColors.goldBright)),
                const SizedBox(height: 2),
                Text(message, style: const TextStyle(fontSize: 9, color: AppColors.textMuted)),
              ],
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(time, style: const TextStyle(fontSize: 8, color: AppColors.textMuted)),
              if (unread != null) ...[
                const SizedBox(height: 2),
                Container(
                  width: 16,
                  height: 16,
                  decoration: BoxDecoration(
                    color: AppColors.foldRed,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: AppColors.gold),
                  ),
                  child: Center(
                    child: Text(
                      '$unread',
                      style: const TextStyle(fontSize: 8, color: AppColors.goldBright),
                    ),
                  ),
                ),
              ],
            ],
          ),
        ],
      ),
    );
  }
}
