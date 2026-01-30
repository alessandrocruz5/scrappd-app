import 'package:flutter_dotenv/flutter_dotenv.dart';

class ApiConstants {
  // Get base URL from .env or use localhost as default
  static String get baseUrl {
    try {
      return dotenv.env['API_BASE_URL'] ?? 'http://localhost:8080';
    } catch (e) {
      // If dotenv is not loaded, return default
      return 'http://localhost:8080';
    }
  }

  static const String apiVersion = '/api/v1';
  static const String healthCheck = '/health';
  static const String removeBackground = '$apiVersion/ml/process';

  // Auth
  static const String authRegister = '$apiVersion/auth/register';
  static const String authLogin = '$apiVersion/auth/login';
  static const String authRefresh = '$apiVersion/auth/refresh';
  static const String authLogout = '$apiVersion/auth/logout';
  static const String authMe = '$apiVersion/auth/me';

  // Items
  static const String items = '$apiVersion/items';
  static const String itemsUsage = '$apiVersion/items/usage';

  // Projects & Pages
  static const String projects = '$apiVersion/projects';
  static const String pages = '$apiVersion/pages';
  
  static const Duration connectionTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 120);
  static const Duration sendTimeout = Duration(seconds: 30);
  
  static const Map<String, String> defaultHeaders = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  };
  
  static const int maxFileSize = 10 * 1024 * 1024;
  static const List<String> allowedFormats = ['jpg', 'jpeg', 'png', 'webp'];
}
