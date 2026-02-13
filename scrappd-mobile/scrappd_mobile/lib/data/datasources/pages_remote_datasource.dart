import 'package:dio/dio.dart';

import '../../core/constants/api_constants.dart';
import '../../core/models/api_response.dart';
import '../../core/network/error_helpers.dart';
import '../../core/network/friendly_exception.dart';
import '../models/page_model.dart';

class PagesRemoteDataSource {
  PagesRemoteDataSource(this._dio);

  final Dio _dio;

  Future<PageModel> createPage({
    required String projectId,
    required int pageNumber,
    String? title,
    String? backgroundColor,
  }) async {
    return runWithFriendlyErrors(
      () async {
        final response = await _dio.post(
          ApiConstants.pages,
          data: {
            'project_id': projectId,
            'page_number': pageNumber,
            if (title != null && title.isNotEmpty) 'title': title,
            if (backgroundColor != null && backgroundColor.isNotEmpty)
              'background_color': backgroundColor,
          },
        );

        final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
          response.data as Map<String, dynamic>,
          (data) => data as Map<String, dynamic>,
        );

        if (!apiResponse.success || apiResponse.data == null) {
          throw FriendlyException(
            apiResponse.error?.message ?? 'Failed to create page.',
          );
        }

        return PageModel.fromJson(apiResponse.data!);
      },
      defaultMessage: 'Failed to create page.',
    );
  }

  Future<(List<PageModel>, ApiMeta?)> listPages({
    required String projectId,
    int page = 1,
    int perPage = 50,
  }) async {
    return runWithFriendlyErrors(
      () async {
        final response = await _dio.get(
          ApiConstants.pages,
          queryParameters: {
            'project_id': projectId,
            'page': page,
            'per_page': perPage,
          },
        );

        final apiResponse = ApiResponse<List<dynamic>>.fromJson(
          response.data as Map<String, dynamic>,
          (data) => data as List<dynamic>,
        );

        if (!apiResponse.success || apiResponse.data == null) {
          throw FriendlyException(
            apiResponse.error?.message ?? 'Failed to load pages.',
          );
        }

        final pages = apiResponse.data!
            .map((page) => PageModel.fromJson(page as Map<String, dynamic>))
            .toList();

        return (pages, apiResponse.meta);
      },
      defaultMessage: 'Failed to load pages.',
    );
  }
}
