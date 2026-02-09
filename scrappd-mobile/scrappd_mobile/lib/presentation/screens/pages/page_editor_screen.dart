import 'dart:math' as math;
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:gal/gal.dart';
import 'package:path_provider/path_provider.dart';
import 'package:share_plus/share_plus.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

import '../../../core/constants/theme_constants.dart';
import '../../providers/items_provider.dart';
import '../../providers/page_editor_provider.dart';
import '../../providers/projects_provider.dart';
import '../../../data/services/page_export_service.dart';
import '../../../domain/entities/page.dart' as page_entity;

class PageEditorScreen extends StatefulWidget {
  const PageEditorScreen({super.key});

  @override
  State<PageEditorScreen> createState() => _PageEditorScreenState();
}

const double _baseItemSize = 140;

class _PageEditorScreenState extends State<PageEditorScreen> {
  final GlobalKey _canvasKey = GlobalKey();
  _PageTemplate _activeTemplate = _PageTemplate.clean();
  String? _selectedProjectId;
  bool _autoSelected = false;

  Offset _gestureStart = Offset.zero;
  Offset _itemStart = Offset.zero;
  double _widthStart = _baseItemSize;
  double _heightStart = _baseItemSize;
  double _rotationStart = 0.0;
  bool _isExporting = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<ItemsProvider>().loadItems(refresh: true);
      context.read<ProjectsProvider>().loadProjects();
    });
  }

  Future<void> _openItemPicker(Size canvasSize) async {
    final provider = context.read<ItemsProvider>();
    if (provider.items.isEmpty && !provider.isLoading) {
      provider.loadItems(refresh: true);
    }

    showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      isScrollControlled: true,
      builder: (context) {
        return Padding(
          padding: const EdgeInsets.all(AppTheme.spacing16),
          child: Consumer<ItemsProvider>(
            builder: (context, itemsProvider, _) {
              if (itemsProvider.isLoading) {
                return const SizedBox(
                  height: 200,
                  child: Center(child: CircularProgressIndicator()),
                );
              }

              if (itemsProvider.items.isEmpty) {
                return const SizedBox(
                  height: 200,
                  child: Center(child: Text('No items yet. Upload one first.')),
                );
              }

              return SizedBox(
                height: 360,
                child: GridView.builder(
                  itemCount: itemsProvider.items.length,
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 3,
                    crossAxisSpacing: AppTheme.spacing12,
                    mainAxisSpacing: AppTheme.spacing12,
                  ),
                  itemBuilder: (context, index) {
                    final item = itemsProvider.items[index];
                    final imageUrl =
                        item.processedImageUrl ?? item.originalImageUrl;
                    return GestureDetector(
                      onTap: () {
                        Navigator.pop(context);
                        _addItem(item.id, canvasSize);
                      },
                      child: ClipRRect(
                        borderRadius:
                            BorderRadius.circular(AppTheme.radiusMedium),
                        child: Image.network(
                          imageUrl,
                          fit: BoxFit.cover,
                        ),
                      ),
                    );
                  },
                ),
              );
            },
          ),
        );
      },
    );
  }

  Future<void> _addItem(String itemId, Size canvasSize) async {
    final pageEditor = context.read<PageEditorProvider>();
    if (pageEditor.currentPage == null) return;

    final position = Offset(
      canvasSize.width * 0.5 - _baseItemSize / 2,
      canvasSize.height * 0.3 - _baseItemSize / 2,
    );

    await pageEditor.addPageItem(
      itemId: itemId,
      positionX: position.dx,
      positionY: position.dy,
      width: _baseItemSize,
      height: _baseItemSize,
      rotation: 0.0,
    );
  }

  void _applyTemplate(_PageTemplate template) {
    setState(() {
      _activeTemplate = template;
    });
  }

  Future<void> _showExportSheet() async {
    final currentPage = context.read<PageEditorProvider>().currentPage;
    if (currentPage == null || _isExporting) return;

    final presets = _ExportPreset.defaults;
    await showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      isScrollControlled: true,
      builder: (context) {
        return SafeArea(
          child: ConstrainedBox(
            constraints: BoxConstraints(
              maxHeight: MediaQuery.of(context).size.height * 0.75,
            ),
            child: ListView(
              padding: const EdgeInsets.all(AppTheme.spacing16),
              children: [
                const SizedBox(height: AppTheme.spacing8),
                const Text(
                  'Export page',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 18,
                  ),
                ),
                const SizedBox(height: AppTheme.spacing12),
                ...presets.map(
                  (preset) => ListTile(
                    title: Text(preset.label),
                    subtitle: Text('${preset.width} x ${preset.height}'),
                    trailing: const Icon(Icons.download),
                    onTap: () {
                      Navigator.pop(context);
                      _exportPage(preset, currentPage.id);
                    },
                  ),
                ),
                ListTile(
                  title: const Text('Custom size'),
                  subtitle: const Text('Set your own width and height'),
                  trailing: const Icon(Icons.tune),
                  onTap: () {
                    Navigator.pop(context);
                    _showCustomExportDialog(currentPage);
                  },
                ),
                const SizedBox(height: AppTheme.spacing8),
              ],
            ),
          ),
        );
      },
    );
  }

  Future<void> _showCustomExportDialog(page_entity.Page page) async {
    final widthController =
        TextEditingController(text: page.canvasWidth.toString());
    final heightController =
        TextEditingController(text: page.canvasHeight.toString());
    String format = 'jpeg';
    int quality = 92;

    final result = await showDialog<_ExportPreset>(
      context: context,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            return AlertDialog(
              title: const Text('Custom export'),
              content: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    TextField(
                      controller: widthController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        labelText: 'Width (px)',
                      ),
                    ),
                    const SizedBox(height: AppTheme.spacing12),
                    TextField(
                      controller: heightController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        labelText: 'Height (px)',
                      ),
                    ),
                    const SizedBox(height: AppTheme.spacing12),
                    DropdownButtonFormField<String>(
                      initialValue: format,
                      decoration: const InputDecoration(labelText: 'Format'),
                      items: const [
                        DropdownMenuItem(value: 'jpeg', child: Text('JPEG')),
                        DropdownMenuItem(value: 'png', child: Text('PNG')),
                      ],
                      onChanged: (value) {
                        setDialogState(() {
                          format = value ?? 'jpeg';
                        });
                      },
                    ),
                    const SizedBox(height: AppTheme.spacing12),
                    DropdownButtonFormField<int>(
                      initialValue: quality,
                      decoration: const InputDecoration(labelText: 'Quality'),
                      items: const [
                        DropdownMenuItem(value: 85, child: Text('85')),
                        DropdownMenuItem(value: 92, child: Text('92')),
                        DropdownMenuItem(value: 98, child: Text('98')),
                      ],
                      onChanged: (value) {
                        setDialogState(() {
                          quality = value ?? 92;
                        });
                      },
                    ),
                  ],
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('Cancel'),
                ),
                ElevatedButton(
                  onPressed: () {
                    final width = int.tryParse(widthController.text.trim()) ?? 0;
                    final height =
                        int.tryParse(heightController.text.trim()) ?? 0;
                    if (width <= 0 || height <= 0) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(
                          content: Text('Enter valid width and height'),
                          backgroundColor: AppTheme.errorColor,
                        ),
                      );
                      return;
                    }
                    Navigator.pop(
                      context,
                      _ExportPreset(
                        label: 'Custom',
                        width: width,
                        height: height,
                        format: format,
                        scale: 2.0,
                        quality: quality,
                      ),
                    );
                  },
                  child: const Text('Export'),
                ),
              ],
            );
          },
        );
      },
    );

    if (result != null) {
      _exportPage(result, page.id);
    }
  }

  Future<void> _exportPage(_ExportPreset preset, String pageId) async {
    if (_isExporting) return;
    setState(() {
      _isExporting = true;
    });

    final exportService = context.read<PageExportService>();

    showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (_) => const Center(child: CircularProgressIndicator()),
    );

    try {
      final bytes = await exportService.exportPage(
        pageId: pageId,
        scale: preset.scale,
        format: preset.format,
        width: preset.width,
        height: preset.height,
        quality: preset.quality,
      );

      final tempDir = await getTemporaryDirectory();
      final extension = preset.format == 'jpeg' ? 'jpg' : preset.format;
      final file = File(
        '${tempDir.path}/scrappd_page_${pageId}_${preset.width}x${preset.height}.$extension',
      );
      await file.writeAsBytes(bytes);

      if (kIsWeb || !(Platform.isAndroid || Platform.isIOS)) {
        throw Exception(
          'Saving to gallery is only supported on Android and iOS.',
        );
      }

      await Gal.putImage(file.path);

      if (context.mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: const Text('Page saved to gallery'),
            backgroundColor: AppTheme.successColor,
            action: SnackBarAction(
              label: 'Share',
              textColor: Colors.white,
              onPressed: () => _shareFile(file),
            ),
          ),
        );
      }
    } catch (e) {
      if (context.mounted) {
        Navigator.pop(context);
        final message = e is MissingPluginException
            ? 'Save failed. Please fully restart the app after adding plugins.'
            : 'Failed to export page: $e';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(message),
            backgroundColor: AppTheme.errorColor,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isExporting = false;
        });
      }
    }
  }

  Future<void> _shareFile(File file) async {
    try {
      await Share.shareXFiles([XFile(file.path)], text: 'Scrappd page');
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to share: $e'),
            backgroundColor: AppTheme.errorColor,
          ),
        );
      }
    }
  }

  Future<void> _createProject() async {
    final controller = TextEditingController();
    final created = await showDialog<String>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('New project'),
          content: TextField(
            controller: controller,
            decoration: const InputDecoration(
              labelText: 'Project title',
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancel'),
            ),
            ElevatedButton(
              onPressed: () {
                Navigator.pop(context, controller.text.trim());
              },
              child: const Text('Create'),
            ),
          ],
        );
      },
    );

    if (created == null || created.isEmpty) return;
    final projectsProvider = context.read<ProjectsProvider>();
    final project =
        await projectsProvider.createProject(title: created, description: null);
    if (project == null) return;

    setState(() {
      _selectedProjectId = project.id;
    });

    await context.read<PageEditorProvider>().loadPageForProject(project.id);
  }

  @override
  Widget build(BuildContext context) {
    final projectsProvider = context.watch<ProjectsProvider>();
    final pageEditor = context.watch<PageEditorProvider>();
    final itemsProvider = context.watch<ItemsProvider>();
    final currentPage = pageEditor.currentPage;

    if (!_autoSelected &&
        _selectedProjectId == null &&
        projectsProvider.projects.isNotEmpty) {
      _autoSelected = true;
      final projectId = projectsProvider.projects.first.id;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        setState(() {
          _selectedProjectId = projectId;
        });
        context.read<PageEditorProvider>().loadPageForProject(projectId);
      });
    }

    return Padding(
      padding: const EdgeInsets.all(AppTheme.spacing16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: DropdownButtonFormField<String>(
                  initialValue: _selectedProjectId,
                  decoration: const InputDecoration(
                    labelText: 'Project',
                  ),
                  items: projectsProvider.projects
                      .map(
                        (project) => DropdownMenuItem(
                          value: project.id,
                          child: Text(project.title),
                        ),
                      )
                      .toList(),
                  onChanged: (value) async {
                    if (value == null) return;
                    setState(() {
                      _selectedProjectId = value;
                    });
                    await context
                        .read<PageEditorProvider>()
                        .loadPageForProject(value);
                  },
                ),
              ),
              const SizedBox(width: AppTheme.spacing12),
              IconButton(
                onPressed: _createProject,
                icon: const Icon(Icons.add_circle_outline),
                tooltip: 'New project',
              ),
            ],
          ),
          const SizedBox(height: AppTheme.spacing16),
          Row(
            children: [
              Expanded(
                child: Text(
                  'Templates',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
              ),
              OutlinedButton.icon(
                onPressed: currentPage == null ? null : _showExportSheet,
                icon: const Icon(Icons.download),
                label: const Text('Export'),
              ),
            ],
          ),
          const SizedBox(height: AppTheme.spacing12),
          SizedBox(
            height: 56,
            child: ListView(
              scrollDirection: Axis.horizontal,
              children: _PageTemplate.defaults
                  .map(
                    (template) => Padding(
                      padding: const EdgeInsets.only(right: AppTheme.spacing12),
                      child: ChoiceChip(
                        label: Text(template.name),
                        selected: _activeTemplate.name == template.name,
                        onSelected: (_) => _applyTemplate(template),
                      ),
                    ),
                  )
                  .toList(),
            ),
          ),
          const SizedBox(height: AppTheme.spacing16),
          Expanded(
            child: LayoutBuilder(
              builder: (context, constraints) {
                final canvasSize = Size(
                  constraints.maxWidth,
                  constraints.maxHeight,
                );
                final imageLookup = {
                  for (final item in itemsProvider.items)
                    item.id: (item.processedImageUrl ?? item.originalImageUrl),
                };

                return Stack(
                  children: [
                    Container(
                      key: _canvasKey,
                      decoration: BoxDecoration(
                        color: currentPage != null
                            ? _colorFromHex(currentPage.backgroundColor)
                            : _activeTemplate.backgroundColor,
                        borderRadius:
                            BorderRadius.circular(AppTheme.radiusLarge),
                        border:
                            Border.all(color: const Color(0xFFE5E7EB)),
                      ),
                      child: CustomPaint(
                        painter: _activeTemplate.painter,
                        child: const SizedBox.expand(),
                      ),
                    ),
                    if (pageEditor.isLoading)
                      const Center(child: CircularProgressIndicator()),
                    ...pageEditor.pageItems.map((item) {
                      final imageUrl = imageLookup[item.itemId];
                      final width = item.width;
                      final height = item.height;
                      return Positioned(
                        left: item.positionX,
                        top: item.positionY,
                        child: GestureDetector(
                          onScaleStart: (details) {
                            final box = _canvasKey.currentContext
                                ?.findRenderObject() as RenderBox?;
                            if (box == null) return;
                            _gestureStart =
                                box.globalToLocal(details.focalPoint);
                            _itemStart =
                                Offset(item.positionX, item.positionY);
                            _widthStart = item.width;
                            _heightStart = item.height;
                            _rotationStart = item.rotation;
                          },
                          onScaleUpdate: (details) {
                            final box = _canvasKey.currentContext
                                ?.findRenderObject() as RenderBox?;
                            if (box == null) return;
                            final focal =
                                box.globalToLocal(details.focalPoint);
                            final delta = focal - _gestureStart;
                            final nextWidth =
                                (_widthStart * details.scale).clamp(80.0, 420.0);
                            final nextHeight =
                                (_heightStart * details.scale).clamp(80.0, 420.0);
                            final nextOffset = _clampOffset(
                              _itemStart + delta,
                              canvasSize,
                              nextWidth,
                              nextHeight,
                            );

                            pageEditor.setItemTransform(
                              pageItemId: item.id,
                              positionX: nextOffset.dx,
                              positionY: nextOffset.dy,
                              width: nextWidth,
                              height: nextHeight,
                              rotation: _rotationStart + details.rotation,
                            );
                          },
                          onScaleEnd: (_) {
                            pageEditor.persistItemTransform(
                              pageItemId: item.id,
                            );
                          },
                          onLongPress: () =>
                              pageEditor.deletePageItem(item.id),
                          child: Transform(
                            alignment: Alignment.center,
                            transform: Matrix4.identity()
                              ..rotateZ(item.rotation),
                            child: Container(
                              width: width,
                              height: height,
                              decoration: BoxDecoration(
                                borderRadius:
                                    BorderRadius.circular(AppTheme.radiusMedium),
                              ),
                              child: ClipRRect(
                                borderRadius: BorderRadius.circular(
                                  AppTheme.radiusMedium,
                                ),
                                child: imageUrl == null
                                    ? Container(
                                        color: const Color(0xFFE5E7EB),
                                        child: const Center(
                                          child: Icon(Icons.image_not_supported),
                                        ),
                                      )
                                    : Image.network(
                                        imageUrl,
                                        fit: BoxFit.cover,
                                      ),
                              ),
                            ),
                          ),
                        ),
                      );
                    }),
                    Positioned(
                      right: AppTheme.spacing16,
                      bottom: AppTheme.spacing16,
                      child: FloatingActionButton.extended(
                        onPressed: currentPage == null
                            ? null
                            : () => _openItemPicker(canvasSize),
                        icon: const Icon(Icons.add),
                        label: const Text('Add item'),
                      ),
                    ),
                  ],
                );
              },
            ),
          ),
          const SizedBox(height: AppTheme.spacing12),
          Text(
            'Tip: drag to move, pinch to scale/rotate, long-press to remove.',
            style: Theme.of(context)
                .textTheme
                .bodySmall
                ?.copyWith(color: AppTheme.textSecondary),
          ),
        ],
      ),
    );
  }

  Offset _clampOffset(
    Offset value,
    Size canvasSize,
    double width,
    double height,
  ) {
    final maxX = canvasSize.width - width;
    final maxY = canvasSize.height - height;
    return Offset(
      value.dx.clamp(0.0, math.max(0.0, maxX)),
      value.dy.clamp(0.0, math.max(0.0, maxY)),
    );
  }

  Color _colorFromHex(String hex) {
    final sanitized = hex.replaceAll('#', '');
    if (sanitized.length != 6) {
      return _activeTemplate.backgroundColor;
    }
    final value = int.tryParse('FF$sanitized', radix: 16);
    if (value == null) {
      return _activeTemplate.backgroundColor;
    }
    return Color(value);
  }
}

