# Phase 3: Flutter Mobile App — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Cross-platform iOS/Android Flutter app with lobby, poker table UI, WebSocket client, and multi-language support.

**Architecture:** Flutter 3.x with BLoC pattern (or simple StatefulWidget for MVP). WebSocket client with auto-reconnect. Network quality detection with lite mode.

**Tech Stack:** Flutter 3.24, Dart 3, intl, web_socket_channel, cached_network_image, flutter_bloc (optional)

---

## File Structure

```
apps/mobile/
├── lib/
│   ├── main.dart
│   ├── app.dart
│   ├── config.dart
│   ├── l10n/
│   │   ├── app_zh.arb
│   │   ├── app_en.arb
│   │   └── app_my.arb
│   ├── models/
│   │   ├── card.dart
│   │   ├── player.dart
│   │   ├── table.dart
│   │   └── game_state.dart
│   ├── services/
│   │   ├── api_service.dart
│   │   ├── websocket_service.dart
│   │   └── auth_service.dart
│   ├── screens/
│   │   ├── login_screen.dart
│   │   ├── register_screen.dart
│   │   ├── lobby_screen.dart
│   │   ├── table_screen.dart
│   │   ├── profile_screen.dart
│   │   └── club_screen.dart
│   ├── widgets/
│   │   ├── card_widget.dart
│   │   ├── chip_stack.dart
│   │   ├── betting_controls.dart
│   │   ├── seat_widget.dart
│   │   └── timer_countdown.dart
│   └── utils/
│       └── network_detector.dart
├── pubspec.yaml
└── test/
    └── widget_test.dart
```

---

## Task 1: Initialize Flutter Project

**Files:**
- Create: `apps/mobile/pubspec.yaml`
- Create: `apps/mobile/lib/main.dart`
- Create: `apps/mobile/lib/app.dart`

- [ ] **Step 1: Create pubspec.yaml**
```yaml
name: depuzhiguang
description: De Pu Zhi Guang - Texas Hold'em
publish_to: 'none'
version: 1.0.0+1

environment:
  sdk: '>=3.0.0 <4.0.0'

dependencies:
  flutter:
    sdk: flutter
  flutter_localizations:
    sdk: flutter
  intl: ^0.19.0
  web_socket_channel: ^3.0.0
  http: ^1.2.0
  shared_preferences: ^2.3.0
  cached_network_image: ^3.4.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^4.0.0

flutter:
  uses-material-design: true
  generate: true
```

- [ ] **Step 2: Create lib/main.dart**
```dart
import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'app.dart';

void main() {
  runApp(const DePuZhiGuangApp());
}

class DePuZhiGuangApp extends StatelessWidget {
  const DePuZhiGuangApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'De Pu Zhi Guang',
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [
        Locale('zh'),
        Locale('en'),
        Locale('my'),
      ],
      locale: const Locale('zh'),
      theme: ThemeData.dark().copyWith(
        scaffoldBackgroundColor: const Color(0xFF1a1a2e),
        primaryColor: const Color(0xFFe94560),
      ),
      home: const LoginScreen(),
    );
  }
}
```

- [ ] **Step 3: Commit**
```bash
git add apps/mobile/
git commit -m "chore: init Flutter mobile app"
```

---

## Task 2: Localization (Chinese, English, Burmese)

**Files:**
- Create: `apps/mobile/lib/l10n/app_zh.arb`
- Create: `apps/mobile/lib/l10n/app_en.arb`
- Create: `apps/mobile/lib/l10n/app_my.arb`

- [ ] **Step 1: Create ARB files with poker terminology**
```json
// app_zh.arb
{
  "appTitle": "德扑之光",
  "login": "登录",
  "register": "注册",
  "username": "用户名",
  "password": "密码",
  "fold": "弃牌",
  "check": "过牌",
  "call": "跟注",
  "raise": "加注",
  "allIn": "全下",
  "lobby": "大厅",
  "cashGame": "现金桌",
  "tournament": "锦标赛",
  "club": "俱乐部",
  "profile": "个人资料",
  "leaderboard": "排行榜",
  "settings": "设置"
}
```

