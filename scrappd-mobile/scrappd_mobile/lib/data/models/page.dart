class ScrapbookPage {
  final String id;
  final String projectId;
  final int pageNumber;
  final String? title;
  final int canvasWidth;
  final int canvasHeight;
  final String backgroundColor;
  final String? backgroundImageUrl;
  final String? backgroundPattern;
  final Map<String, dynamic>? layoutTemplate;
  final DateTime createdAt;
  final DateTime updatedAt;

  ScrapbookPage({
    required this.id,
    required this.projectId,
    required this.pageNumber,
    this.title,
    required this.canvasWidth,
    required this.canvasHeight,
    required this.backgroundColor,
    this.backgroundImageUrl,
    this.backgroundPattern,
    this.layoutTemplate,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ScrapbookPage.fromJson(Map<String, dynamic> json) {
    return ScrapbookPage(
      id: json['id'] ?? '',
      projectId: json['project_id'] ?? '',
      pageNumber: json['page_number'] ?? 1,
      title: json['title'],
      canvasWidth: json['canvas_width'] ?? 1080,
      canvasHeight: json['canvas_height'] ?? 1920,
      backgroundColor: json['background_color'] ?? '#FFFFFF',
      backgroundImageUrl: json['background_image_url'],
      backgroundPattern: json['background_pattern'],
      layoutTemplate: json['layout_template'] is Map
          ? Map<String, dynamic>.from(json['layout_template'])
          : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'])
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'project_id': projectId,
      'page_number': pageNumber,
      'title': title,
      'canvas_width': canvasWidth,
      'canvas_height': canvasHeight,
      'background_color': backgroundColor,
      'background_image_url': backgroundImageUrl,
      'background_pattern': backgroundPattern,
      'layout_template': layoutTemplate,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  ScrapbookPage copyWith({
    String? id,
    String? projectId,
    int? pageNumber,
    String? title,
    int? canvasWidth,
    int? canvasHeight,
    String? backgroundColor,
    String? backgroundImageUrl,
    String? backgroundPattern,
    Map<String, dynamic>? layoutTemplate,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return ScrapbookPage(
      id: id ?? this.id,
      projectId: projectId ?? this.projectId,
      pageNumber: pageNumber ?? this.pageNumber,
      title: title ?? this.title,
      canvasWidth: canvasWidth ?? this.canvasWidth,
      canvasHeight: canvasHeight ?? this.canvasHeight,
      backgroundColor: backgroundColor ?? this.backgroundColor,
      backgroundImageUrl: backgroundImageUrl ?? this.backgroundImageUrl,
      backgroundPattern: backgroundPattern ?? this.backgroundPattern,
      layoutTemplate: layoutTemplate ?? this.layoutTemplate,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}

class CreatePageRequest {
  final String projectId;
  final int? pageNumber;
  final String? title;
  final int canvasWidth;
  final int canvasHeight;
  final String backgroundColor;

  CreatePageRequest({
    required this.projectId,
    this.pageNumber,
    this.title,
    this.canvasWidth = 1080,
    this.canvasHeight = 1920,
    this.backgroundColor = '#FFFFFF',
  });

  Map<String, dynamic> toJson() {
    return {
      'project_id': projectId,
      if (pageNumber != null) 'page_number': pageNumber,
      if (title != null) 'title': title,
      'canvas_width': canvasWidth,
      'canvas_height': canvasHeight,
      'background_color': backgroundColor,
    };
  }
}

class UpdatePageRequest {
  final int? pageNumber;
  final String? title;
  final int? canvasWidth;
  final int? canvasHeight;
  final String? backgroundColor;
  final String? backgroundImageUrl;
  final String? backgroundPattern;
  final Map<String, dynamic>? layoutTemplate;

  UpdatePageRequest({
    this.pageNumber,
    this.title,
    this.canvasWidth,
    this.canvasHeight,
    this.backgroundColor,
    this.backgroundImageUrl,
    this.backgroundPattern,
    this.layoutTemplate,
  });

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
    if (pageNumber != null) json['page_number'] = pageNumber;
    if (title != null) json['title'] = title;
    if (canvasWidth != null) json['canvas_width'] = canvasWidth;
    if (canvasHeight != null) json['canvas_height'] = canvasHeight;
    if (backgroundColor != null) json['background_color'] = backgroundColor;
    if (backgroundImageUrl != null) {
      json['background_image_url'] = backgroundImageUrl;
    }
    if (backgroundPattern != null) json['background_pattern'] = backgroundPattern;
    if (layoutTemplate != null) json['layout_template'] = layoutTemplate;
    return json;
  }
}
