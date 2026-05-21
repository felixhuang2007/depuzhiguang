import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:depuzhiguang/screens/login_screen.dart';
import 'package:depuzhiguang/blocs/auth_bloc.dart';

void main() {
  testWidgets('LoginScreen shows brand title and login button', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('zh'),
        home: BlocProvider(
          create: (_) => AuthBloc(),
          child: const LoginScreen(),
        ),
      ),
    );

    expect(find.text('德扑之光'), findsOneWidget);
    expect(find.text('登录'), findsOneWidget);
    expect(find.byType(TextField), findsNWidgets(2));
  });
}
