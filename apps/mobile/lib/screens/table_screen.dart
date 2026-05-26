import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../blocs/table_bloc.dart';
import '../models/card.dart';
import '../models/player.dart';
import '../repositories/auth_repository.dart';
import '../theme.dart';
import '../widgets/join_table_dialog.dart';
import '../widgets/leave_table_dialog.dart';
import '../widgets/poker_card_widget.dart';
import '../widgets/player_avatar.dart';
import '../widgets/action_button.dart';

class TableScreen extends StatelessWidget {
  final String tableId;
  const TableScreen({super.key, required this.tableId});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => TableBloc(),
      child: _TableView(tableId: tableId),
    );
  }
}

class _TableView extends StatefulWidget {
  final String tableId;
  const _TableView({required this.tableId});

  @override
  State<_TableView> createState() => _TableViewState();
}

class _TableViewState extends State<_TableView> {
  final _community = [
    const PokerCard(2, 14), // A♥
    const PokerCard(1, 13), // K♠
    const PokerCard(4, 12), // Q♣
  ];

  @override
  void initState() {
    super.initState();
    _loadTable();
  }

  Future<void> _loadTable() async {
    final token = await AuthRepository().getToken();
    if (mounted) {
      context.read<TableBloc>().add(TableLoadPlayers(widget.tableId, token: token));
    }
  }

  Future<void> _showJoinDialog() async {
    final token = await AuthRepository().getToken();
    if (!mounted || token == null) return;

    final confirmed = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (_) => const JoinTableDialog(),
    );

