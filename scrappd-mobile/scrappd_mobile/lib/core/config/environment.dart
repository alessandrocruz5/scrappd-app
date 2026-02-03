import 'package:flutter/foundation.dart';

enum Environment { development, staging, production }

class EnvironmentConfig {
  static Environment _current = Environment.development;
  
  static Environment get current => _current;
  
  /// Initialize environment based on build mode
  static void initialize() {
    // In release mode, always use production
    // In debug mode, use development
    const bool isRelease = bool.fromEnvironment('dart.vm.product');
    
    // Check for explicit override via --dart-define
    const String envOverride = String.fromEnvironment(
      'ENVIRONMENT',
      defaultValue: '',
    );
    
    if (envOverride.isNotEmpty) {
      _current = _parseEnvironment(envOverride);
    } else {
      _current = isRelease ? Environment.production : Environment.development;
    }
  }
  
  static Environment _parseEnvironment(String value) {
    switch (value.toLowerCase()) {
      case 'production':
      case 'prod':
        return Environment.production;
      case 'staging':
      case 'stage':
        return Environment.staging;
      default:
        return Environment.development;
    }
  }
  
  /// API base URL based on current environment
  static String get apiBaseUrl {
    const String apiBaseUrlOverride = String.fromEnvironment(
      'API_BASE_URL',
      defaultValue: '',
    );
    if (apiBaseUrlOverride.isNotEmpty) {
      return apiBaseUrlOverride;
    }

    switch (_current) {
      case Environment.production:
        return 'https://scrappd-api-j6bicsikba-as.a.run.app';
      case Environment.staging:
        return 'https://scrappd-api-staging-as.a.run.app';
      case Environment.development:
        if (kIsWeb) {
          return 'http://localhost:8080';
        }

        // Android emulator uses 10.0.2.2 to reach host localhost.
        // iOS simulator and desktop can use localhost directly.
        // Physical devices should pass --dart-define=API_BASE_URL=http://<host-ip>:8080.
        switch (defaultTargetPlatform) {
          case TargetPlatform.android:
            return 'http://10.0.2.2:8080';
          default:
            return 'http://localhost:8080';
        }
    }
  }
  
  /// Connection timeout based on environment
  static Duration get connectionTimeout {
    switch (_current) {
      case Environment.production:
        // Production has cold starts, need longer timeout
        return const Duration(seconds: 90);
      case Environment.staging:
        return const Duration(seconds: 60);
      case Environment.development:
        return const Duration(seconds: 30);
    }
  }
  
  /// Receive timeout (for ML processing)
  static Duration get receiveTimeout {
    // ML processing can take 15-20 seconds, plus cold start
    return const Duration(seconds: 180);
  }
  
  /// Whether to show debug info in UI
  static bool get showDebugBanner => _current == Environment.development;
  
  /// Whether to enable verbose logging
  static bool get verboseLogging => _current != Environment.production;
}
