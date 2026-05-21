import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:depuzhiguang/screens/lobby_screen.dart';
import 'package:depuzhiguang/blocs/lobby_bloc.dart';

void main() {
  testWidgets('LobbyScreen shows filter tabs and table cards', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        localizationsDelegates: const [
          AppLocalizations.delegate,
          GlobalMaterialLocalizations.delegate,
          GlobalWidgetsLocalizations.delegate,
          GlobalCupertinoLocalizations.delegate,
        ],
        supportedLocales: const [
          Locale('zh'),
          Locale('en'),
        ],
        locale: const Locale('zh'),
        home: BlocProvider(
          create: (_) => LobbyBloc(),
          child: const LobbyScreen(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('现金桌'), findsOneWidget);
    expect(find.text('SNG'), findsOneWidget);
    expect(find.text('经典六人桌'), findsOneWidget);
  });
}
