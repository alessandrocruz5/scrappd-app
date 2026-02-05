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
  });

  PageItem copyWith({
    double? positionX,
    double? positionY,
    double? width,
    double? height,
    double? rotation,
    int? zIndex,
    double? opacity,
  }) {
    return PageItem(
      id: id,
      pageId: pageId,
      itemId: itemId,
      positionX: positionX ?? this.positionX,
      positionY: positionY ?? this.positionY,
      width: width ?? this.width,
      height: height ?? this.height,
      rotation: rotation ?? this.rotation,
      zIndex: zIndex ?? this.zIndex,
      opacity: opacity ?? this.opacity,
    );
  }
}
