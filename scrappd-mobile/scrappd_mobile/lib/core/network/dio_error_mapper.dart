import 'dart:convert';

import 'package:dio/dio.dart';

class DioErrorResult {
  const DioErrorResult({
    required this.message,
    this.statusCode,
    this.isRetryable = false,
  });

  final String message;
  final int? statusCode;
  final bool isRetryable;
}

class DioErrorMapper {
  static DioErrorResult map(
    DioException error, {
    String? defaultMessage,
  }) {
    final statusCode = error.response?.statusCode;
    final serverMessage = _extractServerMessage(error.response?.data);

    switch (error.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
        return DioErrorResult(
          message: 'Connection timed out. Please try again.',
          statusCode: statusCode,
          isRetryable: true,
        );
      case DioExceptionType.receiveTimeout:
        return DioErrorResult(
          message: 'The server is taking too long to respond. Please try again.',
          statusCode: statusCode,
          isRetryable: true,
        );
      case DioExceptionType.connectionError:
        return DioErrorResult(
          message: 'Unable to connect. Check your internet connection and try again.',
          statusCode: statusCode,
          isRetryable: true,
        );
      case DioExceptionType.badCertificate:
        return DioErrorResult(
          message: 'Secure connection failed. Please try again.',
          statusCode: statusCode,
        );
      case DioExceptionType.cancel:
        return DioErrorResult(
          message: 'Request was canceled.',
          statusCode: statusCode,
        );
      case DioExceptionType.badResponse:
        return _mapBadResponse(statusCode, serverMessage, defaultMessage);
      case DioExceptionType.unknown:
        return DioErrorResult(
          message: serverMessage ??
              error.message ??
              defaultMessage ??
              'Something went wrong. Please try again.',
          statusCode: statusCode,
        );
    }
  }

  static DioErrorResult _mapBadResponse(
    int? statusCode,
    String? serverMessage,
    String? defaultMessage,
  ) {
    if (statusCode == null) {
      return DioErrorResult(
        message: serverMessage ??
            defaultMessage ??
            'Server error. Please try again.',
      );
    }

    if (statusCode >= 500) {
      return DioErrorResult(
        message: 'Server error. Please try again later.',
        statusCode: statusCode,
        isRetryable: true,
      );
    }

    switch (statusCode) {
      case 400:
        return DioErrorResult(
          message: serverMessage ?? 'Bad request. Please check and try again.',
          statusCode: statusCode,
        );
      case 401:
        return DioErrorResult(
          message: serverMessage ??
              'Your session has expired. Please sign in again.',
          statusCode: statusCode,
        );
      case 403:
        return DioErrorResult(
          message: serverMessage ??
              'You do not have permission to perform this action.',
          statusCode: statusCode,
        );
      case 404:
        return DioErrorResult(
          message: serverMessage ?? 'We could not find what you were looking for.',
          statusCode: statusCode,
        );
      case 409:
        return DioErrorResult(
          message: serverMessage ??
              'This action conflicts with existing data. Please review and try again.',
          statusCode: statusCode,
        );
      case 422:
        return DioErrorResult(
          message: serverMessage ??
              'Some of the information is invalid. Please review and try again.',
          statusCode: statusCode,
        );
      case 429:
        return DioErrorResult(
          message: 'Too many requests. Please wait and try again.',
          statusCode: statusCode,
          isRetryable: true,
        );
      default:
        return DioErrorResult(
          message: serverMessage ??
              defaultMessage ??
              'Server error. Please try again.',
          statusCode: statusCode,
        );
    }
  }

  static String? _extractServerMessage(dynamic data) {
    if (data == null) return null;
    if (data is Map) {
      final errorMessage = data['error']?['message'];
      if (errorMessage is String && errorMessage.trim().isNotEmpty) {
        return errorMessage;
      }
      final message = data['message'];
      if (message is String && message.trim().isNotEmpty) {
        return message;
      }
      final detail = data['detail'];
      if (detail is String && detail.trim().isNotEmpty) {
        return detail;
      }
      return null;
    }
    if (data is List<int>) {
      try {
        final decoded = utf8.decode(data);
        final jsonData = jsonDecode(decoded);
        if (jsonData is Map) {
          return _extractServerMessage(jsonData);
        }
        if (decoded.trim().isNotEmpty) {
          return decoded;
        }
      } catch (_) {
        return null;
      }
    }
    if (data is String && data.trim().isNotEmpty) return data;
    return null;
  }
}