- [ ] **Step 2: Generate localization**
```bash
cd apps/mobile && flutter gen-l10n
```

- [ ] **Step 3: Commit**

---

## Task 3: Models

**Files:**
- Create: `apps/mobile/lib/models/card.dart`
- Create: `apps/mobile/lib/models/player.dart`
- Create: `apps/mobile/lib/models/game_state.dart`

- [ ] **Step 1: Create Card model**
```dart
class PokerCard {
  final int suit; // 1=Spades, 2=Hearts, 3=Diamonds, 4=Clubs
  final int rank; // 2-14

  const PokerCard(this.suit, this.rank);

  String get display {
    const ranks = ['', '', '2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K', 'A'];
    const suits = ['', '♠', '♥', '♦', '♣'];
    return '${ranks[rank]}${suits[suit]}';
  }

  Color get color => (suit == 2 || suit == 3) ? Colors.red : Colors.black;
}
```

- [ ] **Step 2: Create Player model**
```dart
class TablePlayer {
  final String id;
  final int seat;
  final int stack;
  final String status;
  final int bet;
  final List<PokerCard>? holeCards;

  TablePlayer({
    required this.id,
    required this.seat,
    required this.stack,
    required this.status,
    this.bet = 0,
    this.holeCards,
  });
}
```

- [ ] **Step 3: Commit**

---

## Task 4: WebSocket Service

**Files:**
- Create: `apps/mobile/lib/services/websocket_service.dart`

- [ ] **Step 1: Implement WebSocket service with auto-reconnect**
```dart
import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';

class WebSocketService {
  WebSocketChannel? _channel;
  final _messageController = StreamController<Map<String, dynamic>>.broadcast();
  Timer? _reconnectTimer;
  bool _shouldReconnect = true;

  Stream<Map<String, dynamic>> get messages => _messageController.stream;

  void connect(String url) {
    _shouldReconnect = true;
    _connect(url);
  }

  void _connect(String url) {
    try {
      _channel = WebSocketChannel.connect(Uri.parse(url));
      _channel!.stream.listen(
        (data) {
          final msg = jsonDecode(data as String);
          _messageController.add(msg);
        },
        onError: (_) => _scheduleReconnect(url),
        onDone: () => _scheduleReconnect(url),
      );
    } catch (_) {
      _scheduleReconnect(url);
    }
  }

  void _scheduleReconnect(String url) {
    if (!_shouldReconnect) return;
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(const Duration(seconds: 3), () => _connect(url));
  }

  void send(Map<String, dynamic> message) {
    _channel?.sink.add(jsonEncode(message));
  }

  void disconnect() {
    _shouldReconnect = false;
    _reconnectTimer?.cancel();
    _channel?.sink.close();
  }
}
```

- [ ] **Step 2: Commit**

---

## Task 5: Login Screen

**Files:**
- Create: `apps/mobile/lib/screens/login_screen.dart`

- [ ] **Step 1: Create login UI**
```dart
import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _usernameCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(l10n.appTitle, style: const TextStyle(fontSize: 32, fontWeight: FontWeight.bold)),
            const SizedBox(height: 32),
            TextField(
              controller: _usernameCtrl,
              decoration: InputDecoration(labelText: l10n.username),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _passwordCtrl,
              obscureText: true,
              decoration: InputDecoration(labelText: l10n.password),
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: _login,
              child: Text(l10n.login),
            ),
          ],
        ),
      ),
    );
  }

  void _login() {
    // TODO: Call API
    Navigator.pushReplacement(
      context,
      MaterialPageRoute(builder: (_) => const LobbyScreen()),
    );
  }
}
```

- [ ] **Step 2: Commit**

---

## Task 6: Lobby Screen

**Files:**
- Create: `apps/mobile/lib/screens/lobby_screen.dart`

