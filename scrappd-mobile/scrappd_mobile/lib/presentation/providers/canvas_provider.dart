import 'package:flutter/foundation.dart';
import '../../data/models/page.dart';
import '../../data/models/page_item.dart';
import '../../data/services/page_items_service.dart';

enum CanvasState {
  initial,
  loading,
  loaded,
  error,
  saving,
}

enum CanvasMode {
  select,
  pan,
}

class CanvasProvider extends ChangeNotifier {
  final PageItemsService _pageItemsService;

  CanvasState _state = CanvasState.initial;
  ScrapbookPage? _page;
  List<PageItem> _items = [];
  PageItem? _selectedItem;
  String? _errorMessage;
  CanvasMode _mode = CanvasMode.select;

  // Undo/Redo stacks
  final List<List<PageItem>> _undoStack = [];
  final List<List<PageItem>> _redoStack = [];
  static const int _maxUndoStackSize = 50;

  // Transform state
  double _scale = 1.0;
  double _offsetX = 0.0;
  double _offsetY = 0.0;

  CanvasProvider(this._pageItemsService);

  // Getters
  CanvasState get state => _state;
  ScrapbookPage? get page => _page;
  List<PageItem> get items => List.unmodifiable(_items);
  PageItem? get selectedItem => _selectedItem;
  String? get errorMessage => _errorMessage;
  CanvasMode get mode => _mode;
  double get scale => _scale;
  double get offsetX => _offsetX;
  double get offsetY => _offsetY;
  bool get isLoading => _state == CanvasState.loading;
  bool get isSaving => _state == CanvasState.saving;
  bool get canUndo => _undoStack.isNotEmpty;
  bool get canRedo => _redoStack.isNotEmpty;

  // Load page items
  Future<void> loadPage(ScrapbookPage page) async {
    _page = page;
    _setState(CanvasState.loading);
    _clearError();

    try {
      final items = await _pageItemsService.listPageItems(page.id);
      _items = items..sort((a, b) => a.zIndex.compareTo(b.zIndex));
      _selectedItem = null;
      _undoStack.clear();
      _redoStack.clear();
      _resetTransform();
      _setState(CanvasState.loaded);
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(CanvasState.error);
    }
  }

  // Add item to canvas
  Future<PageItem?> addItem(String itemId, {double? x, double? y}) async {
    if (_page == null) return null;

    _saveToUndoStack();
    _setState(CanvasState.saving);

    try {
      // Calculate center position if not provided
      final posX = x ?? (_page!.canvasWidth / 2 - 100);
      final posY = y ?? (_page!.canvasHeight / 2 - 100);

      final newItem = await _pageItemsService.addItemToPage(
        _page!.id,
        CreatePageItemRequest(
          itemId: itemId,
          positionX: posX,
          positionY: posY,
          zIndex: _items.isEmpty ? 0 : _items.last.zIndex + 1,
        ),
      );

      _items.add(newItem);
      _items.sort((a, b) => a.zIndex.compareTo(b.zIndex));
      _selectedItem = newItem;
      _redoStack.clear();
      _setState(CanvasState.loaded);
      return newItem;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _undoStack.removeLast(); // Remove failed save from undo
      _setState(CanvasState.loaded);
      return null;
    }
  }

  // Update item transform (position, size, rotation)
  Future<bool> updateItemTransform(
    String itemId, {
    double? positionX,
    double? positionY,
    double? width,
    double? height,
    double? rotation,
  }) async {
    if (_page == null) return false;

    final index = _items.indexWhere((i) => i.id == itemId);
    if (index == -1) return false;

    _saveToUndoStack();

    // Optimistic update
    final oldItem = _items[index];
    _items[index] = oldItem.copyWith(
      positionX: positionX ?? oldItem.positionX,
      positionY: positionY ?? oldItem.positionY,
      width: width ?? oldItem.width,
      height: height ?? oldItem.height,
      rotation: rotation ?? oldItem.rotation,
    );
    notifyListeners();

    try {
      final updated = await _pageItemsService.updatePageItem(
        _page!.id,
        itemId,
        UpdatePageItemRequest(
          positionX: positionX,
          positionY: positionY,
          width: width,
          height: height,
          rotation: rotation,
        ),
      );

      _items[index] = updated;
      if (_selectedItem?.id == itemId) {
        _selectedItem = updated;
      }
      _redoStack.clear();
      notifyListeners();
      return true;
    } catch (e) {
      // Rollback on error
      _items[index] = oldItem;
      _undoStack.removeLast();
      _setError(e.toString().replaceFirst('Exception: ', ''));
      notifyListeners();
      return false;
    }
  }

