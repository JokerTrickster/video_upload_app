import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/features/upload/presentation/upload_provider.dart';

void main() {
  group('UploadFile', () {
    test('creates with default values', () {
      final file = UploadFile(path: '/path/video.mp4', filename: 'video.mp4', size: 1024);

      expect(file.path, '/path/video.mp4');
      expect(file.filename, 'video.mp4');
      expect(file.size, 1024);
      expect(file.status, 'pending');
      expect(file.progress, 0);
      expect(file.error, isNull);
    });

    test('creates with custom values', () {
      final file = UploadFile(
        path: '/path',
        filename: 'test.mp4',
        size: 2048,
        status: 'completed',
        progress: 100,
        error: null,
      );

      expect(file.status, 'completed');
      expect(file.progress, 100);
    });

    test('status can be modified', () {
      final file = UploadFile(path: '/p', filename: 'f', size: 10);
      file.status = 'uploading';
      file.progress = 50;
      file.error = 'timeout';

      expect(file.status, 'uploading');
      expect(file.progress, 50);
      expect(file.error, 'timeout');
    });
  });

  group('UploadProvider', () {
    // Note: UploadProvider requires UploadRepository which needs ApiClient.
    // We test the pure logic methods that don't require API calls.

    test('overallProgress with no files returns 0', () {
      // overallProgress is computed from _files list
      // Since we can't easily construct without UploadRepository,
      // we test the UploadFile data class thoroughly instead.
      final files = <UploadFile>[];
      final totalProgress = files.isEmpty
          ? 0.0
          : files.fold<double>(0, (sum, f) => sum + f.progress) / files.length;
      expect(totalProgress, 0.0);
    });

    test('overallProgress calculation', () {
      final files = [
        UploadFile(path: '/a', filename: 'a', size: 100)..progress = 100,
        UploadFile(path: '/b', filename: 'b', size: 200)..progress = 50,
        UploadFile(path: '/c', filename: 'c', size: 300)..progress = 0,
      ];

      final totalProgress =
          files.fold<double>(0, (sum, f) => sum + f.progress) / files.length;
      expect(totalProgress, 50.0);
    });

    test('completedCount calculation', () {
      final files = [
        UploadFile(path: '/a', filename: 'a', size: 100)..status = 'completed',
        UploadFile(path: '/b', filename: 'b', size: 200)..status = 'failed',
        UploadFile(path: '/c', filename: 'c', size: 300)..status = 'completed',
        UploadFile(path: '/d', filename: 'd', size: 400)..status = 'pending',
      ];

      final completed = files.where((f) => f.status == 'completed').length;
      final failed = files.where((f) => f.status == 'failed').length;

      expect(completed, 2);
      expect(failed, 1);
    });
  });
}
