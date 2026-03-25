import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/features/auth/presentation/login_screen.dart';

void main() {
  group('LoginScreen', () {
    testWidgets('renders without error', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(home: LoginScreen()),
      );

      expect(find.byType(LoginScreen), findsOneWidget);
    });

    testWidgets('displays placeholder text', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(home: LoginScreen()),
      );

      expect(find.text('Login Screen - TODO'), findsOneWidget);
    });

    testWidgets('uses Scaffold', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(home: LoginScreen()),
      );

      expect(find.byType(Scaffold), findsOneWidget);
    });
  });
}
