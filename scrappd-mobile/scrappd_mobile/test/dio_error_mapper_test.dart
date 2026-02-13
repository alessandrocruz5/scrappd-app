import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:scrappd_mobile/core/network/dio_error_mapper.dart';

void main() {
  test('maps connection timeout to retryable message', () {
    final error = DioException(
      requestOptions: RequestOptions(path: '/test'),
      type: DioExceptionType.connectionTimeout,
    );

    final result = DioErrorMapper.map(error);

    expect(result.isRetryable, isTrue);
    expect(result.message, 'Connection timed out. Please try again.');
  });

  test('maps receive timeout to retryable message', () {
    final error = DioException(
      requestOptions: RequestOptions(path: '/test'),
      type: DioExceptionType.receiveTimeout,
    );

    final result = DioErrorMapper.map(error);

    expect(result.isRetryable, isTrue);
    expect(result.message,
        'The server is taking too long to respond. Please try again.');
  });

  test('maps connection error to retryable message', () {
    final error = DioException(
      requestOptions: RequestOptions(path: '/test'),
      type: DioExceptionType.connectionError,
    );

    final result = DioErrorMapper.map(error);

    expect(result.isRetryable, isTrue);
    expect(result.message,
        'Unable to connect. Check your internet connection and try again.');
  });

  test('maps 401 bad response to session expired message', () {
    final requestOptions = RequestOptions(path: '/test');
    final response = Response(
      requestOptions: requestOptions,
      statusCode: 401,
      data: {},
    );
    final error = DioException(
      requestOptions: requestOptions,
      response: response,
      type: DioExceptionType.badResponse,
    );

    final result = DioErrorMapper.map(error);

    expect(result.message,
        'Your session has expired. Please sign in again.');
  });

  test('uses server message for validation error', () {
    final requestOptions = RequestOptions(path: '/test');
    final response = Response(
      requestOptions: requestOptions,
      statusCode: 422,
      data: {
        'error': {'message': 'Email is already taken'},
      },
    );
    final error = DioException(
      requestOptions: requestOptions,
      response: response,
      type: DioExceptionType.badResponse,
    );

    final result = DioErrorMapper.map(error);

    expect(result.message, 'Email is already taken');
  });

  test('extracts message from byte payloads', () {
    final requestOptions = RequestOptions(path: '/test');
    final response = Response(
      requestOptions: requestOptions,
      statusCode: 400,
      data: utf8.encode('{"message":"Bad input"}'),
    );
    final error = DioException(
      requestOptions: requestOptions,
      response: response,
      type: DioExceptionType.badResponse,
    );

    final result = DioErrorMapper.map(error);

    expect(result.message, 'Bad input');
  });
}
