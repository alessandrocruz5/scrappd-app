import 'package:dio/dio.dart';

import 'dio_error_mapper.dart';
import 'friendly_exception.dart';

Future<T> runWithFriendlyErrors<T>(
  Future<T> Function() action, {
  String? defaultMessage,
}) async {
  try {
    return await action();
  } on DioException catch (e) {
    final result = DioErrorMapper.map(
      e,
      defaultMessage: defaultMessage,
    );
    throw FriendlyException(
      result.message,
      statusCode: result.statusCode,
      isRetryable: result.isRetryable,
    );
  }
}

String mapErrorMessage(
  Object error, {
  String? fallback,
}) {
  if (error is FriendlyException) return error.message;
  if (error is DioException) {
    return DioErrorMapper.map(
      error,
      defaultMessage: fallback,
    ).message;
  }
  final raw = error.toString();
  if (raw.startsWith('Exception: ')) {
    return raw.substring('Exception: '.length);
  }
  return fallback ?? raw;
}
