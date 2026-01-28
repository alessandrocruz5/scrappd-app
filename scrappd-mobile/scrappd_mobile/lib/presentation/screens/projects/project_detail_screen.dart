import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../../core/constants/theme_constants.dart';
import '../../../data/models/project.dart';
import '../../../data/models/page.dart';
import '../../providers/pages_provider.dart';
import '../../providers/projects_provider.dart';
import '../../widgets/page_thumbnail.dart';
import '../canvas/canvas_editor_screen.dart';

class ProjectDetailScreen extends StatefulWidget {
  final Project project;

  const ProjectDetailScreen({
    super.key,
    required this.project,
  });

  @override
  State<ProjectDetailScreen> createState() => _ProjectDetailScreenState();
}

class _ProjectDetailScreenState extends State<ProjectDetailScreen> {
  late Project _project;

  @override
  void initState() {
    super.initState();
    _project = widget.project;
    _loadPages();
  }

  void _loadPages() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<PagesProvider>().loadPages(_project.id);
    });
  }

  void _navigateToCanvasEditor(ScrapbookPage page) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => CanvasEditorScreen(page: page),
      ),
    );
  }

  Future<void> _handleRefresh() async {
    await context.read<PagesProvider>().loadPages(_project.id, refresh: true);
  }

  void _showAddPageDialog() {
    final titleController = TextEditingController();

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Add New Page'),
        content: TextField(
          controller: titleController,
          decoration: const InputDecoration(
            labelText: 'Page Title (Optional)',
            hintText: 'Enter page title',
          ),
          textCapitalization: TextCapitalization.sentences,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          Consumer<PagesProvider>(
            builder: (context, provider, _) {
              final isCreating = provider.state == PagesState.creating;
              return ElevatedButton(
                onPressed: isCreating
                    ? null
                    : () async {
                        final page = await provider.createPage(
                          projectId: _project.id,
                          title: titleController.text.trim().isEmpty
                              ? null
                              : titleController.text.trim(),
                        );

                        if (page != null && context.mounted) {
                          Navigator.pop(context);
                          _navigateToCanvasEditor(page);
                        }
                      },
                child: isCreating
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Add'),
              );
            },
          ),
        ],
      ),
    );
  }

  void _showPageOptions(ScrapbookPage page) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.edit_outlined),
              title: const Text('Edit Page'),
              onTap: () {
                Navigator.pop(context);
                _navigateToCanvasEditor(page);
              },
            ),
            ListTile(
              leading: const Icon(Icons.delete_outline, color: AppTheme.errorColor),
              title: const Text('Delete Page', style: TextStyle(color: AppTheme.errorColor)),
              onTap: () {
                Navigator.pop(context);
                _confirmDeletePage(page);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _confirmDeletePage(ScrapbookPage page) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Page'),
        content: Text('Are you sure you want to delete page ${page.pageNumber}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              await context.read<PagesProvider>().deletePage(page.id);
            },
            style: TextButton.styleFrom(foregroundColor: AppTheme.errorColor),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
  }

  void _showEditProjectDialog() {
    final titleController = TextEditingController(text: _project.title);
    final descriptionController = TextEditingController(text: _project.description ?? '');

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Edit Project'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: titleController,
              decoration: const InputDecoration(
                labelText: 'Project Title',
                hintText: 'Enter project title',
              ),
              textCapitalization: TextCapitalization.sentences,
            ),
            const SizedBox(height: AppTheme.spacing16),
            TextField(
              controller: descriptionController,
              decoration: const InputDecoration(
                labelText: 'Description (Optional)',
                hintText: 'Enter project description',
              ),
              maxLines: 2,
              textCapitalization: TextCapitalization.sentences,
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          Consumer<ProjectsProvider>(
            builder: (context, provider, _) {
              final isUpdating = provider.state == ProjectsState.updating;
              return ElevatedButton(
                onPressed: isUpdating
                    ? null
                    : () async {
                        final title = titleController.text.trim();
                        if (title.isEmpty) {
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(
                              content: Text('Please enter a project title'),
                            ),
                          );
                          return;
                        }

                        final updated = await provider.updateProject(
                          _project.id,
                          UpdateProjectRequest(
                            title: title,
                            description: descriptionController.text.trim().isEmpty
                                ? null
                                : descriptionController.text.trim(),
                          ),
                        );

                        if (updated != null && context.mounted) {
                          setState(() {
                            _project = updated;
                          });
                          Navigator.pop(context);
                        }
                      },
                child: isUpdating
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Save'),
              );
            },
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_project.title),
        actions: [
          IconButton(
            icon: const Icon(Icons.edit_outlined),
            onPressed: _showEditProjectDialog,
          ),
          PopupMenuButton<String>(
            onSelected: (value) {
              if (value == 'delete') {
                _confirmDeleteProject();
              }
            },
            itemBuilder: (context) => [
              const PopupMenuItem(
                value: 'delete',
                child: Row(
                  children: [
                    Icon(Icons.delete_outline, color: AppTheme.errorColor),
                    SizedBox(width: AppTheme.spacing8),
                    Text('Delete Project',
                        style: TextStyle(color: AppTheme.errorColor)),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
      body: Consumer<PagesProvider>(
        builder: (context, provider, _) {
          if (provider.state == PagesState.initial ||
              (provider.state == PagesState.loading && provider.pages.isEmpty)) {
            return const Center(child: CircularProgressIndicator());
          }

          if (provider.state == PagesState.error && provider.pages.isEmpty) {
            return _buildErrorState(provider);
          }

          return RefreshIndicator(
            onRefresh: _handleRefresh,
            child: CustomScrollView(
              slivers: [
                // Project Info Header
                SliverToBoxAdapter(
                  child: _buildProjectHeader(),
                ),

                // Pages Grid
                if (provider.pages.isEmpty)
                  SliverFillRemaining(
                    child: _buildEmptyPagesState(),
                  )
                else
                  SliverPadding(
                    padding: const EdgeInsets.all(AppTheme.spacing16),
                    sliver: SliverGrid(
                      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 2,
                        mainAxisSpacing: AppTheme.spacing16,
                        crossAxisSpacing: AppTheme.spacing16,
                        childAspectRatio: 0.65,
                      ),
                      delegate: SliverChildBuilderDelegate(
                        (context, index) {
                          final page = provider.pages[index];
                          return PageThumbnail(
                            page: page,
                            onTap: () => _navigateToCanvasEditor(page),
                            onLongPress: () => _showPageOptions(page),
                          );
                        },
                        childCount: provider.pages.length,
                      ),
                    ),
                  ),
              ],
            ),
          );
        },
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _showAddPageDialog,
        backgroundColor: AppTheme.primaryColor,
        child: const Icon(Icons.add, color: Colors.white),
      ),
    );
  }

  Widget _buildProjectHeader() {
    return Container(
      padding: const EdgeInsets.all(AppTheme.spacing16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (_project.description != null && _project.description!.isNotEmpty) ...[
            Text(
              _project.description!,
              style: const TextStyle(
                fontSize: 14,
                color: AppTheme.textSecondary,
              ),
            ),
            const SizedBox(height: AppTheme.spacing12),
          ],
          Row(
            children: [
              _buildInfoChip(
                icon: _getVisibilityIcon(),
                label: _project.visibility.toUpperCase(),
              ),
              const SizedBox(width: AppTheme.spacing8),
              Consumer<PagesProvider>(
                builder: (context, provider, _) {
                  return _buildInfoChip(
                    icon: Icons.auto_stories_outlined,
                    label: '${provider.pageCount} pages',
                  );
                },
              ),
            ],
          ),
          const SizedBox(height: AppTheme.spacing8),
          const Divider(),
        ],
      ),
    );
  }

  Widget _buildInfoChip({required IconData icon, required String label}) {
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppTheme.spacing12,
        vertical: AppTheme.spacing4,
      ),
      decoration: BoxDecoration(
        color: AppTheme.primaryColor.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(AppTheme.radiusSmall),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: AppTheme.primaryColor),
          const SizedBox(width: AppTheme.spacing4),
          Text(
            label,
            style: const TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w500,
              color: AppTheme.primaryColor,
            ),
          ),
        ],
      ),
    );
  }

  IconData _getVisibilityIcon() {
    switch (_project.visibility) {
      case 'public':
        return Icons.public;
      case 'unlisted':
        return Icons.link;
      default:
        return Icons.lock_outline;
    }
  }

  Widget _buildEmptyPagesState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppTheme.spacing32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.note_add_outlined,
              size: 64,
              color: AppTheme.textSecondary.withValues(alpha: 0.5),
            ),
            const SizedBox(height: AppTheme.spacing16),
            const Text(
              'No Pages Yet',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.w600,
                color: AppTheme.textPrimary,
              ),
            ),
            const SizedBox(height: AppTheme.spacing8),
            const Text(
              'Add your first page to start\ncreating your scrapbook!',
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 14,
                color: AppTheme.textSecondary,
              ),
            ),
            const SizedBox(height: AppTheme.spacing24),
            ElevatedButton.icon(
              onPressed: _showAddPageDialog,
              icon: const Icon(Icons.add),
              label: const Text('Add Page'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildErrorState(PagesProvider provider) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppTheme.spacing32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.error_outline,
              size: 64,
              color: AppTheme.errorColor,
            ),
            const SizedBox(height: AppTheme.spacing16),
            Text(
              provider.errorMessage ?? 'Something went wrong',
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 16,
                color: AppTheme.textSecondary,
              ),
            ),
            const SizedBox(height: AppTheme.spacing24),
            ElevatedButton(
              onPressed: () => provider.loadPages(_project.id, refresh: true),
              child: const Text('Try Again'),
            ),
          ],
        ),
      ),
    );
  }

  void _confirmDeleteProject() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Project'),
        content: Text('Are you sure you want to delete "${_project.title}"? This action cannot be undone.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () async {
              Navigator.pop(context);
              final deleted = await context.read<ProjectsProvider>().deleteProject(_project.id);
              if (deleted && mounted) {
                Navigator.pop(context);
              }
            },
            style: TextButton.styleFrom(foregroundColor: AppTheme.errorColor),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
  }
}
