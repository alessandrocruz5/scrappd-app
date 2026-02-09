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
  static const int maxDimension = 2048;
  // Target ~1MB for testing stability.
  static const int maxBytes = 1 * 1024 * 1024;
  static const int _startQuality = 85;
  static const int _minQuality = 60;

  /// Compresses/resizes images that exceed the backend size limit.
  static Future<File> prepareForUpload(File input) async {
    final inputSize = await input.length();
    if (inputSize <= maxBytes) {
      return input;
    }

    final tempDir = await getTemporaryDirectory();
    File current = input;
    var quality = _startQuality;
    var targetDim = maxDimension;

    for (var attempt = 0; attempt < 4; attempt++) {
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
        quality = (quality - 12).clamp(_minQuality, _startQuality);
      } else {
        targetDim = (targetDim * 0.75).round();
      }
    }

    final finalSize = await current.length();
    if (finalSize > maxBytes) {
      throw ImageTooLargeException(
        maxBytes: maxBytes,
        actualBytes: finalSize,
      );
    }

    return current;
  }
}
