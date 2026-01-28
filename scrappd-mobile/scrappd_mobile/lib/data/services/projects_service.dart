import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import '../../core/constants/api_constants.dart';
import '../models/project.dart';
import 'secure_storage_service.dart';

class ProjectsService {
  late final Dio _dio;
  final SecureStorageService _storageService;

  ProjectsService(this._storageService) {
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

  Future<ProjectsResponse> listProjects({int page = 1, int perPage = 20}) async {
    try {
      final response = await _dio.get(
        ApiConstants.projects,
        queryParameters: {
          'page': page,
          'per_page': perPage,
        },
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        final data = response.data['data'] as List;
        final projects = data.map((json) => Project.fromJson(json)).toList();
        final meta = response.data['meta'];
        return ProjectsResponse(
          projects: projects,
          total: meta?['total'] ?? projects.length,
          page: meta?['page'] ?? page,
          perPage: meta?['per_page'] ?? perPage,
        );
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<Project> createProject(CreateProjectRequest request) async {
    try {
      final response = await _dio.post(
        ApiConstants.projects,
        data: request.toJson(),
      );

      if (response.statusCode == 201 && response.data['success'] == true) {
        return Project.fromJson(response.data['data']);
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<Project> getProject(String id) async {
    try {
      final response = await _dio.get('${ApiConstants.projects}/$id');

      if (response.statusCode == 200 && response.data['success'] == true) {
        return Project.fromJson(response.data['data']);
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<Project> updateProject(String id, UpdateProjectRequest request) async {
    try {
      final response = await _dio.patch(
        '${ApiConstants.projects}/$id',
        data: request.toJson(),
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        return Project.fromJson(response.data['data']);
      }

      throw _handleError(response);
    } on DioException catch (e) {
      throw _handleDioError(e);
    }
  }

  Future<void> deleteProject(String id) async {
    try {
      final response = await _dio.delete('${ApiConstants.projects}/$id');

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
          return Exception('Project not found');
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

class ProjectsResponse {
  final List<Project> projects;
  final int total;
  final int page;
  final int perPage;

  ProjectsResponse({
    required this.projects,
    required this.total,
    required this.page,
    required this.perPage,
  });

  bool get hasMore => projects.length == perPage && page * perPage < total;
}
