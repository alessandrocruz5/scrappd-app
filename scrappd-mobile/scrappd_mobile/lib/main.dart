import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:scrappd_mobile/core/config/environment.dart';

import 'core/constants/api_constants.dart';
import 'core/constants/theme_constants.dart';
import 'data/services/api_service.dart';
import 'presentation/providers/image_provider.dart';
import 'presentation/screens/home_screen.dart';

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

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
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
    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          gradient: AppTheme.primaryGradient,
        ),
        child: const Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.auto_fix_high,
                size: 80,
                color: Colors.white,
              ),
              SizedBox(height: AppTheme.spacing24),
              Text(
                "Scrapp'd",
                style: TextStyle(
                  fontSize: 48,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                  letterSpacing: -1,
                ),
              ),
              SizedBox(height: AppTheme.spacing8),
              Text(
                'AI Background Remover',
                style: TextStyle(
                  fontSize: 16,
                  color: Colors.white70,
                ),
              ),
              SizedBox(height: AppTheme.spacing48),
              CircularProgressIndicator(
                valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
