import 'package:flutter/foundation.dart';
import '../../data/models/project.dart';
import '../../data/services/projects_service.dart';

enum ProjectsState {
  initial,
  loading,
  loaded,
  error,
  creating,
  updating,
  deleting,
}

class ProjectsProvider extends ChangeNotifier {
  final ProjectsService _projectsService;

  ProjectsState _state = ProjectsState.initial;
  List<Project> _projects = [];
  Project? _currentProject;
  String? _errorMessage;
  int _currentPage = 1;
  int _totalProjects = 0;
  bool _hasMore = true;
  static const int _perPage = 20;

  ProjectsProvider(this._projectsService);

  // Getters
  ProjectsState get state => _state;
  List<Project> get projects => _projects;
  Project? get currentProject => _currentProject;
  String? get errorMessage => _errorMessage;
  bool get isLoading => _state == ProjectsState.loading;
  bool get hasMore => _hasMore;
  int get totalProjects => _totalProjects;

  // Load projects (initial load or refresh)
  Future<void> loadProjects({bool refresh = false}) async {
    if (_state == ProjectsState.loading) return;

    if (refresh) {
      _currentPage = 1;
      _projects = [];
      _hasMore = true;
    }

    _setState(ProjectsState.loading);
    _clearError();

    try {
      final response = await _projectsService.listProjects(
        page: _currentPage,
        perPage: _perPage,
      );

      _projects = response.projects;
      _totalProjects = response.total;
      _hasMore = response.hasMore;
      _currentPage = 1;
      _setState(ProjectsState.loaded);
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(ProjectsState.error);
    }
  }

  // Load more projects (pagination)
  Future<void> loadMoreProjects() async {
    if (_state == ProjectsState.loading || !_hasMore) return;

    _setState(ProjectsState.loading);

    try {
      final nextPage = _currentPage + 1;
      final response = await _projectsService.listProjects(
        page: nextPage,
        perPage: _perPage,
      );

      _projects.addAll(response.projects);
      _totalProjects = response.total;
      _hasMore = response.hasMore;
      _currentPage = nextPage;
      _setState(ProjectsState.loaded);
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(ProjectsState.error);
    }
  }

  // Create a new project
  Future<Project?> createProject({
    required String title,
    String? description,
    String visibility = 'private',
  }) async {
    _setState(ProjectsState.creating);
    _clearError();

    try {
      final project = await _projectsService.createProject(
        CreateProjectRequest(
          title: title,
          description: description,
          visibility: visibility,
        ),
      );

      _projects.insert(0, project);
      _totalProjects++;
      _setState(ProjectsState.loaded);
      return project;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(ProjectsState.error);
      return null;
    }
  }

  // Get a specific project
  Future<Project?> getProject(String id) async {
    _clearError();

    try {
      final project = await _projectsService.getProject(id);
      _currentProject = project;
      notifyListeners();
      return project;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      return null;
    }
  }

  // Update a project
  Future<Project?> updateProject(String id, UpdateProjectRequest request) async {
    _setState(ProjectsState.updating);
    _clearError();

    try {
      final updatedProject = await _projectsService.updateProject(id, request);

      // Update in list
      final index = _projects.indexWhere((p) => p.id == id);
      if (index != -1) {
        _projects[index] = updatedProject;
      }

      // Update current project if it's the same
      if (_currentProject?.id == id) {
        _currentProject = updatedProject;
      }

      _setState(ProjectsState.loaded);
      return updatedProject;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(ProjectsState.error);
      return null;
    }
  }

  // Delete a project
  Future<bool> deleteProject(String id) async {
    _setState(ProjectsState.deleting);
    _clearError();

    try {
      await _projectsService.deleteProject(id);

      _projects.removeWhere((p) => p.id == id);
      _totalProjects--;

      if (_currentProject?.id == id) {
        _currentProject = null;
      }

      _setState(ProjectsState.loaded);
      return true;
    } catch (e) {
      _setError(e.toString().replaceFirst('Exception: ', ''));
      _setState(ProjectsState.error);
      return false;
    }
  }

  // Set current project
  void setCurrentProject(Project? project) {
    _currentProject = project;
    notifyListeners();
  }

  // Clear current project
  void clearCurrentProject() {
    _currentProject = null;
    notifyListeners();
  }

  void _setState(ProjectsState newState) {
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

  // Reset state (on logout)
  void reset() {
    _state = ProjectsState.initial;
    _projects = [];
    _currentProject = null;
    _errorMessage = null;
    _currentPage = 1;
    _totalProjects = 0;
    _hasMore = true;
    notifyListeners();
  }
}
