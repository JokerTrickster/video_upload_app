import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/core/router/app_router.dart';

void main() {
  group('AppRouter', () {
    test('router is not null', () {
      expect(AppRouter.router, isNotNull);
    });

    test('initial location is /login', () {
      // GoRouter configuration check
      final config = AppRouter.router.configuration;
      expect(config.routes.isNotEmpty, isTrue);
    });

    testWidgets('navigates to login screen by default', (tester) async {
      await tester.pumpWidget(
        MaterialApp.router(
          routerConfig: AppRouter.router,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Login Screen - TODO'), findsOneWidget);
    });

    testWidgets('navigates to media list screen', (tester) async {
      await tester.pumpWidget(
        MaterialApp.router(
          routerConfig: AppRouter.router,
        ),
      );
      await tester.pumpAndSettle();

      AppRouter.router.go('/media');
      await tester.pumpAndSettle();

      expect(find.text('Media List Screen - TODO'), findsOneWidget);
    });

    testWidgets('navigates to upload screen', (tester) async {
      await tester.pumpWidget(
        MaterialApp.router(
          routerConfig: AppRouter.router,
        ),
      );
      await tester.pumpAndSettle();

      AppRouter.router.go('/upload');
      await tester.pumpAndSettle();

      expect(find.text('Upload Screen - TODO'), findsOneWidget);
    });
  });
}
