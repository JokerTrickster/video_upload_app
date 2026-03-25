import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:video_upload_app/core/api/api_client.dart';
import 'package:video_upload_app/main.dart';

void main() {
  group('MyApp', () {
    testWidgets('renders without error and shows login screen', (tester) async {
      await tester.pumpWidget(const MyApp());
      await tester.pumpAndSettle();

      expect(find.text('Login Screen - TODO'), findsOneWidget);
    });

    testWidgets('provides ApiClient via Provider', (tester) async {
      await tester.pumpWidget(const MyApp());
      await tester.pumpAndSettle();

      final element = tester.element(find.byType(MaterialApp).first);
      final apiClient = Provider.of<ApiClient>(element, listen: false);
      expect(apiClient, isNotNull);
    });
  });
}
