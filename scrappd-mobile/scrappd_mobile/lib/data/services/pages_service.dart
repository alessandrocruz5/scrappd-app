import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../../core/constants/api_constants.dart';
import '../models/page.dart';
import 'secure_storage_service.dart';

class PagesService {
  late final Dio _dio;
  final SecureStorageService _storageService;

  PagesService(this._storageService) {
    _dio = Dio(BaseOptions(
      baseUrl: ApiConstants.baseUrl,
      connectTimeout: ApiConstants.connectionTimeout,
      receiveTimeout: ApiConstants.receiveTimeout,
      sendTimeout: ApiConstants.sendTimeout,
      headers: ApiConstants.defaultHeaders,
    ));

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
    ));
  }

  Future<List<ScrapbookPage>> listPages(String projectId) async {
    try {
      final response = await _dio.get(
        ApiConstants.pages,
        queryParameters: {'project_id': projectId},
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        final data = response.data['data'] as List;
        return data.map((json) => ScrapbookPage.fromJson(json)).toList();
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<ScrapbookPage> createPage(CreatePageRequest request) async {
    try {
      final response = await _dio.post(
        ApiConstants.pages,
        data: request.toJson(),
      );

      if (response.statusCode == 201 && response.data['success'] == true) {
        return ScrapbookPage.fromJson(response.data['data']);
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<ScrapbookPage> getPage(String id) async {
    try {
      final response = await _dio.get('${ApiConstants.pages}/$id');

      if (response.statusCode == 200 && response.data['success'] == true) {
        return ScrapbookPage.fromJson(response.data['data']);
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<ScrapbookPage> updatePage(String id, UpdatePageRequest request) async {
    try {
      final response = await _dio.patch(
        '${ApiConstants.pages}/$id',
        data: request.toJson(),
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        return ScrapbookPage.fromJson(response.data['data']);
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<void> deletePage(String id) async {
    try {
      final response = await _dio.delete('${ApiConstants.pages}/$id');

      if (response.statusCode == 204 || response.data['success'] == true) {
        return;
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
          return Exception('Unauthorized');
        case 404:
          return Exception('Page not found');
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
