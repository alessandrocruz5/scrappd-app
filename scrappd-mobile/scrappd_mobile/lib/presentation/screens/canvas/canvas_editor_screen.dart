import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../../core/constants/theme_constants.dart';
import '../../../data/models/page.dart';
import '../../../data/models/page_item.dart';
import '../../providers/canvas_provider.dart';
import '../../widgets/canvas_item.dart';
import 'item_picker_sheet.dart';

class CanvasEditorScreen extends StatefulWidget {
  final ScrapbookPage page;

  const CanvasEditorScreen({
    super.key,
    required this.page,
  });

  @override
  State<CanvasEditorScreen> createState() => _CanvasEditorScreenState();
}

class _CanvasEditorScreenState extends State<CanvasEditorScreen> {
  final TransformationController _transformController = TransformationController();

  @override
  void initState() {
    super.initState();
    _loadPage();
  }

  void _loadPage() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<CanvasProvider>().loadPage(widget.page);
    });
  }

  @override
  void dispose() {
    _transformController.dispose();
    super.dispose();
  }

  void _showItemPicker() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => const ItemPickerSheet(),
    );
  }

  void _showLayerOptions() {
    final provider = context.read<CanvasProvider>();
    final selectedItem = provider.selectedItem;
    if (selectedItem == null) return;

    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.flip_to_front),
              title: const Text('Bring to Front'),
              onTap: () {
                provider.bringToFront(selectedItem.id);
                Navigator.pop(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.flip_to_back),
              title: const Text('Send to Back'),
              onTap: () {
                provider.sendToBack(selectedItem.id);
                Navigator.pop(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.arrow_upward),
              title: const Text('Bring Forward'),
              onTap: () {
                provider.bringForward(selectedItem.id);
                Navigator.pop(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.arrow_downward),
              title: const Text('Send Backward'),
              onTap: () {
                provider.sendBackward(selectedItem.id);
                Navigator.pop(context);
              },
            ),
            const Divider(),
            ListTile(
              leading: const Icon(Icons.copy),
              title: const Text('Duplicate'),
              onTap: () {
                provider.duplicateSelected();
                Navigator.pop(context);
              },
            ),
            ListTile(
              leading: const Icon(Icons.delete_outline, color: AppTheme.errorColor),
              title: const Text('Delete', style: TextStyle(color: AppTheme.errorColor)),
              onTap: () {
                provider.deleteSelected();
                Navigator.pop(context);
              },
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey.shade900,
      appBar: _buildAppBar(),
      body: Column(
        children: [
          // Canvas area
          Expanded(
            child: Consumer<CanvasProvider>(
              builder: (context, provider, _) {
                if (provider.state == CanvasState.loading) {
                  return const Center(
                    child: CircularProgressIndicator(color: Colors.white),
                  );
                }

                if (provider.state == CanvasState.error) {
                  return Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(Icons.error_outline,
                            color: AppTheme.errorColor, size: 48),
                        const SizedBox(height: AppTheme.spacing16),
                        Text(
                          provider.errorMessage ?? 'Failed to load page',
                          style: const TextStyle(color: Colors.white),
                        ),
                        const SizedBox(height: AppTheme.spacing16),
                        ElevatedButton(
                          onPressed: _loadPage,
                          child: const Text('Retry'),
                        ),
                      ],
                    ),
                  );
                }

                return _buildCanvas(provider);
              },
            ),
          ),

          // Bottom toolbar
          _buildBottomToolbar(),
        ],
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      backgroundColor: Colors.grey.shade900,
      foregroundColor: Colors.white,
      title: Text(widget.page.title ?? 'Page ${widget.page.pageNumber}'),
      actions: [
        // Undo
        Consumer<CanvasProvider>(
          builder: (context, provider, _) {
            return IconButton(
              icon: const Icon(Icons.undo),
              onPressed: provider.canUndo ? provider.undo : null,
              color: provider.canUndo ? Colors.white : Colors.grey,
            );
          },
        ),
        // Redo
        Consumer<CanvasProvider>(
          builder: (context, provider, _) {
            return IconButton(
              icon: const Icon(Icons.redo),
              onPressed: provider.canRedo ? provider.redo : null,
              color: provider.canRedo ? Colors.white : Colors.grey,
            );
          },
        ),
        // Zoom controls
        PopupMenuButton<String>(
          icon: const Icon(Icons.zoom_in, color: Colors.white),
          color: Colors.grey.shade800,
          onSelected: (value) {
            final provider = context.read<CanvasProvider>();
            switch (value) {
              case 'zoom_in':
                provider.zoomIn();
                break;
              case 'zoom_out':
                provider.zoomOut();
                break;
              case 'reset':
                provider.resetView();
                _transformController.value = Matrix4.identity();
                break;
            }
          },
          itemBuilder: (context) => [
            const PopupMenuItem(
              value: 'zoom_in',
              child: Row(
                children: [
                  Icon(Icons.zoom_in, color: Colors.white),
                  SizedBox(width: 8),
                  Text('Zoom In', style: TextStyle(color: Colors.white)),
                ],
              ),
            ),
            const PopupMenuItem(
              value: 'zoom_out',
              child: Row(
                children: [
                  Icon(Icons.zoom_out, color: Colors.white),
                  SizedBox(width: 8),
                  Text('Zoom Out', style: TextStyle(color: Colors.white)),
                ],
              ),
            ),
            const PopupMenuItem(
              value: 'reset',
              child: Row(
                children: [
                  Icon(Icons.fit_screen, color: Colors.white),
                  SizedBox(width: 8),
                  Text('Reset View', style: TextStyle(color: Colors.white)),
                ],
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildCanvas(CanvasProvider provider) {
    // Parse background color
    Color bgColor;
    try {
      final colorString = widget.page.backgroundColor.replaceFirst('#', '');
      bgColor = Color(int.parse('FF$colorString', radix: 16));
    } catch (_) {
      bgColor = Colors.white;
    }

    return GestureDetector(
      onTap: () => provider.clearSelection(),
      child: InteractiveViewer(
        transformationController: _transformController,
        boundaryMargin: const EdgeInsets.all(100),
        minScale: 0.1,
        maxScale: 5.0,
        child: Center(
          child: Container(
            width: widget.page.canvasWidth.toDouble(),
            height: widget.page.canvasHeight.toDouble(),
            decoration: BoxDecoration(
              color: bgColor,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.3),
                  blurRadius: 20,
                  spreadRadius: 5,
                ),
              ],
            ),
            child: Stack(
              clipBehavior: Clip.none,
              children: [
                // Background pattern
                if (widget.page.backgroundPattern != null)
                  Positioned.fill(
                    child: _buildBackgroundPattern(widget.page.backgroundPattern!),
                  ),

                // Page items
                ...provider.items.map((item) {
                  final scale = _transformController.value.getMaxScaleOnAxis();
                  return TransformableCanvasItem(
                    key: ValueKey(item.id),
                    pageItem: item,
                    isSelected: provider.selectedItem?.id == item.id,
                    canvasScale: scale,
                    onTap: () => provider.selectItem(item),
                    onPositionChanged: (x, y) {
                      // Position will be updated on transform end
                    },
                    onSizeChanged: (w, h) {
                      // Size will be updated on transform end
                    },
                    onRotationChanged: (r) {
                      // Rotation will be updated on transform end
                    },
                    onTransformEnd: () {
                      // Get current values and update
                      // This is handled by the TransformableCanvasItem internally
                    },
                  );
                }),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildBackgroundPattern(String pattern) {
    switch (pattern) {
      case 'dots':
        return CustomPaint(
          painter: _DotPatternPainter(
            color: Colors.grey.withValues(alpha: 0.2),
            spacing: 20,
          ),
        );
      case 'grid':
        return CustomPaint(
          painter: _GridPatternPainter(
            color: Colors.grey.withValues(alpha: 0.15),
            spacing: 20,
          ),
        );
      case 'lines':
        return CustomPaint(
          painter: _LinePatternPainter(
            color: Colors.grey.withValues(alpha: 0.15),
            spacing: 20,
          ),
        );
      default:
        return const SizedBox.shrink();
    }
  }

  Widget _buildBottomToolbar() {
    return Consumer<CanvasProvider>(
      builder: (context, provider, _) {
        return Container(
          padding: const EdgeInsets.symmetric(
            horizontal: AppTheme.spacing16,
            vertical: AppTheme.spacing8,
          ),
          decoration: BoxDecoration(
            color: Colors.grey.shade900,
            border: Border(
              top: BorderSide(color: Colors.grey.shade800),
            ),
          ),
          child: SafeArea(
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                // Add item
                _buildToolbarButton(
                  icon: Icons.add_photo_alternate_outlined,
                  label: 'Add',
                  onTap: _showItemPicker,
                ),

                // Layers (only when item selected)
                _buildToolbarButton(
                  icon: Icons.layers_outlined,
                  label: 'Layers',
                  onTap: provider.selectedItem != null ? _showLayerOptions : null,
                  isEnabled: provider.selectedItem != null,
                ),

                // Delete (only when item selected)
                _buildToolbarButton(
                  icon: Icons.delete_outline,
                  label: 'Delete',
                  onTap: provider.selectedItem != null
                      ? () => provider.deleteSelected()
                      : null,
                  isEnabled: provider.selectedItem != null,
                  color: provider.selectedItem != null ? AppTheme.errorColor : null,
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildToolbarButton({
    required IconData icon,
    required String label,
    VoidCallback? onTap,
    bool isEnabled = true,
    Color? color,
  }) {
    final effectiveColor = isEnabled
        ? (color ?? Colors.white)
        : Colors.grey.shade600;

    return GestureDetector(
      onTap: isEnabled ? onTap : null,
      child: Padding(
        padding: const EdgeInsets.symmetric(
          horizontal: AppTheme.spacing12,
          vertical: AppTheme.spacing8,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, color: effectiveColor, size: 24),
            const SizedBox(height: AppTheme.spacing4),
            Text(
              label,
              style: TextStyle(
                color: effectiveColor,
                fontSize: 12,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// Pattern painters
class _DotPatternPainter extends CustomPainter {
  final Color color;
  final double spacing;

  _DotPatternPainter({required this.color, required this.spacing});

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = color
      ..style = PaintingStyle.fill;

    for (double x = spacing; x < size.width; x += spacing) {
      for (double y = spacing; y < size.height; y += spacing) {
        canvas.drawCircle(Offset(x, y), 2, paint);
      }
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _GridPatternPainter extends CustomPainter {
  final Color color;
  final double spacing;

  _GridPatternPainter({required this.color, required this.spacing});

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = color
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1;

    for (double x = spacing; x < size.width; x += spacing) {
      canvas.drawLine(Offset(x, 0), Offset(x, size.height), paint);
    }

    for (double y = spacing; y < size.height; y += spacing) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), paint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _LinePatternPainter extends CustomPainter {
  final Color color;
  final double spacing;

  _LinePatternPainter({required this.color, required this.spacing});

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = color
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1;

    for (double y = spacing; y < size.height; y += spacing) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), paint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
