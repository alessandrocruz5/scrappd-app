import '../../domain/entities/page.dart';

class PageModel extends Page {
  PageModel({
    required super.id,
    required super.projectId,
    required super.pageNumber,
    required super.title,
    required super.canvasWidth,
    required super.canvasHeight,
    required super.backgroundColor,
    required super.backgroundImageUrl,
    required super.backgroundPattern,
  });

  factory PageModel.fromJson(Map<String, dynamic> json) {
    return PageModel(
      id: json['id'] as String,
      projectId: json['project_id'] as String,
      pageNumber: json['page_number'] as int? ?? 1,
      title: json['title'] as String?,
      canvasWidth: json['canvas_width'] as int? ?? 1080,
      canvasHeight: json['canvas_height'] as int? ?? 1920,
      backgroundColor: json['background_color'] as String? ?? '#FFFFFF',
      backgroundImageUrl: json['background_image_url'] as String?,
      backgroundPattern: json['background_pattern'] as String?,
    );
  }
}
