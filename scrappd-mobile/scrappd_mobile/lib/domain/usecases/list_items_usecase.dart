import '../repositories/item_repository.dart';

class ListItemsUseCase {
  ListItemsUseCase(this._repository);

  final ItemRepository _repository;

  Future<PagedItems> call({int page = 1, int perPage = 20}) {
    return _repository.listItems(page: page, perPage: perPage);
  }
}
