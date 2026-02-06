import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../../core/constants/api_constants.dart';

class PageExportService {
  PageExportService(this._dio);

  final Dio _dio;

  Future<Uint8List> exportPage({
    required String pageId,
    double scale = 2.0,
    String format = 'png',
    int quality = 92,
    int? width,
    int? height,
  }) async {
    try {
      final response = await _dio.get(
        ApiConstants.pageRender(pageId),
        queryParameters: {
          'scale': scale,
          'format': format,
          'quality': quality,
          if (width != null && width > 0) 'width': width,
          if (height != null && height > 0) 'height': height,
        },
        options: Options(
          responseType: ResponseType.bytes,
          receiveTimeout: const Duration(seconds: 120),
        ),
      );

      if (response.statusCode == 200 && response.data != null) {
        return Uint8List.fromList(response.data);
      }

      throw const PageExportException('Failed to export page');
    } on DioException catch (e) {
      throw PageExportException(_extractErrorMessage(e));
    } catch (e) {
      throw PageExportException(e.toString());
    }
  }

  String _extractErrorMessage(DioException error) {
    final data = error.response?.data;
    if (data is Map) {
      return data['error']?['message'] ?? data['message'] ?? 'Export failed';
    }
    if (data is List<int>) {
      try {
        final decoded = utf8.decode(data);
        final jsonData = jsonDecode(decoded);
        if (jsonData is Map) {
          return jsonData['error']?['message'] ??
              jsonData['message'] ??
              'Export failed';
        }
        return decoded;
      } catch (_) {
        return 'Export failed';
      }
    }
    if (data is String) {
      return data;
    }
    return error.message ?? 'Export failed';
  }
}

class PageExportException implements Exception {
  const PageExportException(this.message);

  final String message;

  @override
  String toString() => message;
}
