import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:depuzhiguang/main.dart' as app;
import 'package:depuzhiguang/widgets/lobby_card.dart';

// ── Integration Test: 牌桌完整流程端到端测试 ───────────────────────
// 运行方式:
//   flutter test integration_test/table_flow_test.dart

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.clear();
  });

  group('牌桌游戏流程 E2E', () {
    testWidgets('完整流程: 登录 -> 大厅 -> 牌桌(观战) -> 加入 -> 离开', (tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // ── 1. 登录 ──
      // 使用长按 logo 自动填充测试账号
      final logo = find.byType(GestureDetector).first;
      await tester.longPress(logo);
      await tester.pumpAndSettle(const Duration(seconds: 2));

      // 点击登录按钮
      await tester.tap(find.byType(ElevatedButton));
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // 验证已进入主界面（通过底部导航或 IndexedStack）
      expect(find.byType(IndexedStack), findsOneWidget);

      // ── 2. 点击第一个牌桌卡片进入牌桌 ──
      final cards = find.byType(LobbyCard);
      expect(cards, findsWidgets);
      await tester.tap(cards.first);
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // ── 3. 验证牌桌界面已加载 ──
      expect(find.byType(AppBar), findsNothing);

      // 等待牌桌数据加载完成
      await tester.pumpAndSettle(const Duration(seconds: 8));

      // ── 4. 验证观战模式: 查找"加入游戏"按钮 ──
      // 如果加载失败可能显示错误，先检查错误状态
      final errorWidgets = find.textContaining('Failed');
      if (errorWidgets.evaluate().isNotEmpty) {
        // 网络错误，尝试截图诊断
        await tester.pumpAndSettle(const Duration(seconds: 5));
      }

      final joinBtn = find.textContaining('加入游戏');
      expect(joinBtn, findsOneWidget, reason: '观战模式应显示加入游戏按钮');

      // ── 5. 点击加入游戏 ──
      await tester.tap(joinBtn);
      await tester.pumpAndSettle(const Duration(seconds: 2));

      // ── 6. 验证弹窗出现并点击确定 ──
      final confirmBtn = find.textContaining('确定');
      expect(confirmBtn, findsWidgets, reason: '应显示确认弹窗');
      await tester.tap(confirmBtn.last);
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // ── 7. 验证已加入: 查找操作按钮(弃牌/跟分) ──
      final foldBtn = find.textContaining('弃牌');
      expect(foldBtn, findsOneWidget, reason: '加入后应显示弃牌按钮');

      // ── 8. 点击 toolbar 的离开按钮 (logout icon) ──
      // toolbar 按钮是 GestureDetector 包裹的 Icon
      final toolbarIcons = find.byType(Icon);
      // 查找 logout icon (Icons.logout)
      final logoutIcon = find.byIcon(Icons.logout);
      if (logoutIcon.evaluate().isNotEmpty) {
        await tester.tap(logoutIcon);
        await tester.pumpAndSettle(const Duration(seconds: 2));

        // ── 9. 确认离开弹窗 ──
        final leaveConfirm = find.textContaining('确定');
        expect(leaveConfirm, findsWidgets);
        await tester.tap(leaveConfirm.last);
        await tester.pumpAndSettle(const Duration(seconds: 5));

        // ── 10. 验证回到观战模式 ──
        final rejoinBtn = find.textContaining('加入游戏');
        expect(rejoinBtn, findsOneWidget, reason: '离开后应回到观战模式');
      }
    });
  });
}
