import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:video_upload_app/core/api/api_client.dart';
import 'package:video_upload_app/features/auth/presentation/auth_provider.dart';
import 'package:video_upload_app/features/media/presentation/media_provider.dart';
import 'package:video_upload_app/features/upload/presentation/upload_provider.dart';
import 'package:video_upload_app/main.dart';

void main() {
  group('MyApp', () {
    testWidgets('renders without error and shows login screen', (tester) async {
      await tester.pumpWidget(const MyApp());
      await tester.pumpAndSettle();

      // Login screen should show the app title
      expect(find.text('Video Upload'), findsOneWidget);
      expect(find.text('Sign in with Google'), findsOneWidget);
    });

    testWidgets('provides all required providers', (tester) async {
      await tester.pumpWidget(const MyApp());
      await tester.pumpAndSettle();

      final element = tester.element(find.byType(MaterialApp).first);
      expect(Provider.of<ApiClient>(element, listen: false), isNotNull);
      expect(Provider.of<AuthProvider>(element, listen: false), isNotNull);
      expect(Provider.of<MediaProvider>(element, listen: false), isNotNull);
      expect(Provider.of<UploadProvider>(element, listen: false), isNotNull);
    });
  });
}
