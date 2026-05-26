import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/player.dart';

class TableRepository {
  static const String _baseUrl = 'http://43.163.117.74:3000/api';

  Future<List<Player>> getTablePlayers(String tableId) async {
    final response = await http.get(Uri.parse('$_baseUrl/tables/$tableId/players'));
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      final players = data['players'] as List<dynamic>;
      return players.map((p) => _mapToPlayer(p as Map<String, dynamic>)).toList();
    }
    throw Exception('Failed to load players: ${response.statusCode}');
  }

  Future<Player?> getMyPlayer(String tableId, String token) async {
    final response = await http.get(
      Uri.parse('$_baseUrl/tables/$tableId/me'),
      headers: {'Authorization': 'Bearer $token'},
    );
    if (response.statusCode == 200) {
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      if (data['seated'] == true && data['player'] != null) {
        return _mapToPlayer(data['player'] as Map<String, dynamic>, isHero: true);
      }
    }
    return null;
  }

  Future<Map<String, dynamic>> joinTable(String tableId, String token, {int? seat, int? chips}) async {
    final response = await http.post(
      Uri.parse('$_baseUrl/tables/$tableId/join'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({'seat': seat, 'chips': chips}),
    );
    if (response.statusCode == 201) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }
    final error = jsonDecode(response.body) as Map<String, dynamic>;
    throw Exception(error['error'] ?? 'Failed to join table: ${response.statusCode}');
  }

  Future<Map<String, dynamic>> leaveTable(String tableId, String token) async {
    final response = await http.post(
      Uri.parse('$_baseUrl/tables/$tableId/leave'),
      headers: {'Authorization': 'Bearer $token'},
    );
    if (response.statusCode == 200) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }
    final error = jsonDecode(response.body) as Map<String, dynamic>;
    throw Exception(error['error'] ?? 'Failed to leave table: ${response.statusCode}');
  }

  Player _mapToPlayer(Map<String, dynamic> p, {bool isHero = false}) {
    return Player(
      id: p['userId'] ?? p['id'] ?? '',
      name: p['nickname'] ?? p['username'] ?? 'Player',
      chips: (p['chips'] as num?)?.toDouble() ?? 0.0,
      seat: p['seat'] as int?,
      avatar: p['avatar'] as String?,
      isHero: isHero,
    );
  }
}
