import 'dart:async';
import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:path_provider/path_provider.dart';

import '../../domain/entities/item.dart';
import '../../domain/repositories/item_repository.dart';
import '../../core/utils/image_preprocessor.dart';
import '../../core/network/error_helpers.dart';

enum UploadState { idle, uploading, success, error }

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
  Timer? _pollTimer;
  int _pollAttempts = 0;
  Duration _pollDelay = const Duration(seconds: 2);

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
      _errorMessage = mapErrorMessage(e, fallback: 'Failed to load items.');
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
      _startPollingItem(item.id);
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
    _stopPolling();
    notifyListeners();
  }

  Future<void> retryItem(Item item) async {
    if (item.originalImageUrl.isEmpty) {
      _errorMessage = 'Original image not available for retry.';
      notifyListeners();
      return;
    }

    _uploadState = UploadState.uploading;
    _errorMessage = null;
    notifyListeners();

    try {
      final file = await _downloadToTempFile(item.originalImageUrl);
      await createItem(
        imageFile: file,
        itemName: item.itemName,
        itemCategory: item.itemCategory,
        tags: item.tags,
        format: 'png',
      );
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Retry failed. Please try again.',
      );
      _uploadState = UploadState.error;
      notifyListeners();
    }
  }

  Future<void> cancelItem(String itemId) async {
    try {
      await _repository.cancelItem(itemId);
      _replaceItemStatus(itemId, 'failed');
      notifyListeners();
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Failed to cancel processing.',
      );
      notifyListeners();
    }
  }

  Future<void> deleteItem(String itemId) async {
    try {
      await _repository.deleteItem(itemId);
      _items = _items.where((item) => item.id != itemId).toList();
      if (_latestItem?.id == itemId) {
        _latestItem = null;
      }
      notifyListeners();
    } catch (e) {
      _errorMessage = mapErrorMessage(e, fallback: 'Failed to delete item.');
      notifyListeners();
    }
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

  void _startPollingItem(String itemId) {
    _pollAttempts = 0;
    _pollDelay = const Duration(seconds: 2);
    _pollTimer?.cancel();
    _schedulePoll(itemId);
  }

  void _stopPolling() {
    _pollTimer?.cancel();
    _pollTimer = null;
  }

  void _schedulePoll(String itemId) {
    _pollTimer?.cancel();
    _pollTimer = Timer(_pollDelay, () async {
      _pollAttempts++;
      if (_pollAttempts > 90) {
        _stopPolling();
        return;
      }

      try {
        final result = await _repository.listItems(page: 1, perPage: 30);
        _items = result.items;
        _totalPages = result.totalPages;
        final updated = result.items
            .where((item) => item.id == itemId)
            .toList();
        if (updated.isNotEmpty) {
          _latestItem = updated.first;
          if (_latestItem!.processingStatus == 'completed' ||
              _latestItem!.processingStatus == 'failed') {
            _stopPolling();
            notifyListeners();
            return;
          }
        }
        notifyListeners();
      } catch (_) {
        // Ignore polling errors; try again until timeout.
      }

      final nextSeconds = (_pollDelay.inSeconds + 2).clamp(2, 15);
      _pollDelay = Duration(seconds: nextSeconds);
      _schedulePoll(itemId);
    });
  }

  Future<File> _downloadToTempFile(String url) async {
    final uri = Uri.parse(url);
    final client = HttpClient();
    final request = await client.getUrl(uri);
    final response = await request.close();
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to download image (${response.statusCode})');
    }

    final bytes = await consolidateHttpClientResponseBytes(response);
    final tempDir = await getTemporaryDirectory();
    final fileName = 'retry_${DateTime.now().millisecondsSinceEpoch}.jpg';
    final file = File('${tempDir.path}/$fileName');
    await file.writeAsBytes(bytes);
    return file;
  }

  void _replaceItemStatus(String itemId, String status) {
    Item? updated;
    _items = _items.map((item) {
      if (item.id != itemId) return item;
      updated = Item(
        id: item.id,
        originalImageUrl: item.originalImageUrl,
        processedImageUrl: item.processedImageUrl,
        processingStatus: status,
        itemName: item.itemName,
        itemCategory: item.itemCategory,
        tags: item.tags,
        createdAt: item.createdAt,
      );
      return updated!;
    }).toList();

    if (_latestItem?.id == itemId && updated != null) {
      _latestItem = updated;
    }
  }

  @override
  void dispose() {
    _stopTimer();
    _stopPolling();
    super.dispose();
  }
}
