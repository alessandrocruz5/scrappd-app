import 'package:dio/dio.dart';

import '../../core/constants/api_constants.dart';
import '../../core/models/api_response.dart';
import '../../core/network/error_helpers.dart';
import '../../core/network/friendly_exception.dart';
import '../models/page_item_model.dart';

class PageItemsRemoteDataSource {
  PageItemsRemoteDataSource(this._dio);

  final Dio _dio;

  Future<List<PageItemModel>> listPageItems({
    required String pageId,
  }) async {
    return runWithFriendlyErrors(
      () async {
        final response = await _dio.get(
          '${ApiConstants.pages}/$pageId/items',
        );

        final apiResponse = ApiResponse<List<dynamic>>.fromJson(
          response.data as Map<String, dynamic>,
          (data) => data as List<dynamic>,
        );

        if (!apiResponse.success || apiResponse.data == null) {
          throw FriendlyException(
            apiResponse.error?.message ?? 'Failed to load page items.',
          );
        }

        return apiResponse.data!
            .map((item) => PageItemModel.fromJson(item as Map<String, dynamic>))
            .toList();
      },
      defaultMessage: 'Failed to load page items.',
    );
  }

  Future<PageItemModel> createPageItem({
    required String pageId,
    required String itemId,
    required double positionX,
    required double positionY,
    required double width,
    required double height,
    required double rotation,
    int? zIndex,
    double? opacity,
  }) async {
    return runWithFriendlyErrors(
      () async {
        final response = await _dio.post(
          '${ApiConstants.pages}/$pageId/items',
          data: {
            'item_id': itemId,
            'position_x': positionX,
            'position_y': positionY,
            'width': width,
            'height': height,
            'rotation': rotation,
            if (zIndex != null) 'z_index': zIndex,
            if (opacity != null) 'opacity': opacity,
          },
        );

        final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
          response.data as Map<String, dynamic>,
          (data) => data as Map<String, dynamic>,
        );

        if (!apiResponse.success || apiResponse.data == null) {
          throw FriendlyException(
            apiResponse.error?.message ?? 'Failed to add page item.',
          );
        }

        return PageItemModel.fromJson(apiResponse.data!);
      },
      defaultMessage: 'Failed to add page item.',
    );
  }

  Future<PageItemModel> updatePageItem({
    required String pageId,
    required String pageItemId,
    required double positionX,
    required double positionY,
    required double width,
    required double height,
    required double rotation,
    int? zIndex,
    double? opacity,
  }) async {
    return runWithFriendlyErrors(
      () async {
        final response = await _dio.patch(
          '${ApiConstants.pages}/$pageId/items/$pageItemId',
          data: {
            'position_x': positionX,
            'position_y': positionY,
            'width': width,
            'height': height,
            'rotation': rotation,
            if (zIndex != null) 'z_index': zIndex,
            if (opacity != null) 'opacity': opacity,
          },
        );

        final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
          response.data as Map<String, dynamic>,
          (data) => data as Map<String, dynamic>,
        );

        if (!apiResponse.success || apiResponse.data == null) {
          throw FriendlyException(
            apiResponse.error?.message ?? 'Failed to update page item.',
          );
        }

        return PageItemModel.fromJson(apiResponse.data!);
      },
      defaultMessage: 'Failed to update page item.',
    );
  }

  Future<void> deletePageItem({
    required String pageId,
    required String pageItemId,
  }) async {
    return runWithFriendlyErrors(
      () async {
        final response = await _dio.delete(
          '${ApiConstants.pages}/$pageId/items/$pageItemId',
        );

        if (response.statusCode != 204) {
          throw FriendlyException('Failed to delete page item.');
        }
      },
      defaultMessage: 'Failed to delete page item.',
    );
  }
}
