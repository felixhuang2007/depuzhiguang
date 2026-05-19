import 'package:flutter/material.dart';
import '../models/card.dart';
import '../services/websocket_service.dart';
import '../widgets/card_widget.dart';

class TableScreen extends StatefulWidget {
  final String tableId;
  const TableScreen({super.key, required this.tableId});

  @override
  State<TableScreen> createState() => _TableScreenState();
}

class _TableScreenState extends State<TableScreen> {
  final _ws = WebSocketService();
  List<PokerCard> _community = [];
  int _pot = 0;

  @override
  void initState() {
    super.initState();
    _ws.connect('ws://localhost:8443/ws?player_id=user1');
    _ws.messages.listen(_handleMessage);
  }

  void _handleMessage(Map<String, dynamic> msg) {
    setState(() {
      // TODO: parse state snapshot / delta
    });
  }

  void _sendAction(String action) {
    _ws.send({'type': 'action', 'payload': {'action': action}});
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          Container(color: const Color(0xFF35654d)),
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
          Positioned(
            top: 80,
            left: 0,
            right: 0,
            child: Center(
              child: Text('Pot: $_pot bb', style: const TextStyle(fontSize: 18, color: Colors.white)),
            ),
          ),
          Positioned(
            bottom: 20,
            left: 0,
            right: 0,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                ElevatedButton(onPressed: () => _sendAction('fold'), child: const Text('Fold')),
                ElevatedButton(onPressed: () => _sendAction('check'), child: const Text('Check')),
                ElevatedButton(onPressed: () => _sendAction('call'), child: const Text('Call')),
                ElevatedButton(onPressed: () => _sendAction('raise'), child: const Text('Raise')),
              ],
            ),
          ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _ws.disconnect();
    super.dispose();
  }
}
