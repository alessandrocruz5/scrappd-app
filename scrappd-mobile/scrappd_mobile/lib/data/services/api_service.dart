import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
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

  /// Remove background from image file
  Future<File> removeBackground(File imageFile) async {
    try {
      // Create multipart form data
      final formData = FormData.fromMap({
        'image': await MultipartFile.fromFile(
          imageFile.path,
          filename: imageFile.path.split('/').last,
        ),
      });

      // Send request
      final response = await _dio.post(
        ApiConstants.removeBackground,
        data: formData,
        options: Options(
          responseType: ResponseType.bytes,
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        ),
      );

      if (response.statusCode == 200) {
        // Save processed image to temporary file
        final tempDir = Directory.systemTemp;
        final timestamp = DateTime.now().millisecondsSinceEpoch;
        final processedFile = File('${tempDir.path}/processed_$timestamp.png');
        
        await processedFile.writeAsBytes(response.data);
        
        // Get processing time from headers
        final processingTime = response.headers.value('X-Processing-Time');
        if (processingTime != null) {
          debugPrint('Processing time: ${processingTime}s');
        }
        
        return processedFile;
      } else {
        throw Exception('Failed to process image: ${response.statusCode}');
      }
    } on DioException catch (e) {
      debugPrint('Dio error: ${e.message}');
      if (e.response != null) {
        debugPrint('Response data: ${e.response?.data}');
        throw Exception('Server error: ${e.response?.statusCode}');
      } else {
        throw Exception('Network error: ${e.message}');
      }
    } catch (e) {
      debugPrint('Error removing background: $e');
      throw Exception('Failed to process image: $e');
    }
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