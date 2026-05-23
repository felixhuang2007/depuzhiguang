import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:depuzhiguang/main.dart';

void main() {
  testWidgets('App launches and shows login screen', (WidgetTester tester) async {
    await tester.pumpWidget(const MyApp());

    // Verify login screen elements
    expect(find.text('德扑之光'), findsOneWidget);
    expect(find.text('登录'), findsOneWidget);
    expect(find.byType(TextField), findsNWidgets(2));
  });

  testWidgets('Theme uses red-gold colors', (WidgetTester tester) async {
    await tester.pumpWidget(const MyApp());

    final materialApp = tester.widget<MaterialApp>(find.byType(MaterialApp));
    expect(materialApp.theme?.scaffoldBackgroundColor, const Color(0xFF2D0A0F));
  });
}
