import 'package:flutter_bloc/flutter_bloc.dart';
import '../models/table_info.dart';
import '../repositories/lobby_repository.dart';

abstract class LobbyEvent {}

class LobbyLoadRequested extends LobbyEvent {
  final String type;
  LobbyLoadRequested({this.type = 'cash'});
}

class LobbyFilterChanged extends LobbyEvent {
  final String filter;
  LobbyFilterChanged(this.filter);
}

abstract class LobbyState {}

class LobbyInitial extends LobbyState {}

class LobbyLoading extends LobbyState {}

class LobbyLoaded extends LobbyState {
  final List<TableInfo> tables;
  final String activeFilter;
  LobbyLoaded(this.tables, {this.activeFilter = 'cash'});
}

class LobbyError extends LobbyState {
  final String message;
  LobbyError(this.message);
}

class LobbyBloc extends Bloc<LobbyEvent, LobbyState> {
  final LobbyRepository _repo;

  LobbyBloc({LobbyRepository? repo})
      : _repo = repo ?? LobbyRepository(),
        super(LobbyInitial()) {
    on<LobbyLoadRequested>(_onLoad);
    on<LobbyFilterChanged>(_onFilter);
  }

  Future<void> _onLoad(LobbyLoadRequested event, Emitter<LobbyState> emit) async {
    emit(LobbyLoading());
    try {
      final tables = await _repo.fetchTables(event.type);
      emit(LobbyLoaded(tables, activeFilter: event.type));
    } catch (e) {
      emit(LobbyError(e.toString()));
    }
  }

  Future<void> _onFilter(LobbyFilterChanged event, Emitter<LobbyState> emit) async {
    emit(LobbyLoading());
    try {
      final tables = await _repo.fetchTables(event.filter);
      emit(LobbyLoaded(tables, activeFilter: event.filter));
    } catch (e) {
      emit(LobbyError(e.toString()));
    }
  }
}