  // Update item properties (z-index, opacity, filters)
  Future<bool> updateItemProperties(
    String itemId, {
    int? zIndex,
    double? opacity,
    Map<String, dynamic>? filters,
  }) async {
    if (_page == null) return false;

    final index = _items.indexWhere((i) => i.id == itemId);
    if (index == -1) return false;

    _saveToUndoStack();

    final oldItem = _items[index];

    try {
      final updated = await _pageItemsService.updatePageItem(
        _page!.id,
        itemId,
        UpdatePageItemRequest(
          zIndex: zIndex,
          opacity: opacity,
          filters: filters,
        ),
      );

      _items[index] = updated;
      _items.sort((a, b) => a.zIndex.compareTo(b.zIndex));
      if (_selectedItem?.id == itemId) {
        _selectedItem = updated;
      }
      _redoStack.clear();
      notifyListeners();
      return true;
    } catch (e) {
      _items[index] = oldItem;
      _undoStack.removeLast();
      _setError(e.toString().replaceFirst('Exception: ', ''));
      notifyListeners();
      return false;
    }
  }

  // Remove item from canvas
  Future<bool> removeItem(String itemId) async {
    if (_page == null) return false;

    final index = _items.indexWhere((i) => i.id == itemId);
    if (index == -1) return false;

    _saveToUndoStack();

    final removedItem = _items.removeAt(index);
    if (_selectedItem?.id == itemId) {
      _selectedItem = null;
    }
    notifyListeners();

    try {
      await _pageItemsService.removeItemFromPage(_page!.id, itemId);
      _redoStack.clear();
      return true;
    } catch (e) {
      // Rollback on error
      _items.insert(index, removedItem);
      _undoStack.removeLast();
      _setError(e.toString().replaceFirst('Exception: ', ''));
      notifyListeners();
      return false;
    }
  }

  // Selection
  void selectItem(PageItem? item) {
    _selectedItem = item;
    notifyListeners();
  }

  void clearSelection() {
    _selectedItem = null;
    notifyListeners();
  }

  // Canvas mode
  void setMode(CanvasMode mode) {
    _mode = mode;
    notifyListeners();
  }

  // Transform controls
  void setScale(double scale) {
    _scale = scale.clamp(0.1, 5.0);
    notifyListeners();
  }

  void setOffset(double x, double y) {
    _offsetX = x;
    _offsetY = y;
    notifyListeners();
  }

  void _resetTransform() {
    _scale = 1.0;
    _offsetX = 0.0;
    _offsetY = 0.0;
  }

  void resetView() {
    _resetTransform();
    notifyListeners();
  }

  void zoomIn() {
    setScale(_scale * 1.2);
  }

  void zoomOut() {
    setScale(_scale / 1.2);
  }

  // Layer management
  void bringToFront(String itemId) {
    final maxZ = _items.isEmpty ? 0 : _items.map((i) => i.zIndex).reduce((a, b) => a > b ? a : b);
    updateItemProperties(itemId, zIndex: maxZ + 1);
  }

  void sendToBack(String itemId) {
    final minZ = _items.isEmpty ? 0 : _items.map((i) => i.zIndex).reduce((a, b) => a < b ? a : b);
    updateItemProperties(itemId, zIndex: minZ - 1);
  }

  void bringForward(String itemId) {
    final index = _items.indexWhere((i) => i.id == itemId);
    if (index == -1 || index >= _items.length - 1) return;

    final currentZ = _items[index].zIndex;
    final nextZ = _items[index + 1].zIndex;
    updateItemProperties(itemId, zIndex: nextZ + 1);
  }

  void sendBackward(String itemId) {
    final index = _items.indexWhere((i) => i.id == itemId);
    if (index <= 0) return;

    final currentZ = _items[index].zIndex;
    final prevZ = _items[index - 1].zIndex;
    updateItemProperties(itemId, zIndex: prevZ - 1);
  }

  // Undo/Redo
  void _saveToUndoStack() {
    _undoStack.add(List.from(_items.map((i) => i.copyWith())));
    if (_undoStack.length > _maxUndoStackSize) {
      _undoStack.removeAt(0);
    }
  }

  void undo() {
    if (_undoStack.isEmpty) return;

    _redoStack.add(List.from(_items.map((i) => i.copyWith())));
    _items = _undoStack.removeLast();
    _selectedItem = null;
    notifyListeners();
  }

  void redo() {
    if (_redoStack.isEmpty) return;

    _undoStack.add(List.from(_items.map((i) => i.copyWith())));
    _items = _redoStack.removeLast();
    _selectedItem = null;
    notifyListeners();
  }

  // Duplicate selected item
  Future<void> duplicateSelected() async {
    if (_selectedItem == null) return;

    await addItem(
      _selectedItem!.itemId,
      x: _selectedItem!.positionX + 20,
      y: _selectedItem!.positionY + 20,
    );
  }

  // Delete selected item
  Future<void> deleteSelected() async {
    if (_selectedItem == null) return;
    await removeItem(_selectedItem!.id);
  }

  void _setState(CanvasState newState) {
    _state = newState;
    notifyListeners();
  }

  void _setError(String message) {
    _errorMessage = message;
  }

  void _clearError() {
    _errorMessage = null;
  }

  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }

  // Reset state
  void reset() {
    _state = CanvasState.initial;
    _page = null;
    _items = [];
    _selectedItem = null;
    _errorMessage = null;
    _mode = CanvasMode.select;
    _undoStack.clear();
    _redoStack.clear();
    _resetTransform();
    notifyListeners();
  }
}
