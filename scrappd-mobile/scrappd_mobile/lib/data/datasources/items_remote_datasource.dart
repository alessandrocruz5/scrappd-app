import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../../core/constants/api_constants.dart';
import '../../core/config/environment.dart';
import '../../core/models/api_response.dart';
import '../../core/network/error_helpers.dart';
import '../../core/network/friendly_exception.dart';
import '../models/item_model.dart';

class ItemsRemoteDataSource {
  ItemsRemoteDataSource(this._dio);

  final Dio _dio;

  Future<ItemModel> createItem({
    required File imageFile,
    String? itemName,
    String? itemCategory,
    List<String>? tags,
    String? format,
  }) async {
    return runWithFriendlyErrors(
      () async {
        final start = DateTime.now();
        final fileSize = await imageFile.length();

        final formData = FormData.fromMap({
          'image': await MultipartFile.fromFile(
            imageFile.path,
            filename: imageFile.path.split('/').last,
          ),
          if (format != null && format.isNotEmpty) 'format': format,
          if (itemName != null && itemName.isNotEmpty) 'item_name': itemName,
          if (itemCategory != null && itemCategory.isNotEmpty)
            'item_category': itemCategory,
          if (tags != null && tags.isNotEmpty) 'tags': tags.join(', '),
        });

        final response = await _dio.post(
          ApiConstants.items,
          data: formData,
          options: Options(
            headers: {'Content-Type': 'multipart/form-data'},
            sendTimeout: const Duration(seconds: 180),
            receiveTimeout: const Duration(seconds: 180),
          ),
        );

        if (EnvironmentConfig.verboseLogging) {
          final elapsed = DateTime.now().difference(start);
          final mb = (fileSize / (1024 * 1024)).toStringAsFixed(2);
          debugPrint('🧾 Item upload completed in ${elapsed.inSeconds}s '
              '(${elapsed.inMilliseconds}ms). File size: ${mb}MB');
        }

        final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
          response.data as Map<String, dynamic>,
          (data) => data as Map<String, dynamic>,
        );

        if (!apiResponse.success || apiResponse.data == null) {
          throw FriendlyException(
            apiResponse.error?.message ?? 'Upload failed.',
          );
        }

        return ItemModel.fromJson(apiResponse.data!);
      },
      defaultMessage: 'Upload failed.',
    );
  }

  Future<(List<ItemModel>, ApiMeta?)> listItems({
    int page = 1,
    int perPage = 20,
  }) async {
    return runWithFriendlyErrors(
      () async {
        final response = await _dio.get(
          ApiConstants.items,
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
          throw FriendlyException(
            apiResponse.error?.message ?? 'Failed to load items.',
          );
        }

        final items = apiResponse.data!
            .map((item) => ItemModel.fromJson(item as Map<String, dynamic>))
            .toList();

        return (items, apiResponse.meta);
      },
      defaultMessage: 'Failed to load items.',
    );
  }
}
