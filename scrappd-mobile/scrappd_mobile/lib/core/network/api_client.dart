import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../constants/api_constants.dart';
import '../storage/token_storage.dart';
import 'auth_interceptor.dart';

class ApiClient {
  ApiClient(this._tokenStorage) {
    final options = BaseOptions(
      baseUrl: ApiConstants.baseUrl,
      connectTimeout: ApiConstants.connectionTimeout,
      receiveTimeout: ApiConstants.receiveTimeout,
      sendTimeout: ApiConstants.sendTimeout,
      headers: ApiConstants.defaultHeaders,
    );

    _dio = Dio(options);
    _refreshDio = Dio(options);

    _dio.interceptors.add(
      AuthInterceptor(
        tokenStorage: _tokenStorage,
        refreshDio: _refreshDio,
      ),
    );

    if (kDebugMode) {
      _dio.interceptors.add(LogInterceptor(
        requestBody: true,
        responseBody: true,
        error: true,
      ));
    }
  }

  final TokenStorage _tokenStorage;
  late final Dio _dio;
  late final Dio _refreshDio;

  Dio get dio => _dio;
}
