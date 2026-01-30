import 'dart:async';

import 'package:dio/dio.dart';

import '../constants/api_constants.dart';
import '../models/api_response.dart';
import '../storage/token_storage.dart';

class AuthInterceptor extends Interceptor {
  AuthInterceptor({
    required TokenStorage tokenStorage,
    required Dio refreshDio,
  })  : _tokenStorage = tokenStorage,
        _refreshDio = refreshDio;

  final TokenStorage _tokenStorage;
  final Dio _refreshDio;
  Completer<String?>? _refreshCompleter;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final accessToken = _tokenStorage.accessToken;
    if (accessToken != null && accessToken.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $accessToken';
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    final statusCode = err.response?.statusCode ?? 0;
    final hasRefreshToken = _tokenStorage.refreshToken?.isNotEmpty ?? false;
    final isRetry = err.requestOptions.extra['retry'] == true;
    final isAuthRequest = err.requestOptions.path.contains('/auth/');

    if (statusCode == 401 && hasRefreshToken && !isRetry && !isAuthRequest) {
      try {
        final newAccessToken = await _refreshAccessToken();
        if (newAccessToken != null) {
          final requestOptions = err.requestOptions;
          requestOptions.extra['retry'] = true;
          requestOptions.headers['Authorization'] = 'Bearer $newAccessToken';

          final response = await _refreshDio.fetch(requestOptions);
          return handler.resolve(response);
        }
      } catch (_) {
        // Fall through to original error.
      }
    }

    handler.next(err);
  }

  Future<String?> _refreshAccessToken() async {
    if (_refreshCompleter != null) {
      return _refreshCompleter!.future;
    }

    _refreshCompleter = Completer<String?>();
    try {
      final refreshToken = _tokenStorage.refreshToken;
      if (refreshToken == null || refreshToken.isEmpty) {
        _refreshCompleter!.complete(null);
        _refreshCompleter = null;
        return null;
      }

      final response = await _refreshDio.post(
        ApiConstants.authRefresh,
        data: {'refresh_token': refreshToken},
      );

      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
        response.data as Map<String, dynamic>,
        (data) => data as Map<String, dynamic>,
      );

      final payload = apiResponse.data ?? {};
      final newAccessToken = payload['access_token'] as String?;
      final newRefreshToken = payload['refresh_token'] as String?;

      if (newAccessToken != null && newRefreshToken != null) {
        await _tokenStorage.saveTokens(
          accessToken: newAccessToken,
          refreshToken: newRefreshToken,
        );
        _refreshCompleter!.complete(newAccessToken);
        _refreshCompleter = null;
        return newAccessToken;
      }
    } catch (_) {
      _refreshCompleter?.complete(null);
      _refreshCompleter = null;
    }

    return null;
  }
}
