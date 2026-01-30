import '../../domain/entities/item.dart';

class ItemModel extends Item {
  ItemModel({
    required super.id,
    required super.originalImageUrl,
    required super.processedImageUrl,
    required super.processingStatus,
    required super.itemName,
    required super.itemCategory,
    required super.tags,
    required super.createdAt,
  });

  factory ItemModel.fromJson(Map<String, dynamic> json) {
    return ItemModel(
      id: json['id'] ?? '',
      originalImageUrl: json['original_image_url'] ?? '',
      processedImageUrl: json['processed_image_url'],
      processingStatus: json['processing_status'] ?? 'unknown',
      itemName: json['item_name'],
      itemCategory: json['item_category'],
      tags: (json['tags'] as List<dynamic>? ?? []).cast<String>(),
      createdAt: DateTime.tryParse(json['created_at'] ?? '') ??
          DateTime.fromMillisecondsSinceEpoch(0),
    );
  }
}
