import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';

import '../../domain/entities/item.dart';
import '../../domain/repositories/item_repository.dart';
import '../../core/utils/image_preprocessor.dart';
import '../../core/network/error_helpers.dart';

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
  DateTime? _startTime;
  Duration _elapsed = Duration.zero;
  int? _uploadBytes;
  Timer? _timer;

  bool _isLoading = false;
  List<Item> _items = [];
  int _page = 1;
  int _totalPages = 1;

  UploadState get uploadState => _uploadState;
  Item? get latestItem => _latestItem;
  String? get errorMessage => _errorMessage;
  Duration get elapsed => _elapsed;
  int? get uploadBytes => _uploadBytes;
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
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Failed to load items.',
      );
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
    _startTimer();
    notifyListeners();

    try {
      final preparedImage = await ImagePreprocessor.prepareForUpload(imageFile);
      _uploadBytes = await preparedImage.length();
      final item = await _repository.createItem(
        imageFile: preparedImage,
        itemName: itemName,
        itemCategory: itemCategory,
        tags: tags,
        format: format,
      );
      _latestItem = item;
      _uploadState = UploadState.success;
      _items = [item, ..._items];
      _stopTimer();
    } on ImageTooLargeException catch (e) {
      _errorMessage = e.toString();
      _uploadState = UploadState.error;
      _stopTimer();
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Upload failed. Please try again.',
      );
      _uploadState = UploadState.error;
      _stopTimer();
    } finally {
      notifyListeners();
    }
  }

  void resetUpload() {
    _uploadState = UploadState.idle;
    _latestItem = null;
    _errorMessage = null;
    _startTime = null;
    _elapsed = Duration.zero;
    _uploadBytes = null;
    _stopTimer();
    notifyListeners();
  }

  void _setLoading(bool value) {
    _isLoading = value;
    notifyListeners();
  }

  void _startTimer() {
    _startTime = DateTime.now();
    _elapsed = Duration.zero;
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (_startTime == null) return;
      _elapsed = DateTime.now().difference(_startTime!);
      notifyListeners();
    });
  }

  void _stopTimer() {
    _timer?.cancel();
    _timer = null;
  }

  @override
  void dispose() {
    _stopTimer();
    super.dispose();
  }
}
