import 'package:go_router/go_router.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/media/presentation/media_detail_screen.dart';
import '../../features/media/presentation/media_list_screen.dart';
import '../../features/queue/presentation/queue_screen.dart';
import '../../features/settings/presentation/settings_screen.dart';
import '../../features/upload/presentation/session_status_screen.dart';
import '../../features/upload/presentation/upload_screen.dart';

class AppRouter {
  static final GoRouter router = GoRouter(
    initialLocation: '/login',
    routes: [
      GoRoute(
        path: '/login',
        name: 'login',
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/media',
        name: 'media',
        builder: (context, state) => const MediaListScreen(),
      ),
      GoRoute(
        path: '/media/:assetId',
        name: 'mediaDetail',
        builder: (context, state) => MediaDetailScreen(
          assetId: state.pathParameters['assetId']!,
        ),
      ),
      GoRoute(
        path: '/upload',
        name: 'upload',
        builder: (context, state) => const UploadScreen(),
      ),
      GoRoute(
        path: '/queue',
        name: 'queue',
        builder: (context, state) => const QueueScreen(),
      ),
      GoRoute(
        path: '/upload/status/:sessionId',
        name: 'sessionStatus',
        builder: (context, state) => SessionStatusScreen(
          sessionId: state.pathParameters['sessionId']!,
        ),
      ),
      GoRoute(
        path: '/settings',
        name: 'settings',
        builder: (context, state) => const SettingsScreen(),
      ),
    ],
  );
}
