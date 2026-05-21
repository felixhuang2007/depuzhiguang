import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import '../blocs/lobby_bloc.dart';
import '../theme.dart';
import '../widgets/lobby_card.dart';
import 'table_screen.dart';

class LobbyScreen extends StatelessWidget {
  const LobbyScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return BlocProvider(
      create: (_) => LobbyBloc()..add(LobbyLoadRequested()),
      child: BlocBuilder<LobbyBloc, LobbyState>(
        builder: (context, state) {
          return Column(
            children: [
              // Filter tabs
              _FilterTabs(l10n: l10n),
              const SizedBox(height: 8),
              // Table list
              Expanded(
                child: _buildBody(context, state, l10n),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildBody(BuildContext context, LobbyState state, AppLocalizations l10n) {
    if (state is LobbyLoading) {
      return const Center(
        child: CircularProgressIndicator(color: AppColors.gold),
      );
    }
    if (state is LobbyError) {
      return Center(
        child: Text(state.message, style: const TextStyle(color: AppColors.goldBright)),
      );
    }
    if (state is LobbyLoaded) {
      return ListView.separated(
        padding: const EdgeInsets.all(12),
        itemCount: state.tables.length,
        separatorBuilder: (_, __) => const SizedBox(height: 8),
        itemBuilder: (context, i) {
          final table = state.tables[i];
          return LobbyCard(
            table: table,
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (_) => TableScreen(tableId: table.id),
              ),
            ),
          );
        },
      );
    }
    return const SizedBox.shrink();
  }
}

class _FilterTabs extends StatelessWidget {
  final AppLocalizations l10n;
  const _FilterTabs({required this.l10n});

  @override
  Widget build(BuildContext context) {
    final filters = [
      ('cash', l10n.cashGame),
      ('sng', l10n.sng),
      ('tournament', l10n.tournament),
      ('training', l10n.training),
    ];

    return BlocBuilder<LobbyBloc, LobbyState>(
      builder: (context, state) {
        final active = state is LobbyLoaded ? state.activeFilter : 'cash';
        return Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
          color: AppColors.bg,
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: filters.map((f) {
                final isActive = f.$1 == active;
                return Padding(
                  padding: const EdgeInsets.only(right: 6),
                  child: GestureDetector(
                    onTap: () => context.read<LobbyBloc>().add(LobbyFilterChanged(f.$1)),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 5),
                      decoration: BoxDecoration(
                        color: isActive ? AppColors.foldRed : AppColors.surface.withOpacity(0.6),
                        borderRadius: BorderRadius.circular(14),
                        border: Border.all(
                          color: isActive ? AppColors.gold : AppColors.gold.withOpacity(0.3),
                        ),
                      ),
                      child: Text(
                        f.$2,
                        style: TextStyle(
                          fontSize: 11,
                          color: isActive ? AppColors.goldBright : AppColors.textMuted,
                          fontWeight: isActive ? FontWeight.bold : FontWeight.normal,
                        ),
                      ),
                    ),
                  ),
                );
              }).toList(),
            ),
          ),
        );
      },
    );
  }
}
