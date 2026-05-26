import 'dart:async';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../models/player.dart';
import '../repositories/game_repository.dart';
import '../repositories/table_repository.dart';

// ── Events ──────────────────────────────────────────────────────

abstract class TableEvent {}

class TableConnect extends TableEvent {
  final String wsUrl;
  final String tableId;
  final String token;
  TableConnect(this.wsUrl, this.tableId, this.token);
}

class TableLoadPlayers extends TableEvent {
  final String tableId;
  final String? token;
  TableLoadPlayers(this.tableId, {this.token});
}

class TableJoinRequested extends TableEvent {
  final String tableId;
  final String token;
  final int? seat;
  final int? chips;
  TableJoinRequested(this.tableId, this.token, {this.seat, this.chips});
}

class TableLeaveRequested extends TableEvent {
  final String tableId;
  final String token;
  TableLeaveRequested(this.tableId, this.token);
}

class TableGameStateUpdated extends TableEvent {
  final Map<String, dynamic> state;
  TableGameStateUpdated(this.state);
}

class TablePlayerAction extends TableEvent {
  final String action;
  final int? amount;
  TablePlayerAction(this.action, {this.amount});
}

class TableDisconnect extends TableEvent {}

class _TableConnectionError extends TableEvent {
  final String message;
  _TableConnectionError(this.message);
}

// ── States ──────────────────────────────────────────────────────

abstract class TableState {}

class TableInitial extends TableState {}

class TableLoadingPlayers extends TableState {}

class TableSpectating extends TableState {
  final List<Player> players;
  final Map<String, dynamic>? tableState;
  TableSpectating(this.players, {this.tableState});
}

class TableJoining extends TableState {
  final List<Player> players;
  TableJoining(this.players);
}

class TableJoined extends TableState {
  final Map<String, dynamic> tableState;
  final List<Player> players;
  final Player myPlayer;
  TableJoined(this.tableState, {required this.players, required this.myPlayer});
}

class TableBetting extends TableState {
  final Map<String, dynamic> tableState;
  final int timeout;
  TableBetting(this.tableState, {this.timeout = 30});
}

class TableShowdown extends TableState {
  final Map<String, dynamic> tableState;
  TableShowdown(this.tableState);
}

class TableLeaving extends TableState {}

class TableDisconnected extends TableState {}

class TableError extends TableState {
  final String message;
  TableError(this.message);
}

// ── Bloc ────────────────────────────────────────────────────────

class TableBloc extends Bloc<TableEvent, TableState> {
  final GameRepository _gameRepo;
  final TableRepository _tableRepo;
  StreamSubscription? _stateSub;

  TableBloc({GameRepository? gameRepo, TableRepository? tableRepo})
      : _gameRepo = gameRepo ?? GameRepository(),
        _tableRepo = tableRepo ?? TableRepository(),
        super(TableInitial()) {
    on<TableConnect>(_onConnect);
    on<TableLoadPlayers>(_onLoadPlayers);
    on<TableJoinRequested>(_onJoin);
    on<TableLeaveRequested>(_onLeave);
    on<TableGameStateUpdated>(_onStateUpdate);
    on<TablePlayerAction>(_onAction);
    on<TableDisconnect>(_onDisconnect);
    on<_TableConnectionError>(_onConnectionError);
  }

  Future<void> _onLoadPlayers(TableLoadPlayers event, Emitter<TableState> emit) async {
    emit(TableLoadingPlayers());
    try {
      final players = await _tableRepo.getTablePlayers(event.tableId);
      Player? myPlayer;
      if (event.token != null) {
        myPlayer = await _tableRepo.getMyPlayer(event.tableId, event.token!);
      }

      if (myPlayer != null) {
        // Already seated — connect WS and show joined state
        final allPlayers = _mergePlayers(players, myPlayer);
        emit(TableJoined(
          {},
          players: allPlayers,
          myPlayer: myPlayer,
        ));
        _connectWs('ws://localhost:8080/ws', event.tableId, event.token!);
      } else {
        // Spectating
        emit(TableSpectating(players));
      }
    } catch (e) {
      emit(TableError('Failed to load table: $e'));
    }
  }

