import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../../core/constants/theme_constants.dart';
import '../../data/models/page_item.dart';

class CanvasItem extends StatelessWidget {
  final PageItem pageItem;
  final bool isSelected;
  final double canvasScale;
  final VoidCallback onTap;
  final Function(double dx, double dy) onDragUpdate;
  final VoidCallback onDragEnd;

  const CanvasItem({
    super.key,
    required this.pageItem,
    required this.isSelected,
    required this.canvasScale,
    required this.onTap,
    required this.onDragUpdate,
    required this.onDragEnd,
  });

  @override
  Widget build(BuildContext context) {
    final imageUrl = pageItem.item?.displayImageUrl;

    return Positioned(
      left: pageItem.positionX,
      top: pageItem.positionY,
      child: GestureDetector(
        onTap: onTap,
        onPanUpdate: (details) {
          onDragUpdate(
            details.delta.dx / canvasScale,
            details.delta.dy / canvasScale,
          );
        },
        onPanEnd: (_) => onDragEnd(),
        child: Transform.rotate(
          angle: pageItem.rotation * math.pi / 180,
          child: Opacity(
            opacity: pageItem.opacity,
            child: Container(
              width: pageItem.width,
              height: pageItem.height,
              decoration: BoxDecoration(
                border: isSelected
                    ? Border.all(
                        color: AppTheme.primaryColor,
                        width: 2 / canvasScale,
                      )
                    : null,
              ),
              child: Stack(
                fit: StackFit.expand,
                children: [
                  // Image content
                  if (imageUrl != null && imageUrl.isNotEmpty)
                    CachedNetworkImage(
                      imageUrl: imageUrl,
                      fit: BoxFit.contain,
                      placeholder: (context, url) => Container(
                        color: Colors.grey.shade200,
                        child: const Center(
                          child: CircularProgressIndicator(strokeWidth: 2),
                        ),
                      ),
                      errorWidget: (context, url, error) => Container(
                        color: Colors.grey.shade200,
                        child: const Icon(
                          Icons.broken_image_outlined,
                          color: Colors.grey,
                        ),
                      ),
                    )
                  else
                    Container(
                      color: Colors.grey.shade200,
                      child: const Icon(
                        Icons.image_outlined,
                        color: Colors.grey,
                      ),
                    ),

                  // Selection handles
                  if (isSelected) _buildSelectionHandles(),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSelectionHandles() {
    const handleSize = 12.0;
    const handleColor = AppTheme.primaryColor;

    return Stack(
      children: [
        // Corner handles
        _buildHandle(Alignment.topLeft, handleSize, handleColor),
        _buildHandle(Alignment.topRight, handleSize, handleColor),
        _buildHandle(Alignment.bottomLeft, handleSize, handleColor),
        _buildHandle(Alignment.bottomRight, handleSize, handleColor),
        // Edge handles
        _buildHandle(Alignment.topCenter, handleSize, handleColor),
        _buildHandle(Alignment.bottomCenter, handleSize, handleColor),
        _buildHandle(Alignment.centerLeft, handleSize, handleColor),
        _buildHandle(Alignment.centerRight, handleSize, handleColor),
      ],
    );
  }

  Widget _buildHandle(Alignment alignment, double size, Color color) {
    return Align(
      alignment: alignment,
      child: Container(
        width: size / canvasScale,
        height: size / canvasScale,
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border.all(color: color, width: 2 / canvasScale),
          shape: BoxShape.circle,
        ),
      ),
    );
  }
}

class TransformableCanvasItem extends StatefulWidget {
  final PageItem pageItem;
  final bool isSelected;
  final double canvasScale;
  final VoidCallback onTap;
  final Function(double x, double y) onPositionChanged;
  final Function(double width, double height) onSizeChanged;
  final Function(double rotation) onRotationChanged;
  final VoidCallback onTransformEnd;

  const TransformableCanvasItem({
    super.key,
    required this.pageItem,
    required this.isSelected,
    required this.canvasScale,
    required this.onTap,
    required this.onPositionChanged,
    required this.onSizeChanged,
    required this.onRotationChanged,
    required this.onTransformEnd,
  });

  @override
  State<TransformableCanvasItem> createState() => _TransformableCanvasItemState();
}

class _TransformableCanvasItemState extends State<TransformableCanvasItem> {
  late double _x;
  late double _y;
  late double _width;
  late double _height;
  late double _rotation;
  bool _isDragging = false;
  String? _activeHandle;

  @override
  void initState() {
    super.initState();
    _syncFromPageItem();
  }

  @override
  void didUpdateWidget(TransformableCanvasItem oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!_isDragging) {
      _syncFromPageItem();
    }
  }

  void _syncFromPageItem() {
    _x = widget.pageItem.positionX;
    _y = widget.pageItem.positionY;
    _width = widget.pageItem.width;
    _height = widget.pageItem.height;
    _rotation = widget.pageItem.rotation;
  }

  @override
  Widget build(BuildContext context) {
    final imageUrl = widget.pageItem.item?.displayImageUrl;

    return Positioned(
      left: _x,
      top: _y,
      child: Transform.rotate(
        angle: _rotation * math.pi / 180,
        child: Opacity(
          opacity: widget.pageItem.opacity,
          child: GestureDetector(
            onTap: widget.onTap,
            onPanStart: (_) {
              _isDragging = true;
              _activeHandle = null;
            },
            onPanUpdate: (details) {
              setState(() {
                _x += details.delta.dx / widget.canvasScale;
                _y += details.delta.dy / widget.canvasScale;
              });
            },
            onPanEnd: (_) {
              _isDragging = false;
              widget.onPositionChanged(_x, _y);
              widget.onTransformEnd();
            },
            child: Container(
              width: _width,
              height: _height,
              decoration: BoxDecoration(
                border: widget.isSelected
                    ? Border.all(
                        color: AppTheme.primaryColor,
                        width: 2 / widget.canvasScale,
                      )
                    : null,
              ),
              child: Stack(
                fit: StackFit.expand,
                children: [
                  // Image content
                  if (imageUrl != null && imageUrl.isNotEmpty)
                    CachedNetworkImage(
                      imageUrl: imageUrl,
                      fit: BoxFit.contain,
                      placeholder: (context, url) => Container(
                        color: Colors.grey.shade200,
                        child: const Center(
                          child: CircularProgressIndicator(strokeWidth: 2),
                        ),
                      ),
                      errorWidget: (context, url, error) => Container(
                        color: Colors.grey.shade200,
                        child: const Icon(
                          Icons.broken_image_outlined,
                          color: Colors.grey,
                        ),
                      ),
                    )
                  else
                    Container(
                      color: Colors.grey.shade200,
                      child: const Icon(
                        Icons.image_outlined,
                        color: Colors.grey,
                      ),
                    ),

                  // Resize handles (only when selected)
                  if (widget.isSelected) ...[
                    _buildResizeHandle(Alignment.topLeft, 'tl'),
                    _buildResizeHandle(Alignment.topRight, 'tr'),
                    _buildResizeHandle(Alignment.bottomLeft, 'bl'),
                    _buildResizeHandle(Alignment.bottomRight, 'br'),
                    _buildEdgeHandle(Alignment.topCenter, 't'),
                    _buildEdgeHandle(Alignment.bottomCenter, 'b'),
                    _buildEdgeHandle(Alignment.centerLeft, 'l'),
                    _buildEdgeHandle(Alignment.centerRight, 'r'),
                  ],
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildResizeHandle(Alignment alignment, String handleId) {
    const handleSize = 14.0;

    return Align(
      alignment: alignment,
      child: GestureDetector(
        onPanStart: (_) {
          _isDragging = true;
          _activeHandle = handleId;
        },
        onPanUpdate: (details) {
          final dx = details.delta.dx / widget.canvasScale;
          final dy = details.delta.dy / widget.canvasScale;

          setState(() {
            switch (handleId) {
              case 'tl':
                _x += dx;
                _y += dy;
                _width -= dx;
                _height -= dy;
                break;
              case 'tr':
                _y += dy;
                _width += dx;
                _height -= dy;
                break;
              case 'bl':
                _x += dx;
                _width -= dx;
                _height += dy;
                break;
              case 'br':
                _width += dx;
                _height += dy;
                break;
            }

            // Minimum size constraint
            if (_width < 20) _width = 20;
            if (_height < 20) _height = 20;
          });
        },
        onPanEnd: (_) {
          _isDragging = false;
          _activeHandle = null;
          widget.onPositionChanged(_x, _y);
          widget.onSizeChanged(_width, _height);
          widget.onTransformEnd();
        },
        child: Container(
          width: handleSize / widget.canvasScale,
          height: handleSize / widget.canvasScale,
          decoration: BoxDecoration(
            color: Colors.white,
            border: Border.all(
              color: AppTheme.primaryColor,
              width: 2 / widget.canvasScale,
            ),
            borderRadius: BorderRadius.circular(2 / widget.canvasScale),
          ),
        ),
      ),
    );
  }

  Widget _buildEdgeHandle(Alignment alignment, String handleId) {
    const handleSize = 10.0;
    final isHorizontal = handleId == 't' || handleId == 'b';

    return Align(
      alignment: alignment,
      child: GestureDetector(
        onPanStart: (_) {
          _isDragging = true;
          _activeHandle = handleId;
        },
        onPanUpdate: (details) {
          final dx = details.delta.dx / widget.canvasScale;
          final dy = details.delta.dy / widget.canvasScale;

          setState(() {
            switch (handleId) {
              case 't':
                _y += dy;
                _height -= dy;
                break;
              case 'b':
                _height += dy;
                break;
              case 'l':
                _x += dx;
                _width -= dx;
                break;
              case 'r':
                _width += dx;
                break;
            }

            if (_width < 20) _width = 20;
            if (_height < 20) _height = 20;
          });
        },
        onPanEnd: (_) {
          _isDragging = false;
          _activeHandle = null;
          widget.onPositionChanged(_x, _y);
          widget.onSizeChanged(_width, _height);
          widget.onTransformEnd();
        },
        child: Container(
          width: isHorizontal ? handleSize * 2 / widget.canvasScale : handleSize / widget.canvasScale,
          height: isHorizontal ? handleSize / widget.canvasScale : handleSize * 2 / widget.canvasScale,
          decoration: BoxDecoration(
            color: Colors.white,
            border: Border.all(
              color: AppTheme.primaryColor,
              width: 2 / widget.canvasScale,
            ),
            borderRadius: BorderRadius.circular(2 / widget.canvasScale),
          ),
        ),
      ),
    );
  }
}