    if (confirmed == true && mounted) {
      context.read<TableBloc>().add(TableJoinRequested(widget.tableId, token));
    }
  }

  Future<void> _showLeaveDialog() async {
    final token = await AuthRepository().getToken();
    if (!mounted || token == null) return;

    final state = context.read<TableBloc>().state;
    int remainingChips = 0;
    if (state is TableJoined) {
      remainingChips = state.myPlayer.chips.toInt();
    }
    final returnedGold = remainingChips ~/ 10;

    final confirmed = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (_) => LeaveTableDialog(
        remainingChips: remainingChips,
        returnedGold: returnedGold,
      ),
    );

    if (confirmed == true && mounted) {
      context.read<TableBloc>().add(TableLeaveRequested(widget.tableId, token));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: BlocConsumer<TableBloc, TableState>(
        listener: (context, state) {
          if (state is TableError) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(state.message, style: const TextStyle(color: Colors.white)),
                backgroundColor: AppColors.foldRed,
              ),
            );
          }
        },
        builder: (context, state) {
          return OrientationBuilder(
            builder: (context, orientation) {
              return orientation == Orientation.portrait
                  ? _buildPortrait(context, state)
                  : _buildLandscape(context, state);
            },
          );
        },
      ),
    );
  }

  Widget _buildPortrait(BuildContext context, TableState state) {
    final players = _getPlayers(state);
    final isSpectating = state is TableSpectating || state is TableLoadingPlayers;
    final isJoining = state is TableJoining;
    final isLeaving = state is TableLeaving;
    final myPlayer = state is TableJoined ? state.myPlayer : null;

    return Container(
      color: AppColors.bg,
      child: SafeArea(
        child: Stack(
          children: [
            // Green felt
            Positioned.fill(
              child: Container(
                margin: const EdgeInsets.only(top: 0, bottom: 0),
                decoration: const BoxDecoration(
                  gradient: RadialGradient(
                    center: Alignment(0, -0.1),
                    radius: 0.8,
                    colors: [
                      AppColors.feltLight,
                      Color(0xFF1e6b42),
                      AppColors.feltDark,
                    ],
                    stops: [0.0, 0.4, 1.0],
                  ),
                ),
              ),
            ),
            // Players positioned
            ..._buildPlayers(players),
            // Pot info
            _buildPotInfo(),
            // Community cards
            _buildCommunityCards(),
            // Hero area (only when joined)
            if (myPlayer != null) _buildHero(myPlayer),
            // Action buttons (only when joined)
            if (myPlayer != null) _buildActionButtons(),
            // Join button (spectating)
            if (isSpectating)
              Positioned(
                bottom: 80,
                left: 0,
                right: 0,
                child: Center(child: _buildJoinButton()),
              ),
            // Toolbar
            _buildToolbar(myPlayer != null),
            // Loading overlay
            if (isJoining || isLeaving)
              Container(
                color: Colors.black.withOpacity(0.5),
                child: const Center(
                  child: CircularProgressIndicator(color: AppColors.gold),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildLandscape(BuildContext context, TableState state) {
    final players = _getPlayers(state);
    final isSpectating = state is TableSpectating || state is TableLoadingPlayers;
    final isJoining = state is TableJoining;
    final isLeaving = state is TableLeaving;
    final myPlayer = state is TableJoined ? state.myPlayer : null;

    return Container(
      color: AppColors.bg,
      child: SafeArea(
        child: Stack(
          children: [
            // Green felt
            Positioned.fill(
              child: Container(
                margin: const EdgeInsets.all(8),
                decoration: const BoxDecoration(
                  gradient: RadialGradient(
                    center: Alignment(0, 0),
                    radius: 0.9,
                    colors: [
                      AppColors.feltLight,
                      Color(0xFF1e6b42),
                      AppColors.feltDark,
                    ],
                    stops: [0.0, 0.4, 1.0],
                  ),
                ),
              ),
            ),
            // Top row: 4 players
            Positioned(
              top: 8,
              left: 0,
              right: 0,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: players
                    .where((p) => [0, 1, 7, 8].contains(p.seat))
                    .map((p) => _PlayerWidget(player: p))
                    .toList(),
              ),
            ),
            // Left side
            Positioned(
              left: 8,
              top: 80,
              bottom: 80,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: players
                    .where((p) => p.seat == 6)
                    .map((p) => _PlayerWidget(player: p))
                    .toList(),
              ),
            ),
            // Right side
            Positioned(
              right: 8,
              top: 80,
              bottom: 80,
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: players
                    .where((p) => [2, 3].contains(p.seat))
                    .map((p) => _PlayerWidget(player: p))
                    .toList(),
              ),
            ),
            // Center: pot + cards
            Positioned(
              top: 0,
              bottom: 0,
              left: 80,
              right: 80,
              child: Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 3),
                      decoration: BoxDecoration(
                        color: AppColors.surface.withOpacity(0.8),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(color: AppColors.gold),
                      ),
                      child: const Text(
                        '底池: 5.3BB',
                        style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: AppColors.goldBright),
                      ),
                    ),
                    const SizedBox(height: 8),
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        ..._community.map((c) => Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 2),
                          child: PokerCardWidget(card: c, width: 22, height: 30),
                        )),
                        ...List.generate(2, (_) => const Padding(
                          padding: EdgeInsets.symmetric(horizontal: 2),
                          child: PokerCardWidget(faceDown: true, width: 22, height: 30),
                        )),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '5/10 · 上限5000 · 第 128 手',
                      style: TextStyle(fontSize: 8, color: AppColors.gold.withOpacity(0.6)),
                    ),
                  ],
                ),
              ),
            ),
            // Bottom: hero + actions
            if (myPlayer != null)
              Positioned(
                bottom: 8,
                left: 16,
                child: _buildLandscapeHero(myPlayer),
              ),
            // Right side actions (joined)
            if (myPlayer != null)
              Positioned(
                bottom: 8,
                right: 16,
                child: Row(
                  children: [
                    ActionButton(
                      label: '弃牌',
                      icon: Icons.close,
                      bgColor: AppColors.foldRed,
                      size: 32,
                      onTap: () => context.read<TableBloc>().add(TablePlayerAction('fold')),
                    ),
                    const SizedBox(width: 6),
                    ActionButton(
                      label: '加分',
                      text: '+',
                      bgColor: AppColors.raiseNavy,
                      size: 34,
                      onTap: () => context.read<TableBloc>().add(TablePlayerAction('raise')),
                    ),
                    const SizedBox(width: 6),
                    ActionButton(
                      label: '跟分',
                      text: '2BB',
                      bgColor: AppColors.callGreen,
                      size: 32,
                      onTap: () => context.read<TableBloc>().add(TablePlayerAction('call')),
                    ),
                  ],
                ),
              ),
            // Join button (spectating)
            if (isSpectating)
              Positioned(
                bottom: 60,
                left: 0,
                right: 0,
                child: Center(child: _buildJoinButton()),
              ),
            // Toolbar
            _buildToolbar(myPlayer != null),
            // Loading overlay
            if (isJoining || isLeaving)
              Container(
                color: Colors.black.withOpacity(0.5),
                child: const Center(
                  child: CircularProgressIndicator(color: AppColors.gold),
                ),
              ),
          ],
        ),
      ),
    );
  }

  List<Player> _getPlayers(TableState state) {
    if (state is TableSpectating) return state.players;
    if (state is TableJoining) return state.players;
    if (state is TableJoined) return state.players;
    if (state is TableBetting) return [];
    if (state is TableShowdown) return [];
    return [];
  }

  List<Widget> _buildPlayers(List<Player> players) {
    final positions = [
      const Offset(0.5, 0.03),   // 0: top center
      const Offset(0.88, 0.08),  // 1: top right
      const Offset(0.96, 0.22),  // 2: right upper
      const Offset(0.92, 0.48),  // 3: right lower
      const Offset(0.70, 0.58),  // 4: bottom right
      const Offset(0.30, 0.58),  // 5: bottom left
      const Offset(0.08, 0.48),  // 6: left lower
      const Offset(0.04, 0.22),  // 7: left upper
      const Offset(0.12, 0.08),  // 8: top left
    ];

    final widgets = <Widget>[];

    // Active players
    for (final player in players.where((p) => p.seat != null && p.seat != 9)) {
      final pos = positions[player.seat!];
      widgets.add(Positioned(
        top: pos.dy * MediaQuery.of(context).size.height,
        left: pos.dx * MediaQuery.of(context).size.width,
        child: Transform.translate(
          offset: const Offset(-0.5, 0),
          child: _PlayerWidget(player: player),
        ),
      ));
    }

    // Empty seats
    final takenSeats = players.where((p) => p.seat != null && p.seat != 9).map((p) => p.seat!).toSet();
    for (int i = 0; i < 9; i++) {
      if (!takenSeats.contains(i)) {
        final pos = positions[i];
        widgets.add(Positioned(
          top: pos.dy * MediaQuery.of(context).size.height,
          left: pos.dx * MediaQuery.of(context).size.width,
          child: Transform.translate(
            offset: const Offset(-0.5, 0),
            child: _EmptySeatWidget(),
          ),
        ));
      }
    }

    return widgets;
  }

  Widget _buildJoinButton() {
    return GestureDetector(
      onTap: _showJoinDialog,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            colors: [AppColors.foldRed, Color(0xFFB8860B)],
          ),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: AppColors.gold, width: 1.5),
          boxShadow: [
            BoxShadow(
              color: AppColors.gold.withOpacity(0.3),
              blurRadius: 10,
            ),
          ],
        ),
        child: const Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.videogame_asset, color: AppColors.goldBright, size: 18),
            SizedBox(width: 6),
            Text(
              '加入游戏',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.bold,
                color: AppColors.goldBright,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPotInfo() {
    return Positioned(
      top: MediaQuery.of(context).size.height * 0.18,
      left: 0,
      right: 0,
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              _chipStack('0.5BB'),
              const SizedBox(width: 12),
              _chipStack('1.8BB'),
              const SizedBox(width: 12),
              _chipStack('1BB'),
              const SizedBox(width: 12),
              _chipStack('2BB'),
            ],
          ),
          const SizedBox(height: 4),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
            decoration: BoxDecoration(
              color: AppColors.surface.withOpacity(0.8),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: AppColors.gold),
              boxShadow: [
                BoxShadow(
                  color: AppColors.gold.withOpacity(0.15),
                  blurRadius: 10,
                ),
              ],
            ),
            child: const Text(
              '底池: 5.3BB',
              style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.bold,
                color: AppColors.goldBright,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _chipStack(String amount) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        const Text('🪙', style: TextStyle(fontSize: 10)),
        const SizedBox(width: 2),
        Text(
          amount,
          style: const TextStyle(fontSize: 9, color: AppColors.goldBright),
        ),
      ],
    );
  }

  Widget _buildCommunityCards() {
    return Positioned(
      top: MediaQuery.of(context).size.height * 0.38,
      left: 0,
      right: 0,
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              ..._community.map((c) => Padding(
                padding: const EdgeInsets.symmetric(horizontal: 3),
                child: PokerCardWidget(card: c, width: 24, height: 34),
              )),
              ...List.generate(2, (_) => const Padding(
                padding: EdgeInsets.symmetric(horizontal: 3),
                child: PokerCardWidget(faceDown: true, width: 24, height: 34),
              )),
            ],
          ),
          const SizedBox(height: 8),
          Opacity(
            opacity: 0.2,
            child: Column(
              children: [
                const Text(
                  '♠ 德扑之光',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: AppColors.goldBright,
                  ),
                ),
                Text(
                  'DEEP LIGHT',
                  style: TextStyle(
                    fontSize: 7,
                    color: AppColors.gold.withOpacity(0.6),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 4),
          Text(
            '5/10 · 上限5000 · 第 128 手',
            style: TextStyle(
              fontSize: 9,
              color: AppColors.gold.withOpacity(0.6),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHero(Player hero) {
    return Positioned(
      bottom: 80,
      left: 0,
      right: 0,
      child: Column(
        children: [
          Text(
            hero.name,
            style: const TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.bold,
              color: AppColors.goldBright,
              shadows: [
                Shadow(color: AppColors.gold, blurRadius: 8),
              ],
            ),
          ),
          const SizedBox(height: 4),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              PlayerAvatar(
                emoji: '👤',
                isActive: hero.isActive,
                timerText: hero.isActive ? '11S' : null,
                size: 36,
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: AppColors.surface.withOpacity(0.8),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: AppColors.gold.withOpacity(0.3)),
                ),
                child: Text(
                  '${hero.chips}BB',
                  style: const TextStyle(fontSize: 9, color: AppColors.goldBright),
                ),
              ),
              const SizedBox(width: 8),
              if (hero.holeCards != null)
                ...hero.holeCards!.map((c) => Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 2),
                  child: PokerCardWidget(card: c, width: 26, height: 38),
                )),
            ],
          ),
          const SizedBox(height: 4),
          const Text(
            '高牌',
            style: TextStyle(fontSize: 8, color: AppColors.textMuted),
          ),
        ],
      ),
    );
  }

  Widget _buildLandscapeHero(Player hero) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              hero.name,
              style: const TextStyle(
                fontSize: 10,
                fontWeight: FontWeight.bold,
                color: AppColors.goldBright,
                shadows: [
                  Shadow(color: AppColors.gold, blurRadius: 8),
                ],
              ),
            ),
            const SizedBox(height: 2),
            Row(
              children: [
                PlayerAvatar(
                  emoji: '👤',
                  isActive: hero.isActive,
                  timerText: hero.isActive ? '11S' : null,
                  size: 32,
                ),
                const SizedBox(width: 6),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: AppColors.surface.withOpacity(0.8),
                    borderRadius: BorderRadius.circular(6),
                    border: Border.all(color: AppColors.gold.withOpacity(0.3)),
                  ),
                  child: Text(
                    '${hero.chips}BB',
                    style: const TextStyle(fontSize: 8, color: AppColors.goldBright),
                  ),
                ),
              ],
            ),
          ],
        ),
        const SizedBox(width: 8),
        if (hero.holeCards != null)
          Row(
            children: hero.holeCards!.map((c) => Padding(
              padding: const EdgeInsets.symmetric(horizontal: 2),
              child: PokerCardWidget(card: c, width: 24, height: 34),
            )).toList(),
          ),
      ],
    );
  }

  Widget _buildActionButtons() {
    return Positioned(
      bottom: 12,
      left: 0,
      right: 0,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          ActionButton(
            label: '弃牌',
            icon: Icons.close,
            bgColor: AppColors.foldRed,
            size: 40,
            onTap: () => context.read<TableBloc>().add(TablePlayerAction('fold')),
          ),
          const SizedBox(width: 10),
          Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              _quickBetLabel('50%'),
              const SizedBox(height: 2),
              ActionButton(
                label: '精准加分',
                text: '+',
                bgColor: AppColors.raiseNavy,
                size: 44,
                onTap: () => context.read<TableBloc>().add(TablePlayerAction('raise')),
              ),
              const SizedBox(height: 2),
              _quickBetLabel('67%'),
            ],
          ),
          const SizedBox(width: 10),
          ActionButton(
            label: '跟分',
            text: '2BB',
            bgColor: AppColors.callGreen,
            size: 40,
            onTap: () => context.read<TableBloc>().add(TablePlayerAction('call')),
          ),
        ],
      ),
    );
  }

  Widget _quickBetLabel(String text) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: AppColors.surface.withOpacity(0.8),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppColors.gold.withOpacity(0.3)),
      ),
      child: Text(
        text,
        style: const TextStyle(fontSize: 8, color: AppColors.goldBright),
      ),
    );
  }

  Widget _buildToolbar(bool isJoined) {
    return Positioned(
      bottom: 12,
      left: 8,
      right: 8,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              _toolbarBtn(Icons.grid_view),
              const SizedBox(width: 6),
              _toolbarBtn(Icons.timer),
              if (isJoined) ...[
                const SizedBox(width: 6),
                _toolbarBtn(Icons.logout, onTap: _showLeaveDialog),
              ],
            ],
          ),
          Row(
            children: [
              _toolbarBtn(Icons.chat_bubble_outline),
              const SizedBox(width: 4),
              const Text('4', style: TextStyle(fontSize: 10, color: AppColors.goldBright)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _toolbarBtn(IconData icon, {VoidCallback? onTap}) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 26,
        height: 26,
        decoration: BoxDecoration(
          color: AppColors.surface.withOpacity(0.7),
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: AppColors.gold.withOpacity(0.3)),
        ),
        child: Icon(icon, size: 14, color: AppColors.goldBright),
      ),
    );
  }
}

