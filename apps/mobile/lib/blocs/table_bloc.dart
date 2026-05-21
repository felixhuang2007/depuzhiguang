import 'dart:async';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../repositories/game_repository.dart';

abstract class TableEvent {}

class TableConnect extends TableEvent {
  final String wsUrl;
  final String tableId;
  final String token;
  TableConnect(this.wsUrl, this.tableId, this.token);
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

abstract class TableState {}

class TableInitial extends TableState {}

class TableConnecting extends TableState {}

class TableConnected extends TableState {}

class TableJoined extends TableState {
  final Map<String, dynamic> tableState;
  TableJoined(this.tableState);
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

class TableDisconnected extends TableState {}

class TableError extends TableState {
  final String message;
  TableError(this.message);
}

class TableBloc extends Bloc<TableEvent, TableState> {
  final GameRepository _repo;
  StreamSubscription? _stateSub;

  TableBloc({GameRepository? repo})
      : _repo = repo ?? GameRepository(),
        super(TableInitial()) {
    on<TableConnect>(_onConnect);
    on<TableGameStateUpdated>(_onStateUpdate);
    on<TablePlayerAction>(_onAction);
    on<TableDisconnect>(_onDisconnect);
  }

  void _onConnect(TableConnect event, Emitter<TableState> emit) {
    emit(TableConnecting());
    _repo.connect(event.wsUrl, event.tableId, event.token);
    _stateSub = _repo.stateStream.listen((msg) {
      add(TableGameStateUpdated(msg));
    });
    emit(TableConnected());
  }

  void _onStateUpdate(TableGameStateUpdated event, Emitter<TableState> emit) {
    final msg = event.state;
    final type = msg['type'] as String?;
    switch (type) {
      case 'table_state':
        emit(TableJoined(msg));
      case 'your_turn':
        final timeout = msg['timeout'] as int? ?? 30;
        emit(TableBetting(msg, timeout: timeout));
      case 'showdown':
        emit(TableShowdown(msg));
      default:
        if (state is TableJoined || state is TableBetting || state is TableShowdown) {
          emit(TableJoined(msg));
        }
    }
  }

  void _onAction(TablePlayerAction event, Emitter<TableState> emit) {
    _repo.sendAction(event.action, amount: event.amount);
  }

  void _onDisconnect(TableDisconnect event, Emitter<TableState> emit) {
    _repo.disconnect();
    emit(TableDisconnected());
  }

  @override
  Future<void> close() {
    _stateSub?.cancel();
    _repo.disconnect();
    return super.close();
  }
}
