class Project {
  final String id;
  final String userId;
  final String title;
  final String? description;
  final String? coverImageUrl;
  final String visibility;
  final bool isTemplate;
  final double? templatePrice;
  final int viewCount;
  final int likeCount;
  final int forkCount;
  final DateTime createdAt;
  final DateTime updatedAt;
  final DateTime? publishedAt;

  Project({
    required this.id,
    required this.userId,
    required this.title,
    this.description,
    this.coverImageUrl,
    required this.visibility,
    required this.isTemplate,
    this.templatePrice,
    required this.viewCount,
    required this.likeCount,
    required this.forkCount,
    required this.createdAt,
    required this.updatedAt,
    this.publishedAt,
  });

  factory Project.fromJson(Map<String, dynamic> json) {
    return Project(
      id: json['id'] ?? '',
      userId: json['user_id'] ?? '',
      title: json['title'] ?? '',
      description: json['description'],
      coverImageUrl: json['cover_image_url'],
      visibility: json['visibility'] ?? 'private',
      isTemplate: json['is_template'] ?? false,
      templatePrice: json['template_price']?.toDouble(),
      viewCount: json['view_count'] ?? 0,
      likeCount: json['like_count'] ?? 0,
      forkCount: json['fork_count'] ?? 0,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'])
          : DateTime.now(),
      publishedAt: json['published_at'] != null
          ? DateTime.parse(json['published_at'])
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'title': title,
      'description': description,
      'cover_image_url': coverImageUrl,
      'visibility': visibility,
      'is_template': isTemplate,
      'template_price': templatePrice,
      'view_count': viewCount,
      'like_count': likeCount,
      'fork_count': forkCount,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
      'published_at': publishedAt?.toIso8601String(),
    };
  }

  Project copyWith({
    String? id,
    String? userId,
    String? title,
    String? description,
    String? coverImageUrl,
    String? visibility,
    bool? isTemplate,
    double? templatePrice,
    int? viewCount,
    int? likeCount,
    int? forkCount,
    DateTime? createdAt,
    DateTime? updatedAt,
    DateTime? publishedAt,
  }) {
    return Project(
      id: id ?? this.id,
      userId: userId ?? this.userId,
      title: title ?? this.title,
      description: description ?? this.description,
      coverImageUrl: coverImageUrl ?? this.coverImageUrl,
      visibility: visibility ?? this.visibility,
      isTemplate: isTemplate ?? this.isTemplate,
      templatePrice: templatePrice ?? this.templatePrice,
      viewCount: viewCount ?? this.viewCount,
      likeCount: likeCount ?? this.likeCount,
      forkCount: forkCount ?? this.forkCount,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      publishedAt: publishedAt ?? this.publishedAt,
    );
  }
}

class CreateProjectRequest {
  final String title;
  final String? description;
  final String visibility;

  CreateProjectRequest({
    required this.title,
    this.description,
    this.visibility = 'private',
  });

  Map<String, dynamic> toJson() {
    return {
      'title': title,
      if (description != null) 'description': description,
      'visibility': visibility,
    };
  }
}

class UpdateProjectRequest {
  final String? title;
  final String? description;
  final String? coverImageUrl;
  final String? visibility;
  final bool? isTemplate;
  final double? templatePrice;

  UpdateProjectRequest({
    this.title,
    this.description,
    this.coverImageUrl,
    this.visibility,
    this.isTemplate,
    this.templatePrice,
  });

  Map<String, dynamic> toJson() {
    final json = <String, dynamic>{};
    if (title != null) json['title'] = title;
    if (description != null) json['description'] = description;
    if (coverImageUrl != null) json['cover_image_url'] = coverImageUrl;
    if (visibility != null) json['visibility'] = visibility;
    if (isTemplate != null) json['is_template'] = isTemplate;
    if (templatePrice != null) json['template_price'] = templatePrice;
    return json;
  }
}

enum ProjectVisibility {
  private_('private'),
  unlisted('unlisted'),
  public_('public');

  final String value;
  const ProjectVisibility(this.value);

  static ProjectVisibility fromString(String value) {
    return ProjectVisibility.values.firstWhere(
      (v) => v.value == value,
      orElse: () => ProjectVisibility.private_,
    );
  }
}
