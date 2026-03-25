import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/features/upload/presentation/upload_screen.dart';

void main() {
  group('UploadScreen', () {
    testWidgets('renders without error', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(home: UploadScreen()),
      );

      expect(find.byType(UploadScreen), findsOneWidget);
    });

    testWidgets('displays placeholder text', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(home: UploadScreen()),
      );

      expect(find.text('Upload Screen - TODO'), findsOneWidget);
    });

    testWidgets('uses Scaffold', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(home: UploadScreen()),
      );

      expect(find.byType(Scaffold), findsOneWidget);
    });
  });
}
