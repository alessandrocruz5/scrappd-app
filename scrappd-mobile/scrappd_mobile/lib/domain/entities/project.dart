class Project {
  final String id;
  final String title;
  final String? description;
  final String? coverImageUrl;
  final String visibility;
  final bool isTemplate;
  final DateTime createdAt;

  Project({
    required this.id,
    required this.title,
    required this.description,
    required this.coverImageUrl,
    required this.visibility,
    required this.isTemplate,
    required this.createdAt,
  });
}
