import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/shared/models/queue_model.dart';

void main() {
  group('QueueItemModel', () {
    test('parses from JSON correctly', () {
      final json = {
        'queue_id': 'q-123',
        'filename': 'video.mp4',
        'file_size_bytes': 5242880,
        'title': 'My Video',
        'status': 'PENDING',
        'retry_count': 0,
        'created_at': '2026-03-25T10:00:00Z',
      };

      final item = QueueItemModel.fromJson(json);

      expect(item.queueId, 'q-123');
      expect(item.filename, 'video.mp4');
      expect(item.fileSizeBytes, 5242880);
      expect(item.title, 'My Video');
      expect(item.status, 'PENDING');
      expect(item.retryCount, 0);
      expect(item.errorMessage, isNull);
      expect(item.processedAt, isNull);
    });

    test('parses with optional fields', () {
      final json = {
        'queue_id': 'q-456',
        'filename': 'fail.mp4',
        'file_size_bytes': 100,
        'title': '',
        'status': 'FAILED',
        'retry_count': 3,
        'error_message': 'Upload timeout',
        'created_at': '2026-03-25T10:00:00Z',
        'processed_at': '2026-03-25T11:00:00Z',
      };

      final item = QueueItemModel.fromJson(json);

      expect(item.isFailed, isTrue);
      expect(item.errorMessage, 'Upload timeout');
      expect(item.processedAt, isNotNull);
      expect(item.retryCount, 3);
    });

    test('handles null file_size_bytes and title', () {
      final json = {
        'queue_id': 'q-789',
        'filename': 'video.mp4',
        'status': 'PENDING',
        'created_at': '2026-03-25T10:00:00Z',
      };

      final item = QueueItemModel.fromJson(json);

      expect(item.fileSizeBytes, 0);
      expect(item.title, '');
      expect(item.retryCount, 0);
    });

    test('status checkers work correctly', () {
      QueueItemModel makeItem(String status) => QueueItemModel.fromJson({
            'queue_id': '1',
            'filename': 'a',
            'status': status,
            'created_at': '2026-03-25T10:00:00Z',
          });

      expect(makeItem('PENDING').isPending, isTrue);
      expect(makeItem('PENDING').isProcessing, isFalse);
      expect(makeItem('PROCESSING').isProcessing, isTrue);
      expect(makeItem('COMPLETED').isCompleted, isTrue);
      expect(makeItem('FAILED').isFailed, isTrue);
    });

    test('fileSizeFormatted returns readable sizes', () {
      QueueItemModel makeItem(int size) => QueueItemModel(
            queueId: '1',
            filename: 'a',
            fileSizeBytes: size,
            title: '',
            status: 'PENDING',
            retryCount: 0,
            createdAt: DateTime.now(),
          );

      expect(makeItem(500 * 1024).fileSizeFormatted, '500.0 KB');
      expect(makeItem(5 * 1024 * 1024).fileSizeFormatted, '5.0 MB');
      expect(makeItem(2 * 1024 * 1024 * 1024).fileSizeFormatted, '2.00 GB');
    });
  });

  group('QuotaModel', () {
    test('parses from JSON correctly', () {
      final json = {
        'date': '2026-03-25',
        'units_used': 3200,
        'units_max': 10000,
        'uploads_today': 2,
        'remaining_uploads': 4,
        'can_upload': true,
      };

      final quota = QuotaModel.fromJson(json);

      expect(quota.date, '2026-03-25');
      expect(quota.unitsUsed, 3200);
      expect(quota.unitsMax, 10000);
      expect(quota.uploadsToday, 2);
      expect(quota.remainingUploads, 4);
      expect(quota.canUpload, isTrue);
    });

    test('usagePercent calculates correctly', () {
      final quota = QuotaModel(
        date: '2026-03-25',
        unitsUsed: 5000,
        unitsMax: 10000,
        uploadsToday: 3,
        remainingUploads: 3,
        canUpload: true,
      );

      expect(quota.usagePercent, 50.0);
    });

    test('usagePercent handles zero max', () {
      final quota = QuotaModel(
        date: '2026-03-25',
        unitsUsed: 0,
        unitsMax: 0,
        uploadsToday: 0,
        remainingUploads: 0,
        canUpload: false,
      );

      expect(quota.usagePercent, 0);
    });

    test('exhausted quota', () {
      final quota = QuotaModel(
        date: '2026-03-25',
        unitsUsed: 9600,
        unitsMax: 10000,
        uploadsToday: 6,
        remainingUploads: 0,
        canUpload: false,
      );

      expect(quota.canUpload, isFalse);
      expect(quota.remainingUploads, 0);
      expect(quota.usagePercent, 96.0);
    });
  });
}
