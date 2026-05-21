import 'package:flutter/material.dart';
import '../theme.dart';
import '../widgets/app_header.dart';
import '../widgets/bottom_nav.dart';
import 'lobby_screen.dart';

class _PlaceholderScreen extends StatelessWidget {
  final String label;
  const _PlaceholderScreen({required this.label});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Text(
        label,
        style: const TextStyle(color: AppColors.goldBright),
      ),
    );
  }
}

class MainScreen extends StatefulWidget {
  const MainScreen({super.key});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> {
  int _index = 0;

  final _screens = const [
    LobbyScreen(),
    _PlaceholderScreen(label: '社交'),
    _PlaceholderScreen(label: '排行'),
    _PlaceholderScreen(label: '我的'),
  ];

  final _titles = ['大厅', '社交', '排行', '我的'];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppHeader(title: _titles[_index]),
      body: IndexedStack(
        index: _index,
        children: _screens,
      ),
      bottomNavigationBar: BottomNav(
        currentIndex: _index,
        onTap: (i) => setState(() => _index = i),
      ),
    );
  }
}
