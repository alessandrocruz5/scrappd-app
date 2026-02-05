import '../../domain/entities/page_item.dart';
import '../../domain/repositories/page_item_repository.dart';
import '../datasources/page_items_remote_datasource.dart';

class PageItemRepositoryImpl implements PageItemRepository {
  PageItemRepositoryImpl(this._remoteDataSource);

  final PageItemsRemoteDataSource _remoteDataSource;

  @override
  Future<List<PageItem>> listPageItems({required String pageId}) {
    return _remoteDataSource.listPageItems(pageId: pageId);
  }

  @override
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
  }) {
    return _remoteDataSource.createPageItem(
      pageId: pageId,
      itemId: itemId,
      positionX: positionX,
      positionY: positionY,
      width: width,
      height: height,
      rotation: rotation,
      zIndex: zIndex,
      opacity: opacity,
    );
  }

  @override
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
  }) {
    return _remoteDataSource.updatePageItem(
      pageId: pageId,
      pageItemId: pageItemId,
      positionX: positionX,
      positionY: positionY,
      width: width,
      height: height,
      rotation: rotation,
      zIndex: zIndex,
      opacity: opacity,
    );
  }

  @override
  Future<void> deletePageItem({
    required String pageId,
    required String pageItemId,
  }) {
    return _remoteDataSource.deletePageItem(
      pageId: pageId,
      pageItemId: pageItemId,
    );
  }
}
