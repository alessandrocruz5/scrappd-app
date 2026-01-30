import 'dart:io';

import 'package:flutter/material.dart';

import '../../domain/entities/item.dart';
import '../../domain/repositories/item_repository.dart';

enum UploadState {
  idle,
  uploading,
  success,
  error,
}

class ItemsProvider extends ChangeNotifier {
  ItemsProvider(this._repository);

  final ItemRepository _repository;

  UploadState _uploadState = UploadState.idle;
  Item? _latestItem;
  String? _errorMessage;

  bool _isLoading = false;
  List<Item> _items = [];
  int _page = 1;
  int _totalPages = 1;

  UploadState get uploadState => _uploadState;
  Item? get latestItem => _latestItem;
  String? get errorMessage => _errorMessage;
  bool get isLoading => _isLoading;
  List<Item> get items => _items;
  int get page => _page;
  int get totalPages => _totalPages;

  Future<void> loadItems({bool refresh = false}) async {
    if (_isLoading) return;

    if (refresh) {
      _page = 1;
      _items = [];
    }

    _setLoading(true);
    try {
      final result = await _repository.listItems(page: _page, perPage: 30);
      _items = result.items;
      _totalPages = result.totalPages;
      _errorMessage = null;
    } catch (e) {
      _errorMessage = e.toString();
    } finally {
      _setLoading(false);
    }
  }

  Future<void> createItem({
    required File imageFile,
    String? itemName,
    String? itemCategory,
    List<String>? tags,
    String? format,
  }) async {
    _uploadState = UploadState.uploading;
    _errorMessage = null;
    notifyListeners();

    try {
      final item = await _repository.createItem(
        imageFile: imageFile,
        itemName: itemName,
        itemCategory: itemCategory,
        tags: tags,
        format: format,
      );
      _latestItem = item;
      _uploadState = UploadState.success;
      _items = [item, ..._items];
    } catch (e) {
      _errorMessage = e.toString();
      _uploadState = UploadState.error;
    } finally {
      notifyListeners();
    }
  }

  void resetUpload() {
    _uploadState = UploadState.idle;
    _latestItem = null;
    _errorMessage = null;
    notifyListeners();
  }

  void _setLoading(bool value) {
    _isLoading = value;
    notifyListeners();
  }
}
