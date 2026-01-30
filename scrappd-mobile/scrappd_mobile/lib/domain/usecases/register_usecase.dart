import '../entities/user.dart';
import '../repositories/auth_repository.dart';

class RegisterUseCase {
  RegisterUseCase(this._repository);

  final AuthRepository _repository;

  Future<User> call({
    required String email,
    required String username,
    required String password,
    String? displayName,
  }) {
    return _repository.register(
      email: email,
      username: username,
      password: password,
      displayName: displayName,
    );
  }
}
