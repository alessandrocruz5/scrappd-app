class ApiResponse<T> {
  final bool success;
  final T? data;
  final ApiError? error;
  final ApiMeta? meta;

  ApiResponse({
    required this.success,
    this.data,
    this.error,
    this.meta,
  });

  factory ApiResponse.fromJson(
    Map<String, dynamic> json,
    T Function(dynamic)? fromJsonT,
  ) {
    return ApiResponse(
      success: json['success'] ?? false,
      data: json['data'] != null && fromJsonT != null
          ? fromJsonT(json['data'])
          : null,
      error: json['error'] != null ? ApiError.fromJson(json['error']) : null,
      meta: json['meta'] != null ? ApiMeta.fromJson(json['meta']) : null,
    );
  }
}

class ApiError {
  final String code;
  final String message;

  ApiError({required this.code, required this.message});

  factory ApiError.fromJson(Map<String, dynamic> json) {
    return ApiError(
      code: json['code'] ?? 'UNKNOWN_ERROR',
      message: json['message'] ?? 'An unknown error occurred',
    );
  }
}

class ApiMeta {
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  ApiMeta({
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory ApiMeta.fromJson(Map<String, dynamic> json) {
    return ApiMeta(
      page: json['page'] ?? 1,
      perPage: json['per_page'] ?? 20,
      total: json['total'] ?? 0,
      totalPages: json['total_pages'] ?? 1,
    );
  }
}
