import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:scrappd_mobile/core/config/environment.dart';
import '../../core/constants/api_constants.dart';
import '../models/processed_image.dart';

class ApiService {
  late final Dio _dio;

  ApiService() {
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

    _dio.interceptors.add(_createLogInterceptor());
    _dio.interceptors.add(_createRetryInterceptor());
  }

    Interceptor _createLogInterceptor() {
    return InterceptorsWrapper(
      onRequest: (options, handler) {
        if (EnvironmentConfig.verboseLogging) {
          debugPrint('📤 ${options.method} ${options.uri}');
        }
        handler.next(options);
      },
      onResponse: (response, handler) {
        if (EnvironmentConfig.verboseLogging) {
          debugPrint('📥 ${response.statusCode} ${response.requestOptions.uri}');
        }
        handler.next(response);
      },
      onError: (error, handler) {
        debugPrint('❌ ${error.type}: ${error.message}');
        handler.next(error);
      },
    );
  }

    /// Retry interceptor for handling cold starts
  Interceptor _createRetryInterceptor() {
    return InterceptorsWrapper(
      onError: (error, handler) async {
        // Only retry on connection/timeout errors in production
        if (EnvironmentConfig.current == Environment.production) {
          if (_shouldRetry(error) && _getRetryCount(error) < 2) {
            debugPrint('🔄 Retrying request (cold start recovery)...');
            
            // Wait a bit before retrying
            await Future.delayed(const Duration(seconds: 2));
            
            try {
              final options = error.requestOptions;
              options.extra['retryCount'] = _getRetryCount(error) + 1;
              
              final response = await _dio.fetch(options);
              return handler.resolve(response);
            } catch (e) {
              return handler.next(error);
            }
          }
        }
        handler.next(error);
      },
    );
  }
  
  bool _shouldRetry(DioException error) {
    return error.type == DioExceptionType.connectionTimeout ||
           error.type == DioExceptionType.receiveTimeout ||
           error.type == DioExceptionType.connectionError;
  }
  
  int _getRetryCount(DioException error) {
    return error.requestOptions.extra['retryCount'] ?? 0;
  }

  /// Health check
  Future<bool> healthCheck() async {
    try {
      final response = await _dio.get(ApiConstants.healthCheck);
      return response.statusCode == 200;
    } catch (e) {
      debugPrint('Health check failed: $e');
      return false;
    }
  }

    /// Deep health check - verifies all services
  Future<Map<String, dynamic>?> deepHealthCheck() async {
    try {
      final response = await _dio.get(ApiConstants.healthDeep);
      if (response.statusCode == 200) {
        return response.data;
      }
    } catch (e) {
      debugPrint('Deep health check failed: $e');
    }
    return null;
  }

  /// Get content type from filename extension
  String _getContentTypeFromFilename(String filename) {
    final ext = filename.toLowerCase().split('.').last;
    switch (ext) {
      case 'jpg':
      case 'jpeg':
        return 'image/jpeg';
      case 'png':
        return 'image/png';
      case 'webp':
        return 'image/webp';
      default:
        return 'image/jpeg'; // Default to JPEG
    }
  }

