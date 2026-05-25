import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:depuzhiguang/screens/table_screen.dart';

void main() {
  testWidgets('TableScreen shows green felt and action buttons', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: TableScreen(tableId: 't1')),
    );

    expect(find.text('弃牌'), findsAtLeastNWidgets(1));
    expect(find.text('跟分'), findsOneWidget);
    expect(find.text('底池: 5.3BB'), findsOneWidget);
  });
}
