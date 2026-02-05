import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:scrappd_mobile/core/config/environment.dart';

import 'core/constants/api_constants.dart';
import 'core/constants/theme_constants.dart';
import 'core/network/api_client.dart';
import 'core/storage/token_storage.dart';
import 'data/datasources/auth_remote_datasource.dart';
import 'data/datasources/items_remote_datasource.dart';
import 'data/datasources/page_items_remote_datasource.dart';
import 'data/datasources/pages_remote_datasource.dart';
import 'data/datasources/projects_remote_datasource.dart';
import 'data/repositories/auth_repository_impl.dart';
import 'data/repositories/item_repository_impl.dart';
import 'data/repositories/page_item_repository_impl.dart';
import 'data/repositories/page_repository_impl.dart';
import 'data/repositories/project_repository_impl.dart';
import 'presentation/providers/auth_provider.dart';
import 'presentation/providers/items_provider.dart';
import 'presentation/providers/page_editor_provider.dart';
import 'presentation/providers/projects_provider.dart';
import 'presentation/screens/shell/root_screen.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  // Initialize environment configuration
  EnvironmentConfig.initialize();
  
  // Log environment info (only in debug)
  if (EnvironmentConfig.verboseLogging) {
    debugPrint('🌍 Environment: ${EnvironmentConfig.current.name}');
    debugPrint('🌐 API URL: ${ApiConstants.baseUrl}');
  }
  
  // Set preferred orientations
  await SystemChrome.setPreferredOrientations([
    DeviceOrientation.portraitUp,
    DeviceOrientation.portraitDown,
  ]);
  
  // Set system UI overlay style
  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
      systemNavigationBarColor: Colors.white,
      systemNavigationBarIconBrightness: Brightness.dark,
    ),
  );
  
  runApp(const ScrappdApp());
}

class ScrappdApp extends StatelessWidget {
  const ScrappdApp({super.key});

  runApp(
    MultiProvider(
      providers: [
        // API Service
        Provider<ApiService>(
          create: (_) => ApiService(),
          dispose: (_, service) => service.dispose(),
        ),
        // Image processing provider
        ChangeNotifierProxyProvider<ApiService, ImageProcessingProvider>(
          create: (_) => ImageProcessingProvider(ApiService()),
          update: (_, apiService, previous) => 
            previous ?? ImageProcessingProvider(apiService),
        ),
      ],
      child: MaterialApp(
        title: 'Scrapp\'d',
        debugShowCheckedModeBanner: EnvironmentConfig.showDebugBanner,
        theme: ThemeData.light(),
        darkTheme: ThemeData.dark(),
        themeMode: ThemeMode.system,
        home: const SplashScreen(),
      ),
    );
  }
}

class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen> {
  @override
  void initState() {
    super.initState();
    _checkHealthAndNavigate();
  }

  Future<void> _checkHealthAndNavigate() async {
    // Show splash for minimum 2 seconds
    await Future.delayed(const Duration(seconds: 2));
    if (!mounted) return;

    // Check API health
    final apiService = context.read<ApiService>();
    final isHealthy = await apiService.healthCheck();

    if (!mounted) return;

    if (!isHealthy) {
      // Show error dialog
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (context) => AlertDialog(
          title: const Text('Connection Error'),
          content: const Text(
            'Cannot connect to the server. Please make sure the backend is running.',
          ),
          actions: [
            TextButton(
              onPressed: () {
                Navigator.pop(context);
                _checkHealthAndNavigate();
              },
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    } else {
      // Navigate to home
      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (_) => const HomeScreen()),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: "Scrapp'd",
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      home: const RootScreen(),
    );
  }
}
