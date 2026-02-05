import 'dart:io';

import '../entities/item.dart';
import '../repositories/item_repository.dart';

class CreateItemUseCase {
  CreateItemUseCase(this._repository);

  final ItemRepository _repository;

  Future<Item> call({
    required File imageFile,
    String? itemName,
    String? itemCategory,
    List<String>? tags,
    String? format,
  }) {
    return _repository.createItem(
      imageFile: imageFile,
      itemName: itemName,
      itemCategory: itemCategory,
      tags: tags,
      format: format,
    );
  }
}
