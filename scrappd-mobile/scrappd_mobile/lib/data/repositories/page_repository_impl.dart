import '../../domain/entities/page.dart';
import '../../domain/repositories/page_repository.dart';
import '../datasources/pages_remote_datasource.dart';

class PageRepositoryImpl implements PageRepository {
  PageRepositoryImpl(this._remoteDataSource);

  final PagesRemoteDataSource _remoteDataSource;

  @override
  Future<Page> createPage({
    required String projectId,
    required int pageNumber,
    String? title,
    String? backgroundColor,
  }) {
    return _remoteDataSource.createPage(
      projectId: projectId,
      pageNumber: pageNumber,
      title: title,
      backgroundColor: backgroundColor,
    );
  }

  @override
  Future<PagedPages> listPages({
    required String projectId,
    int page = 1,
    int perPage = 50,
  }) async {
    final (pages, meta) = await _remoteDataSource.listPages(
      projectId: projectId,
      page: page,
      perPage: perPage,
    );

    return PagedPages(
      pages: pages,
      page: meta?.page ?? page,
      totalPages: meta?.totalPages ?? page,
      total: meta?.total ?? pages.length,
    );
  }
}