class _PageTemplate {
  _PageTemplate({
    required this.name,
    required this.backgroundColor,
    required this.painter,
  });

  final String name;
  final Color backgroundColor;
  final CustomPainter? painter;

  static _PageTemplate clean() {
    return _PageTemplate(
      name: 'Clean',
      backgroundColor: const Color(0xFFF9FAFB),
      painter: null,
    );
  }

  static _PageTemplate grid() {
    return _PageTemplate(
      name: 'Grid',
      backgroundColor: const Color(0xFFFFFFFF),
      painter: _GridPainter(),
    );
  }

  static _PageTemplate split() {
    return _PageTemplate(
      name: 'Split',
      backgroundColor: const Color(0xFFF8FAFC),
      painter: _SplitPainter(),
    );
  }

  static List<_PageTemplate> get defaults => [
        clean(),
        grid(),
        split(),
      ];
}

class _ExportPreset {
  const _ExportPreset({
    required this.label,
    required this.width,
    required this.height,
    required this.format,
    required this.scale,
    required this.quality,
  });

  final String label;
  final int width;
  final int height;
  final String format;
  final double scale;
  final int quality;

  static const List<_ExportPreset> defaults = [
    _ExportPreset(
      label: 'Instagram Portrait',
      width: 1080,
      height: 1350,
      format: 'jpeg',
      scale: 2.0,
      quality: 92,
    ),
    _ExportPreset(
      label: 'Instagram Square',
      width: 1080,
      height: 1080,
      format: 'jpeg',
      scale: 2.0,
      quality: 92,
    ),
    _ExportPreset(
      label: 'Story / Reels',
      width: 1080,
      height: 1920,
      format: 'jpeg',
      scale: 2.0,
      quality: 92,
    ),
    _ExportPreset(
      label: 'Transparent PNG (Full)',
      width: 1080,
      height: 1920,
      format: 'png',
      scale: 2.0,
      quality: 100,
    ),
  ];
}

class _GridPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = const Color(0xFFEEF2FF)
      ..strokeWidth = 1;
    const step = 48.0;
    for (double x = step; x < size.width; x += step) {
      canvas.drawLine(Offset(x, 0), Offset(x, size.height), paint);
    }
    for (double y = step; y < size.height; y += step) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), paint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _SplitPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = const Color(0xFFE0E7FF)
      ..strokeWidth = 2;

    final mid = size.width / 2;
    canvas.drawLine(Offset(mid, 0), Offset(mid, size.height), paint);
    canvas.drawLine(
      Offset(0, size.height * 0.33),
      Offset(size.width, size.height * 0.33),
      paint,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
