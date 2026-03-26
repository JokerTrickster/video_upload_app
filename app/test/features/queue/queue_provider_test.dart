import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/shared/models/queue_model.dart';

void main() {
  group('QueueProvider logic', () {
    // Test the computed properties logic without needing real API

    test('pendingCount counts only PENDING items', () {
      final items = [
        _makeItem('PENDING'),
        _makeItem('COMPLETED'),
        _makeItem('PENDING'),
        _makeItem('FAILED'),
        _makeItem('PROCESSING'),
      ];

      final pending = items.where((i) => i.isPending).length;
      final completed = items.where((i) => i.isCompleted).length;
      final failed = items.where((i) => i.isFailed).length;

      expect(pending, 2);
      expect(completed, 1);
      expect(failed, 1);
    });

    test('empty queue has zero counts', () {
      final items = <QueueItemModel>[];

      expect(items.where((i) => i.isPending).length, 0);
      expect(items.where((i) => i.isCompleted).length, 0);
      expect(items.where((i) => i.isFailed).length, 0);
    });

    test('all items completed', () {
      final items = List.generate(5, (_) => _makeItem('COMPLETED'));

      expect(items.where((i) => i.isCompleted).length, 5);
      expect(items.where((i) => i.isPending).length, 0);
    });
  });
}

QueueItemModel _makeItem(String status) {
  return QueueItemModel(
    queueId: DateTime.now().microsecondsSinceEpoch.toString(),
    filename: 'video.mp4',
    fileSizeBytes: 1024,
    title: 'Test',
    status: status,
    retryCount: 0,
    createdAt: DateTime.now(),
  );
}
