import 'package:flutter/foundation.dart';
import '../../data/models/user.dart';
import '../../data/services/auth_service.dart';
import '../../data/services/secure_storage_service.dart';

enum AuthState {
  initial,
  loading,
  authenticated,
  unauthenticated,
  error,
}

class AuthProvider extends ChangeNotifier {
  final AuthService _authService;
  final SecureStorageService _storageService;

  AuthState _state = AuthState.initial;
  User? _user;
  String? _errorMessage;

  AuthProvider(this._authService, this._storageService);

  // Getters
  AuthState get state => _state;
  User? get user => _user;
  String? get errorMessage => _errorMessage;
  bool get isAuthenticated => _state == AuthState.authenticated;
  bool get isLoading => _state == AuthState.loading;

  // Initialize auth state (call on app startup)
  Future<void> initialize() async {
    _setState(AuthState.loading);

    try {
      final isLoggedIn = await _storageService.isLoggedIn();
      if (isLoggedIn) {
        // Try to get cached user
        _user = await _storageService.getUser();
        if (_user != null) {
          _setState(AuthState.authenticated);
          // Refresh user data in background
          _refreshUserInBackground();
        } else {
          // No cached user, try to fetch
          await _fetchCurrentUser();
        }
      } else {
        _setState(AuthState.unauthenticated);
      }
    } catch (e) {
      debugPrint('Auth initialization error: $e');
      _setState(AuthState.unauthenticated);
    }
  }

  void _refreshUserInBackground() async {
    try {
      final user = await _authService.getCurrentUser();
      _user = user;
      notifyListeners();
    } catch (e) {
      debugPrint('Background user refresh failed: $e');
    }
  }

  Future<void> _fetchCurrentUser() async {
    try {
      _user = await _authService.getCurrentUser();
      _setState(AuthState.authenticated);
    } catch (e) {
      await _storageService.clearAuthData();
      _setState(AuthState.unauthenticated);
    }
  }

  // Login
  Future<bool> login(String email, String password) async {
    _clearError();
    _setState(AuthState.loading);

    try {
      final response = await _authService.login(LoginRequest(
        email: email,
        password: password,
      ));
      _user = response.user;
      _setState(AuthState.authenticated);
      return true;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(AuthState.unauthenticated);
      return false;
    }
  }

  // Register
  Future<bool> register(
    String email,
    String username,
    String password, {
    String? displayName,
  }) async {
    _clearError();
    _setState(AuthState.loading);

    try {
      final response = await _authService.register(RegisterRequest(
        email: email,
        username: username,
        password: password,
        displayName: displayName,
      ));
      _user = response.user;
      _setState(AuthState.authenticated);
      return true;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(AuthState.unauthenticated);
      return false;
    }
  }

  // Logout
  Future<void> logout() async {
    _setState(AuthState.loading);

    try {
      await _authService.logout();
    } catch (e) {
      debugPrint('Logout error: $e');
    } finally {
      _user = null;
      _setState(AuthState.unauthenticated);
    }
  }

  // Refresh user data
  Future<void> refreshUser() async {
    if (_state != AuthState.authenticated) return;

    try {
      _user = await _authService.getCurrentUser();
      notifyListeners();
    } catch (e) {
      debugPrint('User refresh error: $e');
    }
  }

  void _setState(AuthState newState) {
    _state = newState;
    notifyListeners();
  }

  void _setError(String message) {
    _errorMessage = message;
    notifyListeners();
  }

  void _clearError() {
    _errorMessage = null;
  }

  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }
}