class _PlayerWidget extends StatelessWidget {
  final Player player;
  const _PlayerWidget({required this.player});

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          player.name,
          style: TextStyle(
            fontSize: 9,
            color: AppColors.goldBright.withOpacity(player.hasFolded ? 0.5 : 1.0),
            shadows: const [Shadow(color: Colors.black, blurRadius: 4)],
          ),
        ),
        const SizedBox(height: 2),
        Stack(
          clipBehavior: Clip.none,
          children: [
            Opacity(
              opacity: player.hasFolded ? 0.5 : 1.0,
              child: PlayerAvatar(
                emoji: _emojiForPlayer(player.id),
                isActive: player.isActive,
                isDealer: player.isDealer,
                size: 30,
              ),
            ),
            if (player.statusTag != null)
              Positioned(
                top: -8,
                right: -8,
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 3, vertical: 1),
                  decoration: BoxDecoration(
                    color: AppColors.foldRed,
                    borderRadius: BorderRadius.circular(4),
                    border: Border.all(color: AppColors.gold),
                  ),
                  child: Text(
                    player.statusTag!,
                    style: const TextStyle(fontSize: 6, color: AppColors.goldBright),
                  ),
                ),
              ),
            if (player.hasFolded)
              Positioned(
                top: -8,
                left: 0,
                right: 0,
                child: Center(
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                    decoration: BoxDecoration(
                      color: AppColors.full.withOpacity(0.9),
                      borderRadius: BorderRadius.circular(4),
                      border: Border.all(color: AppColors.gold.withOpacity(0.2)),
                    ),
                    child: const Text(
                      '弃牌',
                      style: TextStyle(fontSize: 7, color: AppColors.textMuted),
                    ),
                  ),
                ),
              ),
          ],
        ),
        const SizedBox(height: 2),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
          decoration: BoxDecoration(
            color: AppColors.surface.withOpacity(0.8),
            borderRadius: BorderRadius.circular(6),
            border: Border.all(color: AppColors.gold.withOpacity(0.3)),
          ),
          child: Text(
            '${player.chips}BB',
            style: TextStyle(
              fontSize: 7,
              color: AppColors.goldBright.withOpacity(player.hasFolded ? 0.5 : 1.0),
            ),
          ),
        ),
      ],
    );
  }

  String _emojiForPlayer(String id) {
    const map = {
      'p1': '🧔',
      'p2': '🦇',
      'p3': '🦈',
      'p4': '👩',
      'p5': '👩‍🎤',
      'p6': '🌙',
    };
    return map[id] ?? '👤';
  }
}

class _EmptySeatWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const SizedBox(height: 11),
        Container(
          width: 30,
          height: 30,
          decoration: BoxDecoration(
            color: Colors.black.withOpacity(0.3),
            shape: BoxShape.circle,
            border: Border.all(color: AppColors.gold.withOpacity(0.2)),
          ),
          child: Icon(Icons.add, size: 14, color: AppColors.gold.withOpacity(0.3)),
        ),
        const SizedBox(height: 2),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
          decoration: BoxDecoration(
            color: Colors.black.withOpacity(0.3),
            borderRadius: BorderRadius.circular(6),
            border: Border.all(color: AppColors.gold.withOpacity(0.1)),
          ),
          child: Text(
            '空位',
            style: TextStyle(
              fontSize: 7,
              color: AppColors.gold.withOpacity(0.3),
            ),
          ),
        ),
      ],
    );
  }
}
