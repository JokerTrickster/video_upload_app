import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/shared/models/queue_model.dart';

// QueueScreen widget tests are not included because the screen triggers
// API calls in initState via postFrameCallback (loadQueue), which creates
// pending Dio timers that are incompatible with the test framework.
// The screen logic is covered through QueueProvider and QueueItemModel unit tests.

void main() {
  group('QueueScreen data layer', () {
    test('QueueItemModel status icon logic - pending', () {
      final item = QueueItemModel(
        queueId: '1', filename: 'video.mp4', fileSizeBytes: 1024,
        title: 'Test', status: 'PENDING', retryCount: 0,
        createdAt: DateTime.now(),
      );

      expect(item.isPending, isTrue);
      expect(item.isCompleted, isFalse);
      expect(item.isFailed, isFalse);
      expect(item.isProcessing, isFalse);
    });

    test('QueueItemModel status icon logic - processing', () {
      final item = QueueItemModel(
        queueId: '2', filename: 'video.mp4', fileSizeBytes: 2048,
        title: 'Test', status: 'PROCESSING', retryCount: 0,
        createdAt: DateTime.now(),
      );

      expect(item.isProcessing, isTrue);
      expect(item.isPending, isFalse);
    });

    test('QueueItemModel with error message', () {
      final item = QueueItemModel(
        queueId: '3', filename: 'fail.mp4', fileSizeBytes: 100,
        title: 'Failed', status: 'FAILED', retryCount: 5,
        errorMessage: 'Max retries exceeded',
        createdAt: DateTime.now(),
        processedAt: DateTime.now(),
      );

      expect(item.isFailed, isTrue);
      expect(item.errorMessage, 'Max retries exceeded');
      expect(item.retryCount, 5);
      expect(item.processedAt, isNotNull);
    });

    test('QuotaModel display logic - can upload', () {
      final quota = QuotaModel(
        date: '2026-03-25', unitsUsed: 3200, unitsMax: 10000,
        uploadsToday: 2, remainingUploads: 4, canUpload: true,
      );

      expect(quota.canUpload, isTrue);
      expect(quota.usagePercent, 32.0);
    });

    test('QuotaModel display logic - exhausted', () {
      final quota = QuotaModel(
        date: '2026-03-25', unitsUsed: 9600, unitsMax: 10000,
        uploadsToday: 6, remainingUploads: 0, canUpload: false,
      );

      expect(quota.canUpload, isFalse);
      expect(quota.remainingUploads, 0);
    });
  });
}
