import '../entities/project.dart';

class PagedProjects {
  final List<Project> projects;
  final int page;
  final int totalPages;
  final int total;

  PagedProjects({
    required this.projects,
    required this.page,
    required this.totalPages,
    required this.total,
  });
}

abstract class ProjectRepository {
  Future<Project> createProject({required String title, String? description});
  Future<PagedProjects> listProjects({int page, int perPage});
}
