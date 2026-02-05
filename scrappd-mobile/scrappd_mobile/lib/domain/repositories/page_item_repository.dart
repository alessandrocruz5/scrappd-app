import '../entities/page_item.dart';

abstract class PageItemRepository {
  Future<List<PageItem>> listPageItems({required String pageId});

  Future<PageItem> createPageItem({
    required String pageId,
    required String itemId,
    required double positionX,
    required double positionY,
    required double width,
    required double height,
    required double rotation,
    int? zIndex,
    double? opacity,
  });

  Future<PageItem> updatePageItem({
    required String pageId,
    required String pageItemId,
    required double positionX,
    required double positionY,
    required double width,
    required double height,
    required double rotation,
    int? zIndex,
    double? opacity,
  });

  Future<void> deletePageItem({
    required String pageId,
    required String pageItemId,
  });
}
