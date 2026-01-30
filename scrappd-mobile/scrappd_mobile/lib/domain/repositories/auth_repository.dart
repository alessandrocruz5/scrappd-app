import '../entities/user.dart';

abstract class AuthRepository {
  Future<User> login({required String email, required String password});
  Future<User> register({
    required String email,
    required String username,
    required String password,
    String? displayName,
  });
  Future<User> getMe();
  Future<void> logout();
  bool get hasSession;
}
