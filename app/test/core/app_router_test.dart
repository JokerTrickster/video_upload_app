import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:video_upload_app/core/api/api_client.dart';
import 'package:video_upload_app/core/router/app_router.dart';
import 'package:video_upload_app/features/auth/data/auth_repository.dart';
import 'package:video_upload_app/features/auth/presentation/auth_provider.dart';
import 'package:video_upload_app/features/media/data/media_repository.dart';
import 'package:video_upload_app/features/media/presentation/media_provider.dart';
import 'package:video_upload_app/features/upload/data/upload_repository.dart';
import 'package:video_upload_app/features/upload/presentation/upload_provider.dart';

Widget buildTestApp() {
  final apiClient = ApiClient();
  return MultiProvider(
    providers: [
      Provider<ApiClient>.value(value: apiClient),
      ChangeNotifierProvider(
        create: (_) => AuthProvider(AuthRepository(apiClient), apiClient),
      ),
      ChangeNotifierProvider(
        create: (_) => MediaProvider(MediaRepository(apiClient)),
      ),
      ChangeNotifierProvider(
        create: (_) => UploadProvider(UploadRepository(apiClient)),
      ),
    ],
    child: MaterialApp.router(routerConfig: AppRouter.router),
  );
}

void main() {
  group('AppRouter', () {
    test('router is not null', () {
      expect(AppRouter.router, isNotNull);
    });

    testWidgets('navigates to login screen by default', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pumpAndSettle();

      expect(find.text('Video Upload'), findsOneWidget);
    });
  });
}
