import 'package:flutter/material.dart';
import '../theme.dart';

class JoinTableDialog extends StatelessWidget {
  final int buyinGold;
  final int buyinChips;
  final VoidCallback? onConfirm;
  final VoidCallback? onCancel;

  const JoinTableDialog({
    super.key,
    this.buyinGold = 1000,
    this.buyinChips = 10000,
    this.onConfirm,
    this.onCancel,
  });

  @override
  Widget build(BuildContext context) {
    return Dialog(
      backgroundColor: Colors.transparent,
      child: Container(
        width: 280,
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: AppColors.bg,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: AppColors.gold, width: 1.5),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.5),
              blurRadius: 20,
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text(
              '加入游戏',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppColors.goldBright,
              ),
            ),
            const SizedBox(height: 16),
            _infoRow('消耗金币', '$buyinGold', '🪙'),
            const SizedBox(height: 8),
            _infoRow('带入积分', '$buyinChips', '💎'),
            const SizedBox(height: 20),
            Row(
              children: [
                Expanded(
                  child: _dialogButton(
                    '取消',
                    AppColors.surface.withOpacity(0.8),
                    AppColors.goldBright,
                    onCancel ?? () => Navigator.of(context).pop(false),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: _dialogButton(
                    '确定',
                    AppColors.foldRed,
                    AppColors.goldBright,
                    onConfirm ?? () => Navigator.of(context).pop(true),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _infoRow(String label, String value, String icon) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text(
          label,
          style: const TextStyle(fontSize: 13, color: AppColors.textMuted),
        ),
        const SizedBox(width: 8),
        Text(icon, style: const TextStyle(fontSize: 14)),
        const SizedBox(width: 4),
        Text(
          value,
          style: const TextStyle(
            fontSize: 15,
            fontWeight: FontWeight.bold,
            color: AppColors.goldBright,
          ),
        ),
      ],
    );
  }

  Widget _dialogButton(String text, Color bg, Color fg, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 10),
        decoration: BoxDecoration(
          color: bg,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: AppColors.gold.withOpacity(0.3)),
        ),
        alignment: Alignment.center,
        child: Text(
          text,
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.bold,
            color: fg,
          ),
        ),
      ),
    );
  }
}
