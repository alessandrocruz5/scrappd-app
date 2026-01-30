import 'dart:io';

import '../entities/item.dart';

class PagedItems {
  final List<Item> items;
  final int page;
  final int totalPages;
  final int total;

  PagedItems({
    required this.items,
    required this.page,
    required this.totalPages,
    required this.total,
  });
}

abstract class ItemRepository {
  Future<Item> createItem({
    required File imageFile,
    String? itemName,
    String? itemCategory,
    List<String>? tags,
    String? format,
  });

  Future<PagedItems> listItems({int page, int perPage});
}
