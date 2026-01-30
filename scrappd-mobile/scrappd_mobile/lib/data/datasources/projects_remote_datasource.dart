import 'package:dio/dio.dart';

import '../../core/constants/api_constants.dart';
import '../../core/models/api_response.dart';
import '../models/project_model.dart';

class ProjectsRemoteDataSource {
  ProjectsRemoteDataSource(this._dio);

  final Dio _dio;

  Future<ProjectModel> createProject({
    required String title,
    String? description,
  }) async {
    final response = await _dio.post(
      ApiConstants.projects,
      data: {
        'title': title,
        if (description != null && description.isNotEmpty)
          'description': description,
      },
    );

    final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
      response.data as Map<String, dynamic>,
      (data) => data as Map<String, dynamic>,
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Failed to create project');
    }

    return ProjectModel.fromJson(apiResponse.data!);
  }

  Future<(List<ProjectModel>, ApiMeta?)> listProjects({
    int page = 1,
    int perPage = 50,
  }) async {
    final response = await _dio.get(
      ApiConstants.projects,
      queryParameters: {
        'page': page,
        'per_page': perPage,
      },
    );

    final apiResponse = ApiResponse<List<dynamic>>.fromJson(
      response.data as Map<String, dynamic>,
      (data) => data as List<dynamic>,
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Failed to load projects');
    }

    final projects = apiResponse.data!
        .map((project) => ProjectModel.fromJson(project as Map<String, dynamic>))
        .toList();

    return (projects, apiResponse.meta);
  }
}
