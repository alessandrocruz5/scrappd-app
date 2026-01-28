class PageItem {
  final String id;
  final String pageId;
  final String itemId;
  final double positionX;
  final double positionY;
  final double width;
  final double height;
  final double rotation;
  final int zIndex;
  final double opacity;
  final Map<String, dynamic>? filters;
  final DateTime createdAt;
  final DateTime updatedAt;

  // Linked item data (populated from API or cache)
  final Item? item;

  PageItem({
    required this.id,
    required this.pageId,
    required this.itemId,
    required this.positionX,
    required this.positionY,
    required this.width,
    required this.height,
    required this.rotation,
    required this.zIndex,
    required this.opacity,
    this.filters,
    required this.createdAt,
    required this.updatedAt,
    this.item,
  });

  factory PageItem.fromJson(Map<String, dynamic> json) {
    return PageItem(
      id: json['id'] ?? '',
      pageId: json['page_id'] ?? '',
      itemId: json['item_id'] ?? '',
      positionX: (json['position_x'] ?? 0.0).toDouble(),
      positionY: (json['position_y'] ?? 0.0).toDouble(),
      width: (json['width'] ?? 100.0).toDouble(),
      height: (json['height'] ?? 100.0).toDouble(),
      rotation: (json['rotation'] ?? 0.0).toDouble(),
      zIndex: json['z_index'] ?? 0,
      opacity: (json['opacity'] ?? 1.0).toDouble(),
      filters: json['filters'] is Map
          ? Map<String, dynamic>.from(json['filters'])
          : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'])
          : DateTime.now(),
      item: json['item'] != null ? Item.fromJson(json['item']) : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'page_id': pageId,
      'item_id': itemId,
      'position_x': positionX,
      'position_y': positionY,
      'width': width,
      'height': height,
      'rotation': rotation,
      'z_index': zIndex,
      'opacity': opacity,
      'filters': filters,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  PageItem copyWith({
    String? id,
    String? pageId,
    String? itemId,
    double? positionX,
    double? positionY,
    double? width,
    double? height,
    double? rotation,
    int? zIndex,
    double? opacity,
    Map<String, dynamic>? filters,
    DateTime? createdAt,
    DateTime? updatedAt,
    Item? item,
  }) {
    return PageItem(
      id: id ?? this.id,
      pageId: pageId ?? this.pageId,
      itemId: itemId ?? this.itemId,
      positionX: positionX ?? this.positionX,
      positionY: positionY ?? this.positionY,
      width: width ?? this.width,
      height: height ?? this.height,
      rotation: rotation ?? this.rotation,
      zIndex: zIndex ?? this.zIndex,
      opacity: opacity ?? this.opacity,
      filters: filters ?? this.filters,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      item: item ?? this.item,
    );
  }
}

class CreatePageItemRequest {
  final String itemId;
  final double positionX;
  final double positionY;
  final double width;
  final double height;
  final double rotation;
  final int zIndex;
  final double opacity;

  CreatePageItemRequest({
    required this.itemId,
    this.positionX = 0.0,
    this.positionY = 0.0,
    this.width = 200.0,
    this.height = 200.0,
    this.rotation = 0.0,
    this.zIndex = 0,
    this.opacity = 1.0,
  });

  Map<String, dynamic> toJson() {
    return {
      'item_id': itemId,
      'position_x': positionX,
      'position_y': positionY,
      'width': width,
      'height': height,
      'rotation': rotation,
      'z_index': zIndex,
      'opacity': opacity,
    };
  }
}

class UpdatePageItemRequest {
  final double? positionX;
  final double? positionY;
  final double? width;
  final double? height;
  final double? rotation;
  final int? zIndex;
  final double? opacity;
  final Map<String, dynamic>? filters;

  UpdatePageItemRequest({
    this.positionX,
    this.positionY,
    this.width,
    this.height,
    this.rotation,
    this.zIndex,
    this.opacity,
    this.filters,
  });

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
    if (positionX != null) json['position_x'] = positionX;
    if (positionY != null) json['position_y'] = positionY;
    if (width != null) json['width'] = width;
    if (height != null) json['height'] = height;
    if (rotation != null) json['rotation'] = rotation;
    if (zIndex != null) json['z_index'] = zIndex;
    if (opacity != null) json['opacity'] = opacity;
    if (filters != null) json['filters'] = filters;
    return json;
  }
}

class Item {
  final String id;
  final String userId;
  final String originalImageKey;
  final String originalImageUrl;
  final int? originalFileSize;
  final int? originalWidth;
  final int? originalHeight;
  final String? processedImageKey;
  final String? processedImageUrl;
  final int? processedFileSize;
  final String processingStatus;
  final String? mlModelVersion;
  final DateTime? processingStartedAt;
  final DateTime? processingCompletedAt;
  final String? processingError;
  final String? mimeType;
  final String? itemName;
  final String? itemCategory;
  final List<String>? tags;
  final DateTime createdAt;
  final DateTime updatedAt;

  Item({
    required this.id,
    required this.userId,
    required this.originalImageKey,
    required this.originalImageUrl,
    this.originalFileSize,
    this.originalWidth,
    this.originalHeight,
    this.processedImageKey,
    this.processedImageUrl,
    this.processedFileSize,
    required this.processingStatus,
    this.mlModelVersion,
    this.processingStartedAt,
    this.processingCompletedAt,
    this.processingError,
    this.mimeType,
    this.itemName,
    this.itemCategory,
    this.tags,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Item.fromJson(Map<String, dynamic> json) {
    return Item(
      id: json['id'] ?? '',
      userId: json['user_id'] ?? '',
      originalImageKey: json['original_image_key'] ?? '',
      originalImageUrl: json['original_image_url'] ?? '',
      originalFileSize: json['original_file_size_bytes'],
      originalWidth: json['original_width'],
      originalHeight: json['original_height'],
      processedImageKey: json['processed_image_key'],
      processedImageUrl: json['processed_image_url'],
      processedFileSize: json['processed_file_size_bytes'],
      processingStatus: json['processing_status'] ?? 'pending',
      mlModelVersion: json['ml_model_version'],
      processingStartedAt: json['processing_started_at'] != null
          ? DateTime.parse(json['processing_started_at'])
          : null,
      processingCompletedAt: json['processing_completed_at'] != null
          ? DateTime.parse(json['processing_completed_at'])
          : null,
      processingError: json['processing_error'],
      mimeType: json['mime_type'],
      itemName: json['item_name'],
      itemCategory: json['item_category'],
      tags: json['tags'] != null ? List<String>.from(json['tags']) : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'])
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'original_image_key': originalImageKey,
      'original_image_url': originalImageUrl,
      'original_file_size_bytes': originalFileSize,
      'original_width': originalWidth,
      'original_height': originalHeight,
      'processed_image_key': processedImageKey,
      'processed_image_url': processedImageUrl,
      'processed_file_size_bytes': processedFileSize,
      'processing_status': processingStatus,
      'ml_model_version': mlModelVersion,
      'processing_started_at': processingStartedAt?.toIso8601String(),
      'processing_completed_at': processingCompletedAt?.toIso8601String(),
      'processing_error': processingError,
      'mime_type': mimeType,
      'item_name': itemName,
      'item_category': itemCategory,
      'tags': tags,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  /// Returns the best available image URL (processed if available, otherwise original)
  String get displayImageUrl => processedImageUrl ?? originalImageUrl;

  /// Whether processing is complete
  bool get isProcessed => processingStatus == 'completed';

  /// Whether processing failed
  bool get hasFailed => processingStatus == 'failed';

  /// Whether still processing
  bool get isProcessing =>
      processingStatus == 'pending' || processingStatus == 'processing';
}
