class FriendlyException implements Exception {
  const FriendlyException(
    this.message, {
    this.statusCode,
    this.isRetryable = false,
  });

  final String message;
  final int? statusCode;
  final bool isRetryable;

  @override
  String toString() => message;
}