  Future<void> _onJoin(TableJoinRequested event, Emitter<TableState> emit) async {
    final current = state;
    if (current is TableSpectating) {
      emit(TableJoining(current.players));
    }
    try {
      await _tableRepo.joinTable(event.tableId, event.token, seat: event.seat, chips: event.chips);
      final players = await _tableRepo.getTablePlayers(event.tableId);
      final myPlayer = await _tableRepo.getMyPlayer(event.tableId, event.token);
      if (myPlayer != null) {
        final allPlayers = _mergePlayers(players, myPlayer);
        emit(TableJoined(
          {},
          players: allPlayers,
          myPlayer: myPlayer,
        ));
        _connectWs('ws://localhost:8080/ws', event.tableId, event.token);
      }
    } catch (e) {
      if (current is TableSpectating) {
        emit(TableSpectating(current.players));
      }
      emit(TableError(e.toString()));
    }
  }

  Future<void> _onLeave(TableLeaveRequested event, Emitter<TableState> emit) async {
    emit(TableLeaving());
    try {
      await _tableRepo.leaveTable(event.tableId, event.token);
      _gameRepo.disconnect();
      _stateSub?.cancel();
      final players = await _tableRepo.getTablePlayers(event.tableId);
      emit(TableSpectating(players));
    } catch (e) {
      emit(TableError(e.toString()));
    }
  }

  void _onConnect(TableConnect event, Emitter<TableState> emit) {
    _connectWs(event.wsUrl, event.tableId, event.token);
  }

  void _connectWs(String url, String tableId, String token) {
    try {
      _gameRepo.connect(url, tableId, token);
      _stateSub = _gameRepo.stateStream.listen(
        (msg) => add(TableGameStateUpdated(msg)),
        onError: (error) => add(_TableConnectionError(error.toString())),
      );
    } catch (e) {
      // Silently ignore WS connection errors in test/local environments
      add(_TableConnectionError(e.toString()));
    }
  }

  void _onConnectionError(_TableConnectionError event, Emitter<TableState> emit) {
    emit(TableError(event.message));
  }

  void _onStateUpdate(TableGameStateUpdated event, Emitter<TableState> emit) {
    final msg = event.state;
    final type = msg['type'] as String?;
    switch (type) {
      case 'table_state':
        _emitGameState(msg, emit);
      case 'your_turn':
        final timeout = msg['timeout'] as int? ?? 30;
        _emitGameState(msg, emit, isBetting: true, timeout: timeout);
      case 'showdown':
        _emitGameState(msg, emit, isShowdown: true);
      default:
        if (state is TableJoined || state is TableBetting || state is TableShowdown) {
          _emitGameState(msg, emit);
        }
    }
  }

  void _emitGameState(
    Map<String, dynamic> msg,
    Emitter<TableState> emit, {
    bool isBetting = false,
    bool isShowdown = false,
    int timeout = 30,
  }) {
    final current = state;
    List<Player> players = [];
    Player? myPlayer;

    if (current is TableJoined) {
      players = current.players;
      myPlayer = current.myPlayer;
    } else if (current is TableBetting) {
      players = [];
    } else if (current is TableShowdown) {
      players = [];
    }

    if (isBetting) {
      emit(TableBetting(msg, timeout: timeout));
    } else if (isShowdown) {
      emit(TableShowdown(msg));
    } else {
      if (myPlayer != null) {
        emit(TableJoined(msg, players: players, myPlayer: myPlayer));
      } else {
        emit(TableSpectating(players, tableState: msg));
      }
    }
  }

  List<Player> _mergePlayers(List<Player> apiPlayers, Player myPlayer) {
    final map = {for (var p in apiPlayers) p.seat: p};
    map[myPlayer.seat] = myPlayer;
    return map.values.toList();
  }

  void _onAction(TablePlayerAction event, Emitter<TableState> emit) {
    _gameRepo.sendAction(event.action, amount: event.amount);
  }

  void _onDisconnect(TableDisconnect event, Emitter<TableState> emit) {
    _gameRepo.disconnect();
    _stateSub?.cancel();
    emit(TableDisconnected());
  }

  @override
  Future<void> close() {
    _stateSub?.cancel();
    _gameRepo.disconnect();
    return super.close();
  }
}
