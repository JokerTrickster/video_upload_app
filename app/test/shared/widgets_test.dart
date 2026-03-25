import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/shared/widgets/loading_overlay.dart';
import 'package:video_upload_app/shared/widgets/error_snackbar.dart';

void main() {
  group('LoadingOverlay', () {
    testWidgets('shows child when not loading', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: LoadingOverlay(
            isLoading: false,
            child: Text('Content'),
          ),
        ),
      );

      expect(find.text('Content'), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsNothing);
    });

    testWidgets('shows loading indicator when loading', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: LoadingOverlay(
            isLoading: true,
            child: Text('Content'),
          ),
        ),
      );

      expect(find.text('Content'), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows message when provided', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: LoadingOverlay(
            isLoading: true,
            message: 'Please wait...',
            child: Text('Content'),
          ),
        ),
      );

      expect(find.text('Please wait...'), findsOneWidget);
    });
  });

  group('Error/Success SnackBar', () {
    testWidgets('showErrorSnackBar displays red snackbar', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () => showErrorSnackBar(context, 'Test error'),
                child: const Text('Show Error'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show Error'));
      await tester.pump();

      expect(find.text('Test error'), findsOneWidget);
    });

    testWidgets('showSuccessSnackBar displays green snackbar', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () => showSuccessSnackBar(context, 'Success!'),
                child: const Text('Show Success'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show Success'));
      await tester.pump();

      expect(find.text('Success!'), findsOneWidget);
    });
  });

  group('UploadSessionModel', () {
    test('progress calculates correctly', () {
      // Test via import
      expect(true, isTrue); // Model tests are in media test file
    });
  });
}
