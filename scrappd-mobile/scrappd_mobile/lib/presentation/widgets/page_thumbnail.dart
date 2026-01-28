import 'package:flutter/material.dart';
import '../../core/constants/theme_constants.dart';
import '../../data/models/page.dart';

class PageThumbnail extends StatelessWidget {
  final ScrapbookPage page;
  final VoidCallback onTap;
  final VoidCallback? onLongPress;

  const PageThumbnail({
    super.key,
    required this.page,
    required this.onTap,
    this.onLongPress,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      onLongPress: onLongPress,
      child: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(AppTheme.radiusMedium),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.08),
              blurRadius: 8,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Page Preview
            Expanded(
              child: ClipRRect(
                borderRadius: const BorderRadius.vertical(
                  top: Radius.circular(AppTheme.radiusMedium),
                ),
                child: _buildPagePreview(),
              ),
            ),

            // Page Info
            Padding(
              padding: const EdgeInsets.all(AppTheme.spacing12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Page number
                  Text(
                    'Page ${page.pageNumber}',
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: AppTheme.textPrimary,
                    ),
                  ),
                  if (page.title != null && page.title!.isNotEmpty) ...[
                    const SizedBox(height: AppTheme.spacing4),
                    Text(
                      page.title!,
                      style: const TextStyle(
                        fontSize: 12,
                        color: AppTheme.textSecondary,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPagePreview() {
    // Parse background color
    Color bgColor;
    try {
      final colorString = page.backgroundColor.replaceFirst('#', '');
      bgColor = Color(int.parse('FF$colorString', radix: 16));
    } catch (_) {
      bgColor = Colors.white;
    }

    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: bgColor,
        border: Border.all(
          color: Colors.grey.withValues(alpha: 0.2),
          width: 1,
        ),
      ),
      child: Stack(
        children: [
          // Background pattern or image placeholder
          if (page.backgroundImageUrl != null)
            Positioned.fill(
              child: Container(
                color: AppTheme.primaryColor.withValues(alpha: 0.05),
                child: const Icon(
                  Icons.image_outlined,
                  color: AppTheme.textHint,
                  size: 32,
                ),
              ),
            )
          else if (page.backgroundPattern != null)
            Positioned.fill(
              child: _buildPatternPreview(page.backgroundPattern!),
            ),

          // Canvas size indicator
          Positioned(
            bottom: AppTheme.spacing8,
            right: AppTheme.spacing8,
            child: Container(
              padding: const EdgeInsets.symmetric(
                horizontal: AppTheme.spacing8,
                vertical: AppTheme.spacing4,
              ),
              decoration: BoxDecoration(
                color: Colors.black.withValues(alpha: 0.5),
                borderRadius: BorderRadius.circular(AppTheme.radiusSmall),
              ),
              child: Text(
                '${page.canvasWidth}x${page.canvasHeight}',
                style: const TextStyle(
                  fontSize: 10,
                  color: Colors.white,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
          ),

          // Empty page indicator
          if (page.layoutTemplate == null ||
              (page.layoutTemplate is Map && (page.layoutTemplate as Map).isEmpty))
            Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.add_photo_alternate_outlined,
                    size: 32,
                    color: AppTheme.textHint.withValues(alpha: 0.5),
                  ),
                  const SizedBox(height: AppTheme.spacing4),
                  Text(
                    'Empty',
                    style: TextStyle(
                      fontSize: 12,
                      color: AppTheme.textHint.withValues(alpha: 0.5),
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildPatternPreview(String pattern) {
    // Simple pattern visualization
    switch (pattern) {
      case 'dots':
        return CustomPaint(
          painter: _DotPatternPainter(),
        );
      case 'grid':
        return CustomPaint(
          painter: _GridPatternPainter(),
        );
      case 'lines':
        return CustomPaint(
          painter: _LinePatternPainter(),
        );
      default:
        return const SizedBox.shrink();
    }
  }
}

class _DotPatternPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.grey.withValues(alpha: 0.2)
      ..style = PaintingStyle.fill;

    const spacing = 20.0;
    const radius = 2.0;

    for (double x = spacing; x < size.width; x += spacing) {
      for (double y = spacing; y < size.height; y += spacing) {
        canvas.drawCircle(Offset(x, y), radius, paint);
      }
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _GridPatternPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.grey.withValues(alpha: 0.15)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1;

    const spacing = 20.0;

    // Vertical lines
    for (double x = spacing; x < size.width; x += spacing) {
      canvas.drawLine(Offset(x, 0), Offset(x, size.height), paint);
    }

    // Horizontal lines
    for (double y = spacing; y < size.height; y += spacing) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), paint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _LinePatternPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.grey.withValues(alpha: 0.15)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1;

    const spacing = 20.0;

    // Horizontal lines only
    for (double y = spacing; y < size.height; y += spacing) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), paint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
