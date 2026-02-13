import 'package:flutter/material.dart';

import '../../domain/entities/project.dart';
import '../../domain/repositories/project_repository.dart';
import '../../core/network/error_helpers.dart';

class ProjectsProvider extends ChangeNotifier {
  ProjectsProvider(this._repository);

  final ProjectRepository _repository;

  bool _isLoading = false;
  List<Project> _projects = [];
  String? _errorMessage;

  bool get isLoading => _isLoading;
  List<Project> get projects => _projects;
  String? get errorMessage => _errorMessage;

  Future<void> loadProjects({bool refresh = false}) async {
    if (_isLoading) return;
    _setLoading(true);
    try {
      final result = await _repository.listProjects(page: 1, perPage: 50);
      _projects = result.projects;
      _errorMessage = null;
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Failed to load projects.',
      );
    } finally {
      _setLoading(false);
    }
  }

  Future<Project?> createProject({
    required String title,
    String? description,
  }) async {
    _setLoading(true);
    try {
      final project = await _repository.createProject(
        title: title,
        description: description,
      );
      _projects = [project, ..._projects];
      _errorMessage = null;
      return project;
    } catch (e) {
      _errorMessage = mapErrorMessage(
        e,
        fallback: 'Failed to create project.',
      );
      return null;
    } finally {
      _setLoading(false);
    }
  }

  void _setLoading(bool value) {
    _isLoading = value;
    notifyListeners();
  }
}
