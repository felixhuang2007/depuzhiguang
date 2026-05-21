import 'package:shared_preferences/shared_preferences.dart';

class AuthRepository {
  static const _tokenKey = 'auth_token';

  Future<String?> getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_tokenKey);
  }

  Future<void> saveToken(String token) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_tokenKey, token);
  }

  Future<void> clearToken() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
  }

  Future<bool> login(String username, String password) async {
    // TODO: call REST API
    await Future.delayed(const Duration(seconds: 1));
    await saveToken('mock_token_$username');
    return true;
  }
}
