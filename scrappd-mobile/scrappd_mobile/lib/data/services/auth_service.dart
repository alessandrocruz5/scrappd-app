import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../../core/constants/api_constants.dart';
import '../models/user.dart';
import 'secure_storage_service.dart';

class AuthService {
  late final Dio _dio;
  final SecureStorageService _storageService;

  AuthService(this._storageService) {
    _dio = Dio(BaseOptions(
      baseUrl: ApiConstants.baseUrl,
      connectTimeout: ApiConstants.connectionTimeout,
      receiveTimeout: ApiConstants.receiveTimeout,
      sendTimeout: ApiConstants.sendTimeout,
      headers: ApiConstants.defaultHeaders,
    ));

    // Add interceptors for logging
    if (kDebugMode) {
      _dio.interceptors.add(LogInterceptor(
        requestBody: true,
        responseBody: true,
        error: true,
      ));
    }

    // Add auth interceptor
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await _storageService.getAccessToken();
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (error, handler) async {
        if (error.response?.statusCode == 401) {
          // Token expired, try to refresh
          final refreshed = await _tryRefreshToken();
          if (refreshed) {
            // Retry the request
            final token = await _storageService.getAccessToken();
            error.requestOptions.headers['Authorization'] = 'Bearer $token';
            try {
              final response = await _dio.fetch(error.requestOptions);
              return handler.resolve(response);
            } catch (e) {
              return handler.next(error);
            }
          }
        }
        handler.next(error);
      },
    ));
  }

  Future<bool> _tryRefreshToken() async {
    try {
      final refreshToken = await _storageService.getRefreshToken();
      if (refreshToken == null) return false;

      final response = await _dio.post(
        ApiConstants.authRefresh,
        data: {'refresh_token': refreshToken},
        options: Options(headers: {}), // Don't send old token
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        final authResponse = AuthResponse.fromJson(response.data['data']);
        await _storageService.saveAuthData(
          accessToken: authResponse.accessToken,
          refreshToken: authResponse.refreshToken,
          user: authResponse.user,
        );
        return true;
      }
      return false;
    } catch (e) {
      debugPrint('Token refresh failed: $e');
      return false;
    }
  }

  Future<AuthResponse> register(RegisterRequest request) async {
    try {
      final response = await _dio.post(
        ApiConstants.authRegister,
        data: request.toJson(),
      );

      if (response.statusCode == 201 && response.data['success'] == true) {
        // Registration returns user, need to login to get tokens
        final user = User.fromJson(response.data['data']);
        // Auto-login after registration
        return await login(LoginRequest(
          email: request.email,
          password: request.password,
        ));
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<AuthResponse> login(LoginRequest request) async {
    try {
      final response = await _dio.post(
        ApiConstants.authLogin,
        data: request.toJson(),
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        final authResponse = AuthResponse.fromJson(response.data['data']);
        await _storageService.saveAuthData(
          accessToken: authResponse.accessToken,
          refreshToken: authResponse.refreshToken,
          user: authResponse.user,
        );
        return authResponse;
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<User> getCurrentUser() async {
    try {
      final response = await _dio.get(ApiConstants.authMe);

      if (response.statusCode == 200 && response.data['success'] == true) {
        final user = User.fromJson(response.data['data']);
        await _storageService.saveUser(user);
        return user;
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<void> logout() async {
    try {
      final refreshToken = await _storageService.getRefreshToken();
      if (refreshToken != null) {
        await _dio.post(
          ApiConstants.authLogout,
          data: {'refresh_token': refreshToken},
        );
      }
    } catch (e) {
      debugPrint('Logout API call failed: $e');
    } finally {
      await _storageService.clearAuthData();
    }
  }

  Future<AuthResponse> refreshToken() async {
    try {
      final refreshToken = await _storageService.getRefreshToken();
      if (refreshToken == null) {
        throw Exception('No refresh token available');
      }

      final response = await _dio.post(
        ApiConstants.authRefresh,
        data: {'refresh_token': refreshToken},
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        final authResponse = AuthResponse.fromJson(response.data['data']);
        await _storageService.saveAuthData(
          accessToken: authResponse.accessToken,
          refreshToken: authResponse.refreshToken,
          user: authResponse.user,
        );
        return authResponse;
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Exception _handleError(Response response) {
    final error = response.data['error'];
    if (error != null) {
      return Exception(error['message'] ?? 'An error occurred');
    }
    return Exception('An error occurred');
  }

  Exception _handleDioError(DioException e) {
    if (e.response != null) {
      final data = e.response?.data;
      if (data is Map && data['error'] != null) {
        return Exception(data['error']['message'] ?? 'An error occurred');
      }
      switch (e.response?.statusCode) {
        case 400:
          return Exception('Invalid request');
        case 401:
          return Exception('Invalid credentials');
        case 409:
          return Exception('Email or username already exists');
        case 500:
          return Exception('Server error');
        default:
          return Exception('An error occurred');
      }
    }
    if (e.type == DioExceptionType.connectionTimeout ||
        e.type == DioExceptionType.receiveTimeout) {
      return Exception('Connection timeout');
    }
    return Exception('Network error: ${e.message}');
  }
}
