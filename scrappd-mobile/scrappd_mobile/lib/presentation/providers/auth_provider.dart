import 'package:flutter/material.dart';

import '../../domain/entities/user.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../core/network/error_helpers.dart';

enum AuthStatus {
  unknown,
  authenticated,
  unauthenticated,
}

class AuthProvider extends ChangeNotifier {
  AuthProvider(this._authRepository);

  final AuthRepository _authRepository;

  AuthStatus _status = AuthStatus.unknown;
  User? _user;
  bool _isLoading = false;
  String? _errorMessage;

  AuthStatus get status => _status;
  User? get user => _user;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;

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
}
