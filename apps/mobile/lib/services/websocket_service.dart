import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';

class WebSocketService {
  WebSocketChannel? _channel;
  final _messageController = StreamController<Map<String, dynamic>>.broadcast();
  final _connectionController = StreamController<bool>.broadcast();
  Timer? _reconnectTimer;
  bool _shouldReconnect = true;
  String? _url;
  int _reconnectDelay = 1;

  Stream<Map<String, dynamic>> get messages => _messageController.stream;
  Stream<bool> get connectionState => _connectionController.stream;

  void connect(String url) {
    _shouldReconnect = true;
    _url = url;
    _reconnectDelay = 1;
    _connect(url);
  }

  void _connect(String url) {
    try {
      _channel = WebSocketChannel.connect(Uri.parse(url));
      _connectionController.add(true);
      _reconnectDelay = 1; // reset on success
      _channel!.stream.listen(
        (data) {
          try {
            final msg = jsonDecode(data as String);
            _messageController.add(msg);
          } catch (_) {
            // ignore malformed messages
          }
        },
        onError: (_) => _scheduleReconnect(),
        onDone: () => _scheduleReconnect(),
      );
    } catch (_) {
      _scheduleReconnect();
    }
  }

  void _scheduleReconnect() {
    _connectionController.add(false);
    if (!_shouldReconnect || _url == null) return;
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(Duration(seconds: _reconnectDelay), () {
      _connect(_url!);
    });
    // Exponential backoff capped at 30s
    _reconnectDelay = (_reconnectDelay * 2).clamp(1, 30);
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
