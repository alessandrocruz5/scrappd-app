import '../../core/storage/token_storage.dart';
import '../../domain/entities/user.dart';
import '../../domain/repositories/auth_repository.dart';
import '../datasources/auth_remote_datasource.dart';

class AuthRepositoryImpl implements AuthRepository {
  AuthRepositoryImpl({
    required AuthRemoteDataSource remoteDataSource,
    required TokenStorage tokenStorage,
  })  : _remoteDataSource = remoteDataSource,
        _tokenStorage = tokenStorage;

  final AuthRemoteDataSource _remoteDataSource;
  final TokenStorage _tokenStorage;

  @override
  bool get hasSession => _tokenStorage.accessToken?.isNotEmpty ?? false;

  @override
  Future<User> login({required String email, required String password}) async {
    final payload = await _remoteDataSource.login(
      email: email,
      password: password,
    );

    await _tokenStorage.saveTokens(
      accessToken: payload.accessToken,
      refreshToken: payload.refreshToken,
    );

    return payload.user;
  }

  @override
  Future<User> register({
    required String email,
    required String username,
    required String password,
    String? displayName,
  }) async {
    final payload = await _remoteDataSource.register(
      email: email,
      username: username,
      password: password,
      displayName: displayName,
    );

    await _tokenStorage.saveTokens(
      accessToken: payload.accessToken,
      refreshToken: payload.refreshToken,
    );

    return payload.user;
  }

  @override
  Future<User> getMe() {
    return _remoteDataSource.getMe();
  }

  @override
  Future<void> logout() async {
    final refreshToken = _tokenStorage.refreshToken;
    if (refreshToken != null && refreshToken.isNotEmpty) {
      await _remoteDataSource.logout(refreshToken: refreshToken);
    }
    await _tokenStorage.clearTokens();
  }
}
