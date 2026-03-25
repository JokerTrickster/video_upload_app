import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'core/api/api_client.dart';
import 'core/router/app_router.dart';
import 'features/auth/data/auth_repository.dart';
import 'features/auth/presentation/auth_provider.dart';
import 'features/media/data/media_repository.dart';
import 'features/media/presentation/media_provider.dart';
import 'features/queue/data/queue_repository.dart';
import 'features/queue/presentation/queue_provider.dart';
import 'features/upload/data/upload_repository.dart';
import 'features/upload/presentation/upload_provider.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    final apiClient = ApiClient();
    final authRepository = AuthRepository(apiClient);
    final mediaRepository = MediaRepository(apiClient);
    final uploadRepository = UploadRepository(apiClient);
    final queueRepository = QueueRepository(apiClient);

    return MultiProvider(
      providers: [
        Provider<ApiClient>.value(value: apiClient),
        Provider<UploadRepository>.value(value: uploadRepository),
        ChangeNotifierProvider(
          create: (_) => QueueProvider(queueRepository),
        ),
        ChangeNotifierProvider(
          create: (_) => AuthProvider(authRepository, apiClient),
        ),
        ChangeNotifierProvider(
          create: (_) => MediaProvider(mediaRepository),
        ),
        ChangeNotifierProvider(
          create: (_) => UploadProvider(uploadRepository),
        ),
      ],
      child: MaterialApp.router(
        title: 'Video Upload App',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(
          colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
          useMaterial3: true,
        ),
        routerConfig: AppRouter.router,
      ),
    );
  }
}
