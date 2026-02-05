import '../entities/page.dart';

class PagedPages {
  final List<Page> pages;
  final int page;
  final int totalPages;
  final int total;

  PagedPages({
    required this.pages,
    required this.page,
    required this.totalPages,
    required this.total,
  });
}

abstract class PageRepository {
  Future<Page> createPage({
    required String projectId,
    required int pageNumber,
    String? title,
    String? backgroundColor,
  });

  Future<PagedPages> listPages({
    required String projectId,
    int page,
    int perPage,
  });
}
