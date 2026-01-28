import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

import 'core/constants/theme_constants.dart';
import 'data/services/api_service.dart';
import 'data/services/auth_service.dart';
import 'data/services/pages_service.dart';
import 'data/services/projects_service.dart';
import 'data/services/secure_storage_service.dart';
import 'presentation/providers/auth_provider.dart';
import 'presentation/providers/image_provider.dart';
import 'presentation/providers/pages_provider.dart';
import 'presentation/providers/projects_provider.dart';
import 'presentation/screens/auth/login_screen.dart';
import 'presentation/screens/projects/projects_list_screen.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Load environment variables
  try {
    await dotenv.load(fileName: ".env");
  } catch (e) {
    debugPrint('No .env file found, using defaults');
  }

  // Set portrait orientation only
  await SystemChrome.setPreferredOrientations([
    DeviceOrientation.portraitUp,
    DeviceOrientation.portraitDown,
  ]);

  // Set status bar style
  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.light,
    ),
  );

  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        // Services
        Provider<SecureStorageService>(
          create: (_) => SecureStorageService(),
        ),
        Provider<ApiService>(
          create: (_) => ApiService(),
        ),
        ProxyProvider<SecureStorageService, AuthService>(
          update: (_, storage, __) => AuthService(storage),
        ),
        ProxyProvider<SecureStorageService, ProjectsService>(
          update: (_, storage, __) => ProjectsService(storage),
        ),
        ProxyProvider<SecureStorageService, PagesService>(
          update: (_, storage, __) => PagesService(storage),
        ),

        // Providers
        ChangeNotifierProxyProvider2<AuthService, SecureStorageService,
            AuthProvider>(
          create: (context) => AuthProvider(
            context.read<AuthService>(),
            context.read<SecureStorageService>(),
          ),
          update: (_, authService, storage, previous) =>
              previous ?? AuthProvider(authService, storage),
        ),
        ChangeNotifierProxyProvider<ProjectsService, ProjectsProvider>(
          create: (context) => ProjectsProvider(
            context.read<ProjectsService>(),
          ),
          update: (_, service, previous) =>
              previous ?? ProjectsProvider(service),
        ),
        ChangeNotifierProxyProvider<PagesService, PagesProvider>(
          create: (context) => PagesProvider(
            context.read<PagesService>(),
          ),
          update: (_, service, previous) => previous ?? PagesProvider(service),
        ),
        ChangeNotifierProvider<ImageProcessingProvider>(
          create: (context) => ImageProcessingProvider(
            context.read<ApiService>(),
          ),
        ),
      ],
      child: MaterialApp(
        title: 'Scrapp\'d',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme,
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
    _initializeApp();
  }

  Future<void> _initializeApp() async {
    // Show splash for minimum 2 seconds
    await Future.delayed(const Duration(seconds: 2));

    if (!mounted) return;

    // Check API health
    final apiService = context.read<ApiService>();
    final isHealthy = await apiService.healthCheck();

    if (!mounted) return;

    if (!isHealthy) {
      // Show error dialog
      _showConnectionError();
      return;
    }

    // Initialize auth state
    final authProvider = context.read<AuthProvider>();
    await authProvider.initialize();

    if (!mounted) return;

    // Navigate based on auth state
    _navigateBasedOnAuth();
  }

  void _showConnectionError() {
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
              _initializeApp();
            },
            child: const Text('Retry'),
          ),
        ],
      ),
    );
  }

  void _navigateBasedOnAuth() {
    final authProvider = context.read<AuthProvider>();

    if (authProvider.isAuthenticated) {
      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (_) => const ProjectsListScreen()),
      );
    } else {
      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (_) => const AuthWrapper()),
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
                'Digital Scrapbooking',
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

/// Wrapper that listens to auth state changes and navigates accordingly
class AuthWrapper extends StatefulWidget {
  const AuthWrapper({super.key});

  @override
  State<AuthWrapper> createState() => _AuthWrapperState();
}

class _AuthWrapperState extends State<AuthWrapper> {
  @override
  void initState() {
    super.initState();
    // Listen to auth changes
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final authProvider = context.read<AuthProvider>();
      authProvider.addListener(_onAuthStateChanged);
    });
  }

  @override
  void dispose() {
    // Remove listener when disposed
    try {
      final authProvider = context.read<AuthProvider>();
      authProvider.removeListener(_onAuthStateChanged);
    } catch (_) {}
    super.dispose();
  }

  void _onAuthStateChanged() {
    if (!mounted) return;

    final authProvider = context.read<AuthProvider>();
    if (authProvider.isAuthenticated) {
      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (_) => const ProjectsListScreen()),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return const LoginScreen();
  }
}
