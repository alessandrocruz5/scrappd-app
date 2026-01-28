import 'package:flutter/foundation.dart';
import '../../data/models/page.dart';
import '../../data/services/pages_service.dart';

enum PagesState {
  initial,
  loading,
  loaded,
  error,
  creating,
  updating,
  deleting,
}

class PagesProvider extends ChangeNotifier {
  final PagesService _pagesService;

  PagesState _state = PagesState.initial;
  List<ScrapbookPage> _pages = [];
  ScrapbookPage? _currentPage;
  String? _currentProjectId;
  String? _errorMessage;

  PagesProvider(this._pagesService);

  // Getters
  PagesState get state => _state;
  List<ScrapbookPage> get pages => _pages;
  ScrapbookPage? get currentPage => _currentPage;
  String? get currentProjectId => _currentProjectId;
  String? get errorMessage => _errorMessage;
  bool get isLoading => _state == PagesState.loading;
  int get pageCount => _pages.length;

  // Load pages for a project
  Future<void> loadPages(String projectId, {bool refresh = false}) async {
    if (_state == PagesState.loading && !refresh) return;

    if (projectId != _currentProjectId || refresh) {
      _pages = [];
      _currentProjectId = projectId;
    }

    _setState(PagesState.loading);
    _clearError();

    try {
      final pages = await _pagesService.listPages(projectId);
      _pages = pages..sort((a, b) => a.pageNumber.compareTo(b.pageNumber));
      _setState(PagesState.loaded);
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(PagesState.error);
    }
  }

  // Create a new page
  Future<ScrapbookPage?> createPage({
    required String projectId,
    int? pageNumber,
    String? title,
    int canvasWidth = 1080,
    int canvasHeight = 1920,
    String backgroundColor = '#FFFFFF',
  }) async {
    _setState(PagesState.creating);
    _clearError();

    try {
      // Default page number to next in sequence
      final nextPageNumber = pageNumber ?? (_pages.isEmpty ? 1 : _pages.last.pageNumber + 1);

      final page = await _pagesService.createPage(
        CreatePageRequest(
          projectId: projectId,
          pageNumber: nextPageNumber,
          title: title,
          canvasWidth: canvasWidth,
          canvasHeight: canvasHeight,
          backgroundColor: backgroundColor,
        ),
      );

      _pages.add(page);
      _pages.sort((a, b) => a.pageNumber.compareTo(b.pageNumber));
      _setState(PagesState.loaded);
      return page;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(PagesState.error);
      return null;
    }
  }

  // Get a specific page
  Future<ScrapbookPage?> getPage(String id) async {
    _clearError();

    try {
      final page = await _pagesService.getPage(id);
      _currentPage = page;
      notifyListeners();
      return page;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      return null;
    }
  }

  // Update a page
  Future<ScrapbookPage?> updatePage(String id, UpdatePageRequest request) async {
    _setState(PagesState.updating);
    _clearError();

    try {
      final updatedPage = await _pagesService.updatePage(id, request);

      // Update in list
      final index = _pages.indexWhere((p) => p.id == id);
      if (index != -1) {
        _pages[index] = updatedPage;
        _pages.sort((a, b) => a.pageNumber.compareTo(b.pageNumber));
      }

      // Update current page if it's the same
      if (_currentPage?.id == id) {
        _currentPage = updatedPage;
      }

      _setState(PagesState.loaded);
      return updatedPage;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(PagesState.error);
      return null;
    }
  }

  // Delete a page
  Future<bool> deletePage(String id) async {
    _setState(PagesState.deleting);
    _clearError();

    try {
      await _pagesService.deletePage(id);

      _pages.removeWhere((p) => p.id == id);

      if (_currentPage?.id == id) {
        _currentPage = null;
      }

      _setState(PagesState.loaded);
      return true;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(PagesState.error);
      return false;
    }
  }

  // Set current page
  void setCurrentPage(ScrapbookPage? page) {
    _currentPage = page;
    notifyListeners();
  }

  // Clear current page
  void clearCurrentPage() {
    _currentPage = null;
    notifyListeners();
  }

  // Get page by number
  ScrapbookPage? getPageByNumber(int pageNumber) {
    try {
      return _pages.firstWhere((p) => p.pageNumber == pageNumber);
    } catch (_) {
      return null;
    }
  }

  void _setState(PagesState newState) {
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

  // Reset state (when leaving project or on logout)
  void reset() {
    _state = PagesState.initial;
    _pages = [];
    _currentPage = null;
    _currentProjectId = null;
    _errorMessage = null;
    notifyListeners();
  }
}
