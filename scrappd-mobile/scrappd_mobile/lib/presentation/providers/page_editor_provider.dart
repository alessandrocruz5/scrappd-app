import 'package:flutter/material.dart';

import '../../domain/entities/page.dart' as page_entity;
import '../../domain/entities/page_item.dart';
import '../../domain/repositories/page_item_repository.dart';
import '../../domain/repositories/page_repository.dart';
import '../../core/network/error_helpers.dart';

class PageEditorProvider extends ChangeNotifier {
  PageEditorProvider(this._pageRepository, this._pageItemRepository);

  final PageRepository _pageRepository;
  final PageItemRepository _pageItemRepository;

  bool _isLoading = false;
  page_entity.Page? _currentPage;
  List<PageItem> _pageItems = [];
  String? _errorMessage;

  bool get isLoading => _isLoading;
  page_entity.Page? get currentPage => _currentPage;
  List<PageItem> get pageItems => _pageItems;
  String? get errorMessage => _errorMessage;

  Future<void> loadPageForProject(String projectId) async {
    _setLoading(true);
    try {
      final pagesResult =
          await _pageRepository.listPages(projectId: projectId);
      if (pagesResult.pages.isEmpty) {
        _currentPage = await _pageRepository.createPage(
          projectId: projectId,
          pageNumber: 1,
        );
      } else {
        _currentPage = pagesResult.pages.first;
      }

      _pageItems = await _pageItemRepository.listPageItems(
        pageId: _currentPage!.id,
      );
      _errorMessage = null;
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Failed to load page.',
      );
    } finally {
      _setLoading(false);
    }
  }

  Future<PageItem?> addPageItem({
    required String itemId,
    required double positionX,
    required double positionY,
    required double width,
    required double height,
    required double rotation,
  }) async {
    if (_currentPage == null) return null;
    try {
      final item = await _pageItemRepository.createPageItem(
        pageId: _currentPage!.id,
        itemId: itemId,
        positionX: positionX,
        positionY: positionY,
        width: width,
        height: height,
        rotation: rotation,
      );
      _pageItems = [..._pageItems, item];
      notifyListeners();
      return item;
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Failed to add item.',
      );
      notifyListeners();
      return null;
    }
  }

  void setItemTransform({
    required String pageItemId,
    required double positionX,
    required double positionY,
    required double width,
    required double height,
    required double rotation,
  }) {
    _pageItems = _pageItems
        .map((item) => item.id == pageItemId
            ? item.copyWith(
                positionX: positionX,
                positionY: positionY,
                width: width,
                height: height,
                rotation: rotation,
              )
            : item)
        .toList();
    notifyListeners();
  }

  Future<void> persistItemTransform({
    required String pageItemId,
  }) async {
    if (_currentPage == null) return;
    final item = _pageItems.firstWhere(
      (element) => element.id == pageItemId,
      orElse: () => throw StateError('Page item not found'),
    );

    try {
      final updated = await _pageItemRepository.updatePageItem(
        pageId: _currentPage!.id,
        pageItemId: pageItemId,
        positionX: item.positionX,
        positionY: item.positionY,
        width: item.width,
        height: item.height,
        rotation: item.rotation,
      );
      _pageItems = _pageItems
          .map((existing) =>
              existing.id == updated.id ? updated : existing)
          .toList();
      notifyListeners();
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Failed to update item.',
      );
      notifyListeners();
    }
  }

  Future<void> deletePageItem(String pageItemId) async {
    if (_currentPage == null) return;
    try {
      await _pageItemRepository.deletePageItem(
        pageId: _currentPage!.id,
        pageItemId: pageItemId,
      );
      _pageItems = _pageItems.where((item) => item.id != pageItemId).toList();
      notifyListeners();
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Failed to delete item.',
      );
      notifyListeners();
    }
  }

  void _setLoading(bool value) {
    _isLoading = value;
    notifyListeners();
  }
}
