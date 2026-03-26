import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:video_upload_app/core/api/api_client.dart';
import 'package:video_upload_app/features/upload/data/upload_repository.dart';
import 'package:video_upload_app/features/upload/presentation/upload_provider.dart';
import 'package:video_upload_app/shared/widgets/upload_progress_banner.dart';

void main() {
  late ApiClient apiClient;

  setUp(() {
    apiClient = ApiClient();
  });

  Widget buildTestWidget({Widget? child}) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(
          create: (_) => UploadProvider(UploadRepository(apiClient)),
        ),
      ],
      child: MaterialApp(
        home: Scaffold(
          body: Column(
            children: [
              const Expanded(child: SizedBox()),
              const UploadProgressBanner(),
              if (child != null) child,
            ],
          ),
        ),
      ),
    );
  }

  group('UploadProgressBanner', () {
    testWidgets('hidden when not uploading', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pump();

      // Banner should not show any progress text when not uploading
      expect(find.byType(LinearProgressIndicator), findsNothing);
    });

    testWidgets('renders without errors', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pump();

      expect(find.byType(UploadProgressBanner), findsOneWidget);
    });
  });
}
