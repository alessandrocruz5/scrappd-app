import 'package:flutter/material.dart';

import '../../core/network/error_helpers.dart';
import '../../core/storage/token_storage.dart';
import '../../domain/entities/user.dart';
import '../../domain/repositories/auth_repository.dart';

enum AuthStatus { unknown, authenticated, unauthenticated }

class AuthProvider extends ChangeNotifier {
  AuthProvider(this._authRepository, this._tokenStorage) {
    _tokenStorage.addListener(_handleTokenChange);
  }

  final AuthRepository _authRepository;
  final TokenStorage _tokenStorage;

  AuthStatus _status = AuthStatus.unknown;
  User? _user;
  bool _isLoading = false;
  String? _errorMessage;

  AuthStatus get status => _status;
  User? get user => _user;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;
  bool get hasSession => _authRepository.hasSession;

  Future<void> initialize() async {
    if (_status != AuthStatus.unknown) return;

    if (!_authRepository.hasSession) {
      _status = AuthStatus.unauthenticated;
      notifyListeners();
      return;
    }

    try {
      _setLoading(true);
      _user = await _authRepository.getMe();
      _status = AuthStatus.authenticated;
    } catch (_) {
      await _tokenStorage.clearTokens();
      _status = AuthStatus.unauthenticated;
    } finally {
      _setLoading(false);
    }
  }

  Future<void> login({required String email, required String password}) async {
    _setLoading(true);
    _errorMessage = null;
    try {
      _user = await _authRepository.login(email: email, password: password);
      _status = AuthStatus.authenticated;
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Login failed. Please try again.',
      );
      _status = AuthStatus.unauthenticated;
    } finally {
      _setLoading(false);
    }
  }

  Future<void> register({
    required String email,
    required String username,
    required String password,
    String? displayName,
  }) async {
    _setLoading(true);
    _errorMessage = null;
    try {
      _user = await _authRepository.register(
        email: email,
        username: username,
        password: password,
        displayName: displayName,
      );
      _status = AuthStatus.authenticated;
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Registration failed. Please try again.',
      );
      _status = AuthStatus.unauthenticated;
    } finally {
      _setLoading(false);
    }
  }

  Future<void> requestPasswordReset({required String email}) async {
    _setLoading(true);
    _errorMessage = null;
    try {
      await _authRepository.requestPasswordReset(email: email);
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Failed to request password reset.',
      );
    } finally {
      _setLoading(false);
    }
  }

  Future<void> logout() async {
    _setLoading(true);
    try {
      await _authRepository.logout();
    } finally {
      _user = null;
      _status = AuthStatus.unauthenticated;
      _setLoading(false);
    }
  }

  void _setLoading(bool value) {
    _isLoading = value;
    notifyListeners();
  }

  void _handleTokenChange() {
    final hasAccessToken = _tokenStorage.accessToken?.isNotEmpty ?? false;
    if (!hasAccessToken && _status != AuthStatus.unauthenticated) {
      _user = null;
      _status = AuthStatus.unauthenticated;
      notifyListeners();
    }
  }

  @override
  void dispose() {
    _tokenStorage.removeListener(_handleTokenChange);
    super.dispose();
  }
}
