import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../../core/constants/api_constants.dart';
import '../models/page_item.dart';
import 'secure_storage_service.dart';

class ItemsService {
  late final Dio _dio;
  final SecureStorageService _storageService;

  ItemsService(this._storageService) {
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

  /// List user's items
  Future<ItemsResponse> listItems({int page = 1, int perPage = 20}) async {
    try {
      final response = await _dio.get(
        ApiConstants.items,
        queryParameters: {
          'page': page,
          'per_page': perPage,
        },
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        final data = response.data['data'] as List;
        final items = data.map((json) => Item.fromJson(json)).toList();
        final meta = response.data['meta'];
        return ItemsResponse(
          items: items,
          total: meta?['total'] ?? items.length,
          page: meta?['page'] ?? page,
          perPage: meta?['per_page'] ?? perPage,
        );
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Upload a new item (image)
  Future<Item> uploadItem(File imageFile, {String? name}) async {
    try {
      final formData = FormData.fromMap({
        'image': await MultipartFile.fromFile(
          imageFile.path,
          filename: imageFile.path.split('/').last,
        ),
        if (name != null) 'name': name,
      });

      final response = await _dio.post(
        ApiConstants.items,
        data: formData,
        options: Options(
          headers: {'Content-Type': 'multipart/form-data'},
        ),
      );

      if (response.statusCode == 201 && response.data['success'] == true) {
        return Item.fromJson(response.data['data']);
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Get a specific item
  Future<Item> getItem(String id) async {
    try {
      final response = await _dio.get('${ApiConstants.items}/$id');

      if (response.statusCode == 200 && response.data['success'] == true) {
        return Item.fromJson(response.data['data']);
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Delete an item
  Future<void> deleteItem(String id) async {
    try {
      final response = await _dio.delete('${ApiConstants.items}/$id');

      if (response.statusCode == 204 || response.data['success'] == true) {
        return;
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  /// Get usage statistics
  Future<UsageStats> getUsage() async {
    try {
      final response = await _dio.get('${ApiConstants.items}/usage');

      if (response.statusCode == 200 && response.data['success'] == true) {
        return UsageStats.fromJson(response.data['data']);
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
        case 413:
          return Exception('File too large');
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

class ItemsResponse {
  final List<Item> items;
  final int total;
  final int page;
  final int perPage;

  ItemsResponse({
    required this.items,
    required this.total,
    required this.page,
    required this.perPage,
  });

  bool get hasMore => items.length == perPage && page * perPage < total;
}

class UsageStats {
  final int itemCount;
  final int storageBytesUsed;
  final int storageBytesLimit;
  final int bgRemovalsUsed;
  final int bgRemovalsLimit;

  UsageStats({
    required this.itemCount,
    required this.storageBytesUsed,
    required this.storageBytesLimit,
    required this.bgRemovalsUsed,
    required this.bgRemovalsLimit,
  });

  factory UsageStats.fromJson(Map<String, dynamic> json) {
    return UsageStats(
      itemCount: json['item_count'] ?? 0,
      storageBytesUsed: json['storage_bytes_used'] ?? 0,
      storageBytesLimit: json['storage_bytes_limit'] ?? 0,
      bgRemovalsUsed: json['bg_removals_used'] ?? 0,
      bgRemovalsLimit: json['bg_removals_limit'] ?? 0,
    );
  }

  double get storagePercentUsed =>
      storageBytesLimit > 0 ? storageBytesUsed / storageBytesLimit : 0.0;

  double get bgRemovalsPercentUsed =>
      bgRemovalsLimit > 0 ? bgRemovalsUsed / bgRemovalsLimit : 0.0;
}
