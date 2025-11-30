class ProcessedImage {
  final String id;
  final String originalPath;
  final String processedPath;
  final RemovalMetadata metadata;
  final DateTime createdAt;

  ProcessedImage({
    required this.id,
    required this.originalPath,
    required this.processedPath,
    required this.metadata,
    required this.createdAt,
  });

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'originalPath': originalPath,
      'processedPath': processedPath,
      'metadata': metadata.toJson(),
      'createdAt': createdAt.toIso8601String(),
    };
  }

  factory ProcessedImage.fromJson(Map<String, dynamic> json) {
    return ProcessedImage(
      id: json['id'],
      originalPath: json['originalPath'],
      processedPath: json['processedPath'],
      metadata: RemovalMetadata.fromJson(json['metadata']),
      createdAt: DateTime.parse(json['createdAt']),
    );
  }
}

class RemovalMetadata {
  final double processingTime;
  final String model;
  final ImageSize originalSize;
  final ImageSize processedSize;

  RemovalMetadata({
    required this.processingTime,
    required this.model,
    required this.originalSize,
    required this.processedSize,
  });

  Map<String, dynamic> toJson() {
    return {
      'processing_time': processingTime,
      'model': model,
      'original_size': originalSize.toJson(),
      'processed_size': processedSize.toJson(),
    };
  }

  factory RemovalMetadata.fromJson(Map<String, dynamic> json) {
    return RemovalMetadata(
      processingTime: (json['processing_time'] ?? 0.0).toDouble(),
      model: json['model'] ?? 'unknown',
      originalSize: ImageSize.fromJson(json['original_size'] ?? {}),
      processedSize: ImageSize.fromJson(json['processed_size'] ?? {}),
    );
  }
}

class ImageSize {
  final int width;
  final int height;

  ImageSize({required this.width, required this.height});

  Map<String, dynamic> toJson() {
    return {
      'width': width,
      'height': height,
    };
  }

  factory ImageSize.fromJson(Map<String, dynamic> json) {
    return ImageSize(
      width: json['width'] ?? 0,
      height: json['height'] ?? 0,
    );
  }
}

class ApiResponse<T> {
  final bool success;
  final T? data;
  final ApiError? error;

  ApiResponse({
    required this.success,
    this.data,
    this.error,
  });

  factory ApiResponse.fromJson(
    Map<String, dynamic> json,
    T Function(dynamic)? fromJsonT,
  ) {
    return ApiResponse(
      success: json['success'] ?? false,
      data: json['data'] != null && fromJsonT != null
          ? fromJsonT(json['data'])
          : null,
      error: json['error'] != null ? ApiError.fromJson(json['error']) : null,
    );
  }
}

class ApiError {
  final String code;
  final String message;

  ApiError({required this.code, required this.message});

  factory ApiError.fromJson(Map<String, dynamic> json) {
    return ApiError(
      code: json['code'] ?? 'UNKNOWN_ERROR',
      message: json['message'] ?? 'An unknown error occurred',
    );
  }
}