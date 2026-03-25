import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:video_upload_app/core/api/api_client.dart';
import 'package:video_upload_app/features/upload/data/upload_repository.dart';
import 'package:video_upload_app/features/upload/presentation/upload_provider.dart';
import 'package:video_upload_app/features/upload/presentation/upload_screen.dart';

void main() {
  Widget buildTestWidget() {
    final apiClient = ApiClient();
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(
          create: (_) => UploadProvider(UploadRepository(apiClient)),
        ),
      ],
      child: const MaterialApp(home: UploadScreen()),
    );
  }

  group('UploadScreen', () {
    testWidgets('renders app bar with title', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pump();

      expect(find.text('Upload Videos'), findsOneWidget);
    });

    testWidgets('shows empty state when no files selected', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pump();

      expect(find.text('No files selected'), findsOneWidget);
      expect(find.text('Tap + to select videos'), findsOneWidget);
    });

    testWidgets('shows add files button', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pump();

      expect(find.text('Add Files'), findsOneWidget);
    });
  });
}
