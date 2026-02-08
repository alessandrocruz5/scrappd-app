import 'dart:io';
import 'package:flutter_image_compress/flutter_image_compress.dart';
import 'package:path_provider/path_provider.dart';

import '../constants/api_constants.dart';

class ImageTooLargeException implements Exception {
  ImageTooLargeException({
    required this.maxBytes,
    required this.actualBytes,
  });

  final int maxBytes;
  final int actualBytes;

  @override
  String toString() {
    return 'Image is too large (${_formatBytes(actualBytes)}). '
        'Max allowed is ${_formatBytes(maxBytes)}.';
  }

  String _formatBytes(int bytes) {
    final mb = bytes / (1024 * 1024);
    return '${mb.toStringAsFixed(1)}MB';
  }
}

class ImagePreprocessor {
  static const bool enforceLimits = false;
  static const int maxDimension = 4096;
  static const int maxBytes = ApiConstants.maxFileSize;
  static const int _startQuality = 92;
  static const int _minQuality = 70;

  /// Compresses/resizes images that exceed the backend size limit.
  static Future<File> prepareForUpload(File input) async {
    if (!enforceLimits) {
      return input;
    }

    final inputSize = await input.length();
    if (inputSize <= maxBytes) {
      return input;
    }

    final tempDir = await getTemporaryDirectory();
    File current = input;
    var quality = _startQuality;
    var targetDim = maxDimension;

    for (var attempt = 0; attempt < 5; attempt++) {
      final outPath =
          '${tempDir.path}/upload_${DateTime.now().millisecondsSinceEpoch}_$attempt.jpg';

      final result = await FlutterImageCompress.compressAndGetFile(
        current.path,
        outPath,
        quality: quality,
        minWidth: targetDim,
        minHeight: targetDim,
        format: CompressFormat.jpeg,
      );

      if (result != null) {
        current = File(result.path);
        if (await current.length() <= maxBytes) {
          return current;
        }
      }

      if (quality > _minQuality) {
        quality = (quality - 10).clamp(_minQuality, _startQuality) as int;
      } else {
        targetDim = (targetDim * 0.8).round();
      }
    }

    return current;
  }
}
