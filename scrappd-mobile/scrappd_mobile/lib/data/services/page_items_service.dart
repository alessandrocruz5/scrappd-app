import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../../core/constants/api_constants.dart';
import '../models/page_item.dart';
import 'secure_storage_service.dart';

class PageItemsService {
  late final Dio _dio;
  final SecureStorageService _storageService;

  PageItemsService(this._storageService) {
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

  /// List all items on a page
  Future<List<PageItem>> listPageItems(String pageId) async {
    try {
      final response = await _dio.get('${ApiConstants.pages}/$pageId/items');

      if (response.statusCode == 200 && response.data['success'] == true) {
        final data = response.data['data'] as List;
        return data.map((json) => PageItem.fromJson(json)).toList();
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Add an item to a page
  Future<PageItem> addItemToPage(
    String pageId,
    CreatePageItemRequest request,
  ) async {
    try {
      final response = await _dio.post(
        '${ApiConstants.pages}/$pageId/items',
        data: request.toJson(),
      );

      if (response.statusCode == 201 && response.data['success'] == true) {
        return PageItem.fromJson(response.data['data']);
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Update a page item (position, size, rotation, etc.)
  Future<PageItem> updatePageItem(
    String pageId,
    String itemId,
    UpdatePageItemRequest request,
  ) async {
    try {
      final response = await _dio.patch(
        '${ApiConstants.pages}/$pageId/items/$itemId',
        data: request.toJson(),
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        return PageItem.fromJson(response.data['data']);
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Remove an item from a page
  Future<void> removeItemFromPage(String pageId, String itemId) async {
    try {
      final response =
          await _dio.delete('${ApiConstants.pages}/$pageId/items/$itemId');

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
          return Exception('Item not found');
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
