class Page {
  final String id;
  final String projectId;
  final int pageNumber;
  final String? title;
  final int canvasWidth;
  final int canvasHeight;
  final String backgroundColor;
  final String? backgroundImageUrl;
  final String? backgroundPattern;

  Page({
    required this.id,
    required this.projectId,
    required this.pageNumber,
    required this.title,
    required this.canvasWidth,
    required this.canvasHeight,
    required this.backgroundColor,
    required this.backgroundImageUrl,
    required this.backgroundPattern,
  });
}
