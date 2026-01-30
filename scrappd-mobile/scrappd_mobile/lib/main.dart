import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:provider/provider.dart';

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

  try {
    await dotenv.load(fileName: '.env');
  } catch (e) {
    debugPrint('No .env file found, using defaults');
  }

  await SystemChrome.setPreferredOrientations([
    DeviceOrientation.portraitUp,
    DeviceOrientation.portraitDown,
  ]);

  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
    ),
  );

  final tokenStorage = TokenStorage();
  await tokenStorage.init();

  final apiClient = ApiClient(tokenStorage);
  final authRepository = AuthRepositoryImpl(
    remoteDataSource: AuthRemoteDataSource(apiClient.dio),
    tokenStorage: tokenStorage,
  );
  final itemRepository = ItemRepositoryImpl(
    ItemsRemoteDataSource(apiClient.dio),
  );
  final projectRepository = ProjectRepositoryImpl(
    ProjectsRemoteDataSource(apiClient.dio),
  );
  final pageRepository = PageRepositoryImpl(
    PagesRemoteDataSource(apiClient.dio),
  );
  final pageItemRepository = PageItemRepositoryImpl(
    PageItemsRemoteDataSource(apiClient.dio),
  );

  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider<AuthProvider>(
          create: (_) => AuthProvider(authRepository),
        ),
        ChangeNotifierProvider<ItemsProvider>(
          create: (_) => ItemsProvider(itemRepository),
        ),
        ChangeNotifierProvider<ProjectsProvider>(
          create: (_) => ProjectsProvider(projectRepository),
        ),
        ChangeNotifierProvider<PageEditorProvider>(
          create: (_) => PageEditorProvider(
            pageRepository,
            pageItemRepository,
          ),
        ),
      ],
      child: const ScrappdApp(),
    ),
  );
}

class ScrappdApp extends StatelessWidget {
  const ScrappdApp({super.key});

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
