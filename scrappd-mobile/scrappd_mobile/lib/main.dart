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
import 'data/services/page_export_service.dart';
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

  // SharedPreferences-backed token storage must be initialized before providers use it.
  final tokenStorage = TokenStorage();
  await tokenStorage.init();

  runApp(ScrappdApp(tokenStorage: tokenStorage));
}

class ScrappdApp extends StatelessWidget {
  const ScrappdApp({
    required this.tokenStorage,
    super.key,
  });

  final TokenStorage tokenStorage;

  @override
  Widget build(BuildContext context) {
    // Create shared instances
    final apiClient = ApiClient(tokenStorage);
    final dio = apiClient.dio;

    return MultiProvider(
      providers: [
        // Core services
        Provider<TokenStorage>.value(value: tokenStorage),
        Provider<ApiClient>.value(value: apiClient),
        Provider<PageExportService>(
          create: (context) => PageExportService(
            context.read<ApiClient>().dio,
          ),
        ),

        // Auth
        Provider<AuthRemoteDataSource>(
          create: (_) => AuthRemoteDataSource(dio),
        ),
        Provider<AuthRepositoryImpl>(
          create: (context) => AuthRepositoryImpl(
            remoteDataSource: context.read<AuthRemoteDataSource>(),
            tokenStorage: tokenStorage,
          ),
        ),
        ChangeNotifierProvider<AuthProvider>(
          create: (context) => AuthProvider(context.read<AuthRepositoryImpl>()),
        ),

        // Projects
        Provider<ProjectsRemoteDataSource>(
          create: (_) => ProjectsRemoteDataSource(dio),
        ),
        Provider<ProjectRepositoryImpl>(
          create: (context) =>
              ProjectRepositoryImpl(context.read<ProjectsRemoteDataSource>()),
        ),
        ChangeNotifierProvider<ProjectsProvider>(
          create: (context) =>
              ProjectsProvider(context.read<ProjectRepositoryImpl>()),
        ),

        // Items
        Provider<ItemsRemoteDataSource>(
          create: (_) => ItemsRemoteDataSource(dio),
        ),
        Provider<ItemRepositoryImpl>(
          create: (context) =>
              ItemRepositoryImpl(context.read<ItemsRemoteDataSource>()),
        ),
        ChangeNotifierProvider<ItemsProvider>(
          create: (context) =>
              ItemsProvider(context.read<ItemRepositoryImpl>()),
        ),

        // Pages
        Provider<PagesRemoteDataSource>(
          create: (_) => PagesRemoteDataSource(dio),
        ),
        Provider<PageRepositoryImpl>(
          create: (context) =>
              PageRepositoryImpl(context.read<PagesRemoteDataSource>()),
        ),

        // Page Items
        Provider<PageItemsRemoteDataSource>(
          create: (_) => PageItemsRemoteDataSource(dio),
        ),
        Provider<PageItemRepositoryImpl>(
          create: (context) =>
              PageItemRepositoryImpl(context.read<PageItemsRemoteDataSource>()),
        ),

        // Page Editor (depends on page and page item repositories)
        ChangeNotifierProvider<PageEditorProvider>(
          create: (context) => PageEditorProvider(
            context.read<PageRepositoryImpl>(),
            context.read<PageItemRepositoryImpl>(),
          ),
        ),
      ],
      child: MaterialApp(
        title: "Scrapp'd",
        debugShowCheckedModeBanner: EnvironmentConfig.showDebugBanner,
        theme: AppTheme.lightTheme,
        home: const RootScreen(),
      ),
    );
  }
}