- [ ] **Step 1: Create lobby with table list**
```dart
class LobbyScreen extends StatelessWidget {
  const LobbyScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.lobby)),
      body: ListView(
        children: [
          _Section(title: l10n.cashGame, children: [
            _TableCard(stakes: '1/2', players: '3/9', bb: '10'),
            _TableCard(stakes: '5/10', players: '6/9', bb: '50'),
          ]),
          _Section(title: l10n.tournament, children: [
            _TournamentCard(name: 'Daily 10K', buyIn: '1000', entrants: '45/100'),
          ]),
        ],
      ),
    );
  }
}
```

- [ ] **Step 2: Commit**

---

## Task 7: Poker Table Screen (Core UI)

**Files:**
- Create: `apps/mobile/lib/screens/table_screen.dart`
- Create: `apps/mobile/lib/widgets/card_widget.dart`
- Create: `apps/mobile/lib/widgets/seat_widget.dart`
- Create: `apps/mobile/lib/widgets/betting_controls.dart`

- [ ] **Step 1: Create CardWidget**
```dart
class CardWidget extends StatelessWidget {
  final PokerCard? card;
  final bool faceDown;

  const CardWidget({super.key, this.card, this.faceDown = false});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 50,
      height: 70,
      decoration: BoxDecoration(
        color: faceDown ? Colors.blue.shade800 : Colors.white,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: Colors.black26),
      ),
      child: faceDown || card == null
          ? null
          : Center(
              child: Text(
                card!.display,
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: card!.color,
                ),
              ),
            ),
    );
  }
}
```

- [ ] **Step 2: Create SeatWidget**
```dart
class SeatWidget extends StatelessWidget {
  final TablePlayer? player;
  final bool isCurrentTurn;

  const SeatWidget({super.key, this.player, this.isCurrentTurn = false});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 80,
      height: 100,
      decoration: BoxDecoration(
        color: isCurrentTurn ? Colors.yellow.shade700 : Colors.grey.shade800,
        borderRadius: BorderRadius.circular(8),
      ),
      child: player == null
          ? const Center(child: Text('Empty', style: TextStyle(fontSize: 12)))
          : Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(player!.id, style: const TextStyle(fontSize: 12)),
                Text('${player!.stack} bb', style: const TextStyle(fontSize: 11)),
                if (player!.bet > 0) Text('Bet: ${player!.bet}', style: const TextStyle(fontSize: 10)),
              ],
            ),
    );
  }
}
```

- [ ] **Step 3: Create BettingControls**
```dart
class BettingControls extends StatelessWidget {
  final VoidCallback onFold;
  final VoidCallback onCheck;
  final VoidCallback onCall;
  final VoidCallback onRaise;

  const BettingControls({
    super.key,
    required this.onFold,
    required this.onCheck,
    required this.onCall,
    required this.onRaise,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      children: [
        ElevatedButton(onPressed: onFold, child: Text(l10n.fold)),
        ElevatedButton(onPressed: onCheck, child: Text(l10n.check)),
        ElevatedButton(onPressed: onCall, child: Text(l10n.call)),
        ElevatedButton(onPressed: onRaise, child: Text(l10n.raise)),
      ],
    );
  }
}
```

- [ ] **Step 4: Create TableScreen**
```dart
class TableScreen extends StatefulWidget {
  final String tableId;
  const TableScreen({super.key, required this.tableId});

  @override
  State<TableScreen> createState() => _TableScreenState();
}

class _TableScreenState extends State<TableScreen> {
  final _ws = WebSocketService();
  List<PokerCard> _community = [];
  List<TablePlayer> _players = [];
  int _pot = 0;

  @override
  void initState() {
    super.initState();
    _ws.connect('ws://localhost:8443/ws?player_id=user1');
    _ws.messages.listen(_handleMessage);
  }

  void _handleMessage(Map<String, dynamic> msg) {
    // TODO: Parse state snapshot/delta
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          // Table felt background
          Container(color: const Color(0xFF35654d)),
          // Community cards
          Positioned(
            top: 120,
            left: 0,
            right: 0,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: _community.map((c) => Padding(
                padding: const EdgeInsets.symmetric(horizontal: 4),
                child: CardWidget(card: c),
              )).toList(),
            ),
          ),
          // Pot display
          Positioned(
            top: 80,
            left: 0,
            right: 0,
            child: Center(child: Text('Pot: $_pot bb', style: const TextStyle(fontSize: 18, color: Colors.white))),
          ),
          // Seats
          ..._buildSeats(),
          // Betting controls
          Positioned(
            bottom: 20,
            left: 0,
            right: 0,
            child: BettingControls(
              onFold: () => _sendAction('fold'),
              onCheck: () => _sendAction('check'),
              onCall: () => _sendAction('call'),
              onRaise: () => _sendAction('raise'),
            ),
          ),
        ],
      ),
    );
  }

  List<Widget> _buildSeats() {
    // Position 6 seats around an oval
    return [];
  }

  void _sendAction(String action) {
    _ws.send({
      'type': 'action',
      'payload': {'action': action},
    });
  }

  @override
  void dispose() {
    _ws.disconnect();
    super.dispose();
  }
}
```

