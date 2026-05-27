import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:depuzhiguang/main.dart' as app;
import 'package:depuzhiguang/widgets/lobby_card.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.clear();
  });

  testWidgets('调试2: 进入牌桌后立即打印文本', (tester) async {
    app.main();
    await tester.pumpAndSettle(const Duration(seconds: 3));

    // 登录 - 直接填写表单
    final textFields = find.byType(TextField);
    await tester.enterText(textFields.at(0), 'testplayer2');
    await tester.pumpAndSettle(const Duration(seconds: 1));
    await tester.enterText(textFields.at(1), 'Test@1234');
    await tester.pumpAndSettle(const Duration(seconds: 1));

    // 收起键盘后再点击登录
    await tester.testTextInput.receiveAction(TextInputAction.done);
    await tester.pumpAndSettle(const Duration(seconds: 1));

    await tester.tap(find.byType(ElevatedButton));
    await tester.pumpAndSettle(const Duration(seconds: 8));

    expect(find.byType(IndexedStack), findsOneWidget);

    // 点击第一个牌桌
    final cards = find.byType(LobbyCard);
    await tester.tap(cards.first);
    await tester.pumpAndSettle(const Duration(seconds: 3));

    // 打印所有可见文本
    final allTexts = find.byType(Text);
    final count = allTexts.evaluate().length;
    print('=== 进入牌桌后可见文本数量: $count ===');
    for (int i = 0; i < count && i < 60; i++) {
      final widget = tester.widget<Text>(allTexts.at(i));
      print('Text $i: "${widget.data}"');
    }
  });
}
