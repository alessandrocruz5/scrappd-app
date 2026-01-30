import 'dart:io';

import '../../domain/entities/item.dart';
import '../../domain/repositories/item_repository.dart';
import '../datasources/items_remote_datasource.dart';

class ItemRepositoryImpl implements ItemRepository {
  ItemRepositoryImpl(this._remoteDataSource);

  final ItemsRemoteDataSource _remoteDataSource;

  @override
  Future<Item> createItem({
    required File imageFile,
    String? itemName,
    String? itemCategory,
    List<String>? tags,
    String? format,
  }) {
    return _remoteDataSource.createItem(
      imageFile: imageFile,
      itemName: itemName,
      itemCategory: itemCategory,
      tags: tags,
      format: format,
    );
  }

  @override
  Future<PagedItems> listItems({int page = 1, int perPage = 20}) async {
    final result = await _remoteDataSource.listItems(
      page: page,
      perPage: perPage,
    );
    final items = result.$1;
    final meta = result.$2;

    return PagedItems(
      items: items,
      page: meta?.page ?? page,
      totalPages: meta?.totalPages ?? page,
      total: meta?.total ?? items.length,
    );
  }
}
