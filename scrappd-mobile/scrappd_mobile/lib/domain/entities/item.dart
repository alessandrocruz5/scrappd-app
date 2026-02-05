class Item {
  final String id;
  final String originalImageUrl;
  final String? processedImageUrl;
  final String processingStatus;
  final String? itemName;
  final String? itemCategory;
  final List<String> tags;
  final DateTime createdAt;

  Item({
    required this.id,
    required this.originalImageUrl,
    required this.processedImageUrl,
    required this.processingStatus,
    required this.itemName,
    required this.itemCategory,
    required this.tags,
    required this.createdAt,
  });
}
