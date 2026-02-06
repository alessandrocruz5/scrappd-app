import 'package:dio/dio.dart';

import '../../core/constants/api_constants.dart';
import '../../core/models/api_response.dart';
import '../models/user_model.dart';

class LoginPayload {
  final UserModel user;
  final String accessToken;
  final String refreshToken;

  LoginPayload({
    required this.user,
    required this.accessToken,
    required this.refreshToken,
  });
}

class AuthRemoteDataSource {
  AuthRemoteDataSource(this._dio);

  final Dio _dio;

  Future<LoginPayload> login({
    required String email,
    required String password,
  }) async {
    final response = await _dio.post(
      ApiConstants.authLogin,
      data: {
        'email': email,
        'password': password,
      },
    );

    return _parseLoginPayload(response.data as Map<String, dynamic>);
  }

  Future<LoginPayload> register({
    required String email,
    required String username,
    required String password,
    String? displayName,
  }) async {
    final response = await _dio.post(
      ApiConstants.authRegister,
      data: {
        'email': email,
        'username': username,
        'password': password,
        if (displayName != null && displayName.isNotEmpty)
          'display_name': displayName,
      },
    );
    final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
      response.data as Map<String, dynamic>,
      (data) => data as Map<String, dynamic>,
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Registration failed');
    }

    // Backend register returns user only; follow up with login to get tokens.
    return login(email: email, password: password);
  }

  Future<UserModel> getMe() async {
    final response = await _dio.get(ApiConstants.authMe);
    final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
      response.data as Map<String, dynamic>,
      (data) => data as Map<String, dynamic>,
    );

    if (apiResponse.success && apiResponse.data != null) {
      return UserModel.fromJson(apiResponse.data!);
    }

    throw Exception(apiResponse.error?.message ?? 'Failed to load user');
  }

  Future<void> logout({required String refreshToken}) async {
    final response = await _dio.post(
      ApiConstants.authLogout,
      data: {'refresh_token': refreshToken},
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (_) => null,
    );

    if (!apiResponse.success) {
      throw Exception(apiResponse.error?.message ?? 'Logout failed');
    }
  }

  Future<void> requestPasswordReset({required String email}) async {
    final response = await _dio.post(
      ApiConstants.authForgotPassword,
      data: {'email': email},
    );
    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (_) => null,
    );
    if (!apiResponse.success) {
      throw Exception(apiResponse.error?.message ?? 'Failed to request password reset');
    }
  }

  Future<void> resetPassword({
    required String token,
    required String password,
  }) async {
    final response = await _dio.post(
      ApiConstants.authResetPassword,
      data: {
        'token': token,
        'password': password,
      },
    );
    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (_) => null,
    );
    if (!apiResponse.success) {
      throw Exception(apiResponse.error?.message ?? 'Failed to reset password');
    }
  }

  Future<void> resendVerification({required String email}) async {
    final response = await _dio.post(
      ApiConstants.authResendVerification,
      data: {'email': email},
    );
    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (_) => null,
    );
    if (!apiResponse.success) {
      throw Exception(apiResponse.error?.message ?? 'Failed to resend verification email');
    }
  }

  LoginPayload _parseLoginPayload(Map<String, dynamic> json) {
    final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
      json,
      (data) => data as Map<String, dynamic>,
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Authentication failed');
    }

    final data = apiResponse.data!;
    return LoginPayload(
      user: UserModel.fromJson(data['user'] as Map<String, dynamic>),
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
    );
  }
}
