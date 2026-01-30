import '../../domain/entities/project.dart';
import '../../domain/repositories/project_repository.dart';
import '../datasources/projects_remote_datasource.dart';

class ProjectRepositoryImpl implements ProjectRepository {
  ProjectRepositoryImpl(this._remoteDataSource);

  final ProjectsRemoteDataSource _remoteDataSource;

  @override
  Future<Project> createProject({
    required String title,
    String? description,
  }) {
    return _remoteDataSource.createProject(
      title: title,
      description: description,
    );
  }

  @override
  Future<PagedProjects> listProjects({int page = 1, int perPage = 50}) async {
    final (projects, meta) = await _remoteDataSource.listProjects(
      page: page,
      perPage: perPage,
    );

    return PagedProjects(
      projects: projects,
      page: meta?.page ?? page,
      totalPages: meta?.totalPages ?? page,
      total: meta?.total ?? projects.length,
    );
  }
}
