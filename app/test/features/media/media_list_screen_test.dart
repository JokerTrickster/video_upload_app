import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/features/media/presentation/media_list_screen.dart';

void main() {
  group('MediaListScreen', () {
    testWidgets('renders without error', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(home: MediaListScreen()),
      );

      expect(find.byType(MediaListScreen), findsOneWidget);
    });

    testWidgets('displays placeholder text', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(home: MediaListScreen()),
      );

      expect(find.text('Media List Screen - TODO'), findsOneWidget);
    });

    testWidgets('uses Scaffold', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(home: MediaListScreen()),
      );

      expect(find.byType(Scaffold), findsOneWidget);
    });
  });
}
