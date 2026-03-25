import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:video_upload_app/core/api/api_client.dart';
import 'package:video_upload_app/features/auth/data/auth_repository.dart';
import 'package:video_upload_app/features/auth/presentation/auth_provider.dart';
import 'package:video_upload_app/features/auth/presentation/login_screen.dart';

void main() {
  Widget buildTestWidget() {
    final apiClient = ApiClient();
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(
          create: (_) => AuthProvider(AuthRepository(apiClient), apiClient),
        ),
      ],
      child: const MaterialApp(home: LoginScreen()),
    );
  }

  group('LoginScreen', () {
    testWidgets('renders app title', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('Video Upload'), findsOneWidget);
    });

    testWidgets('shows Google sign-in button', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.text('Sign in with Google'), findsOneWidget);
    });

    testWidgets('shows cloud upload icon', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.cloud_upload_rounded), findsOneWidget);
    });
  });
}
