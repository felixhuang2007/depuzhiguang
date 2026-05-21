import 'dart:async';
import '../services/websocket_service.dart';

class GameRepository {
  final WebSocketService _ws;
  final _stateController = StreamController<Map<String, dynamic>>.broadcast();

  Stream<Map<String, dynamic>> get stateStream => _stateController.stream;

  GameRepository({WebSocketService? ws}) : _ws = ws ?? WebSocketService();

  void connect(String url, String tableId, String token) {
    _ws.connect(url);
    _ws.messages.listen(
      (msg) {
        _stateController.add(msg);
      },
      onError: (error) {
        _stateController.addError(error);
      },
    );
    _ws.send({
      'type': 'join_table',
      'payload': {'table_id': tableId, 'token': token},
    });
  }

  void sendAction(String action, {int? amount}) {
    _ws.send({
      'type': 'action',
      'payload': {
        'action': action,
        if (amount != null) 'amount': amount,
      },
    });
  }

  void disconnect() {
    _ws.disconnect();
  }
}