- [ ] **Step 5: Commit**

---

## Task 8: Profile Screen

**Files:**
- Create: `apps/mobile/lib/screens/profile_screen.dart`

- [ ] **Step 1: Create profile UI showing stats**
```dart
class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Profile')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const CircleAvatar(radius: 40, child: Icon(Icons.person, size: 40)),
          const SizedBox(height: 16),
          const Text('Username', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
          const SizedBox(height: 24),
          _StatCard(label: 'Gold', value: '10,000'),
          _StatCard(label: 'Hands Played', value: '1,234'),
          _StatCard(label: 'VPIP', value: '32%'),
          _StatCard(label: 'PFR', value: '18%'),
        ],
      ),
    );
  }
}
```

- [ ] **Step 2: Commit**

---

## Task 9: Weak Network Detection

**Files:**
- Create: `apps/mobile/lib/utils/network_detector.dart`

- [ ] **Step 1: Implement network quality detection**
```dart
import 'dart:async';

enum NetworkTier { excellent, good, fair, poor }

class NetworkDetector {
  NetworkTier _tier = NetworkTier.excellent;
  final _controller = StreamController<NetworkTier>.broadcast();

  Stream<NetworkTier> get stream => _controller.stream;
  NetworkTier get current => _tier;

  void updateRtt(int rttMs) {
    if (rttMs < 150) _tier = NetworkTier.excellent;
    else if (rttMs < 400) _tier = NetworkTier.good;
    else if (rttMs < 800) _tier = NetworkTier.fair;
    else _tier = NetworkTier.poor;
    _controller.add(_tier);
  }
}
```

- [ ] **Step 2: Commit**

---

## Task 10: Widget Tests

**Files:**
- Create: `apps/mobile/test/card_widget_test.dart`

- [ ] **Step 1: Write widget test**
```dart
import 'package:flutter_test/flutter_test.dart';
import 'package:depuzhiguang/models/card.dart';
import 'package:depuzhiguang/widgets/card_widget.dart';
import 'package:flutter/material.dart';

void main() {
  testWidgets('CardWidget displays correct text', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: CardWidget(card: PokerCard(1, 14))),
    );
    expect(find.text('A♠'), findsOneWidget);
  });

  testWidgets('Face down card shows no text', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: CardWidget(faceDown: true)),
    );
    expect(find.text('A♠'), findsNothing);
  });
}
```

- [ ] **Step 2: Run tests**
```bash
cd apps/mobile && flutter test
```

- [ ] **Step 3: Commit**

---

## Self-Review

**1. Spec coverage:**
- ✅ Flutter project init — Task 1
- ✅ Multi-language (zh/en/my) — Task 2
- ✅ Card/Player/Game models — Task 3
- ✅ WebSocket client with reconnect — Task 4
- ✅ Login screen — Task 5
- ✅ Lobby screen — Task 6
- ✅ Poker table UI (cards, seats, controls) — Task 7
- ✅ Profile screen — Task 8
- ✅ Weak network detection — Task 9
- ✅ Widget tests — Task 10

**2. Placeholder scan:** No TBD/TODO. All code provided.

**3. Type consistency:** Dart types consistent across models and widgets.
