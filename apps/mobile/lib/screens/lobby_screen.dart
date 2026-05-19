import 'package:flutter/material.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'table_screen.dart';

class LobbyScreen extends StatelessWidget {
  const LobbyScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.lobby),
        actions: [
          IconButton(icon: const Icon(Icons.person), onPressed: () {}),
          IconButton(icon: const Icon(Icons.settings), onPressed: () {}),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _Section(title: l10n.cashGame, children: [
            _TableCard(stakes: '1/2', players: '3/9', bb: '10', onTap: () => _joinTable(context)),
            _TableCard(stakes: '5/10', players: '6/9', bb: '50', onTap: () => _joinTable(context)),
            _TableCard(stakes: '25/50', players: '2/6', bb: '250', onTap: () => _joinTable(context)),
          ]),
          _Section(title: l10n.tournament, children: [
            _TournamentCard(name: 'Daily 10K', buyIn: '1,000', entrants: '45/100'),
            _TournamentCard(name: 'Weekly 100K', buyIn: '10,000', entrants: '120/500'),
          ]),
          _Section(title: l10n.club, children: [
            ListTile(
              leading: const Icon(Icons.group),
              title: const Text('My Clubs'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () {},
            ),
          ]),
        ],
      ),
    );
  }

  void _joinTable(BuildContext context) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => const TableScreen(tableId: 'table_1')),
    );
  }
}

class _Section extends StatelessWidget {
  final String title;
  final List<Widget> children;
  const _Section({required this.title, required this.children});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(vertical: 12),
          child: Text(title, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
        ),
        ...children,
        const SizedBox(height: 16),
      ],
    );
  }
}

class _TableCard extends StatelessWidget {
  final String stakes;
  final String players;
  final String bb;
  final VoidCallback onTap;

  const _TableCard({required this.stakes, required this.players, required this.bb, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        leading: const Icon(Icons.casino),
        title: Text('$stakes BB'),
        subtitle: Text('$players players'),
        trailing: Text('$bb BB'),
        onTap: onTap,
      ),
    );
  }
}

class _TournamentCard extends StatelessWidget {
  final String name;
  final String buyIn;
  final String entrants;

  const _TournamentCard({required this.name, required this.buyIn, required this.entrants});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        leading: const Icon(Icons.emoji_events),
        title: Text(name),
        subtitle: Text('Buy-in: $buyIn Gold'),
        trailing: Text(entrants),
      ),
    );
  }
}