  /// Remove background from image file
  Future<Uint8List> removeBackground(File imageFile, {
    void Function(int sent, int total)? onProgress,
  }) async {
    try {
      final filename = imageFile.path.split('/').last;
      final contentType = _getContentTypeFromFilename(filename);
      final start = DateTime.now();
      final fileSize = await imageFile.length();

      final formData = FormData.fromMap({
        'image': await MultipartFile.fromFile(
          imageFile.path,
          filename: filename,
          contentType: DioMediaType.parse(contentType),
        ),
        'format': 'png',
      });
      
      final response = await _dio.post(
        ApiConstants.removeBackground,
        data: formData,
        options: Options(
          responseType: ResponseType.bytes,
          // Longer timeout for ML processing
          sendTimeout: const Duration(seconds: 180),
          receiveTimeout: const Duration(seconds: 180),
          // Let Dio set Content-Type automatically for FormData (with boundary)
          contentType: 'multipart/form-data',
        ),
        onSendProgress: onProgress,
      );

      if (response.statusCode == 200 && response.data != null) {
        if (EnvironmentConfig.verboseLogging) {
          final elapsed = DateTime.now().difference(start);
          final mb = (fileSize / (1024 * 1024)).toStringAsFixed(2);
          debugPrint('🧾 Upload+process completed in ${elapsed.inSeconds}s '
              '(${elapsed.inMilliseconds}ms). File size: ${mb}MB');
        }
        return Uint8List.fromList(response.data);
      }
      
      throw ApiException('Failed to process image', response.statusCode);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

    /// Process image from bytes
  Future<Uint8List> removeBackgroundFromBytes(Uint8List imageBytes, {
    String filename = 'image.jpg',
    void Function(int sent, int total)? onProgress,
  }) async {
    try {
      final contentType = _getContentTypeFromFilename(filename);
      final start = DateTime.now();
      final byteCount = imageBytes.length;

      final formData = FormData.fromMap({
        'image': MultipartFile.fromBytes(
          imageBytes,
          filename: filename,
          contentType: DioMediaType.parse(contentType),
        ),
        'format': 'png',
      });
      
      final response = await _dio.post(
        ApiConstants.removeBackground,
        data: formData,
        options: Options(
          responseType: ResponseType.bytes,
          sendTimeout: const Duration(seconds: 180),
          receiveTimeout: const Duration(seconds: 180),
          // Let Dio set Content-Type automatically for FormData (with boundary)
          contentType: 'multipart/form-data',
        ),
        onSendProgress: onProgress,
      );

      if (response.statusCode == 200 && response.data != null) {
        if (EnvironmentConfig.verboseLogging) {
          final elapsed = DateTime.now().difference(start);
          final mb = (byteCount / (1024 * 1024)).toStringAsFixed(2);
          debugPrint('🧾 Upload+process completed in ${elapsed.inSeconds}s '
              '(${elapsed.inMilliseconds}ms). File size: ${mb}MB');
        }
        return Uint8List.fromList(response.data);
      }
      
      throw ApiException('Failed to process image', response.statusCode);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

    ApiException _handleDioError(DioException error) {
    switch (error.type) {
      case DioExceptionType.connectionTimeout:
        return ApiException(
          'Connection timed out. The server might be starting up - please try again.',
          null,
          isRetryable: true,
        );
      case DioExceptionType.receiveTimeout:
        return ApiException(
          'Request timed out. Image processing is taking longer than expected.',
          null,
          isRetryable: true,
        );
      case DioExceptionType.connectionError:
        return ApiException(
          'Unable to connect to server. Please check your internet connection.',
          null,
          isRetryable: true,
        );
      case DioExceptionType.badResponse:
        final statusCode = error.response?.statusCode;
        final message = _parseErrorMessage(error.response?.data);
        return ApiException(message ?? 'Server error', statusCode);
      default:
        return ApiException(
          error.message ?? 'An unexpected error occurred',
          null,
        );
    }
  }

  String? _parseErrorMessage(dynamic data) {
    if (data == null) return null;
    if (data is Map) {
      return data['error']?['message'] ?? data['message'] ?? data['detail'];
    }
    if (data is String) return data;
    return null;
  }
  
  void dispose() {
    _dio.close();
  }


  /// Get processing metadata (if you add this endpoint later)
  Future<RemovalMetadata?> getMetadata(String imageId) async {
    try {
      final response = await _dio.get('/api/v1/images/$imageId/metadata');
      
      if (response.statusCode == 200) {
        final apiResponse = ApiResponse.fromJson(
          response.data,
          (data) => RemovalMetadata.fromJson(data as Map<String, dynamic>),
        );
        return apiResponse.data;
      }
      return null;
    } catch (e) {
      debugPrint('Error fetching metadata: $e');
      return null;
    }
  }

  /// Cancel all pending requests
  void cancelRequests() {
    _dio.close(force: true);
  }
}

/// Custom API Exception
class ApiException implements Exception {
  final String message;
  final int? statusCode;
  final bool isRetryable;
  
  ApiException(this.message, this.statusCode, {this.isRetryable = false});
  
  @override
  String toString() => 'ApiException: $message (status: $statusCode)';
}
