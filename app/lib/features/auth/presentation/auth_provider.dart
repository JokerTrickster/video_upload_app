import 'package:flutter/foundation.dart';
import '../../../core/api/api_client.dart';
import '../../../shared/models/user_model.dart';
import '../data/auth_repository.dart';

class AuthProvider extends ChangeNotifier {
  final AuthRepository _authRepository;
  final ApiClient _apiClient;

  UserModel? _user;
  bool _isLoading = false;
  String? _error;
  bool _isAuthenticated = false;

  AuthProvider(this._authRepository, this._apiClient);

  UserModel? get user => _user;
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isAuthenticated => _isAuthenticated;

  Future<void> checkAuthStatus() async {
    final hasToken = await _apiClient.hasToken();
    if (hasToken) {
      try {
        _user = await _authRepository.getCurrentUser();
        _isAuthenticated = true;
      } catch (_) {
        _isAuthenticated = false;
        await _apiClient.clearTokens();
      }
    }
    notifyListeners();
  }

  Future<String> getGoogleAuthUrl() async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      final url = await _authRepository.getGoogleAuthUrl();
      return url;
    } catch (e) {
      _error = 'Failed to get auth URL: $e';
      notifyListeners();
      rethrow;
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> handleCallback(String code, String state) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      final authResponse = await _authRepository.handleCallback(code, state);
      await _apiClient.saveTokens(
        authResponse.tokens.accessToken,
        authResponse.tokens.refreshToken,
      );
      _user = authResponse.user;
      _isAuthenticated = true;
    } catch (e) {
      _error = 'Authentication failed: $e';
      _isAuthenticated = false;
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> logout() async {
    _isLoading = true;
    notifyListeners();

    try {
      await _authRepository.logout();
    } catch (_) {
      // Logout even if API call fails
    } finally {
      _user = null;
      _isAuthenticated = false;
      _isLoading = false;
      await _apiClient.clearTokens();
      notifyListeners();
    }
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }
}
