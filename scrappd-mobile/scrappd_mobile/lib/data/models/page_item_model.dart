import '../../domain/entities/page_item.dart';

class PageItemModel extends PageItem {
  PageItemModel({
    required super.id,
    required super.pageId,
    required super.itemId,
    required super.positionX,
    required super.positionY,
    required super.width,
    required super.height,
    required super.rotation,
    required super.zIndex,
    required super.opacity,
  });

  factory PageItemModel.fromJson(Map<String, dynamic> json) {
    return PageItemModel(
      id: json['id'] as String,
      pageId: json['page_id'] as String,
      itemId: json['item_id'] as String,
      positionX: (json['position_x'] as num?)?.toDouble() ?? 0.0,
      positionY: (json['position_y'] as num?)?.toDouble() ?? 0.0,
      width: (json['width'] as num?)?.toDouble() ?? 0.0,
      height: (json['height'] as num?)?.toDouble() ?? 0.0,
      rotation: (json['rotation'] as num?)?.toDouble() ?? 0.0,
      zIndex: json['z_index'] as int? ?? 0,
      opacity: (json['opacity'] as num?)?.toDouble() ?? 1.0,
    );
  }
}
