import '../config/environment.dart';

/// API Constants for Scrapp'd
/// 
/// Uses EnvironmentConfig to automatically switch between
/// development and production URLs.
class ApiConstants {
  // Base URL from environment config
  static String get baseUrl => EnvironmentConfig.apiBaseUrl;
  
  // API versioning
  static const String apiVersion = '/api/v1';
  
  // Health check endpoints
  static const String healthCheck = '/health';
  static const String healthDeep = '/health/deep';
  
  // Auth endpoints
  static const String authRegister = '/auth/register';
  static const String authLogin = '/auth/login';
  static const String authRefresh = '/auth/refresh';
  static const String authLogout = '/auth/logout';
  
  // User endpoints
  static String get me => '$apiVersion/me';
  static String get authMe => '$apiVersion/me';
  static String get usage => '$apiVersion/usage';

  // Projects endpoints
  static String get projects => '$apiVersion/projects';
  static String projectById(String id) => '$apiVersion/projects/$id';

  // ML Processing endpoints
  static String get removeBackground => '$apiVersion/ml/process';
  static String get removeBackgroundUpload => '$apiVersion/remove-background/upload';
  
  // Items endpoints
  static String get items => '$apiVersion/items';
  static String itemById(String id) => '$apiVersion/items/$id';
  
  // Pages endpoints
  static String get pages => '$apiVersion/pages';
  static String pageById(String id) => '$apiVersion/pages/$id';
  static String pageItems(String pageId) => '$apiVersion/pages/$pageId/items';
  
  // Timeouts from environment
  static Duration get connectionTimeout => EnvironmentConfig.connectionTimeout;
  static Duration get receiveTimeout => EnvironmentConfig.receiveTimeout;
  static const Duration sendTimeout = Duration(seconds: 60);
  
  // Default headers
  static const Map<String, String> defaultHeaders = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  };
  
  // Image constraints
  static const int maxFileSize = 10 * 1024 * 1024; // 10MB
  static const List<String> allowedFormats = ['jpg', 'jpeg', 'png', 'webp'];
}
