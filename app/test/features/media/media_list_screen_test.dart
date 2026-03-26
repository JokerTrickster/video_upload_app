import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/shared/models/media_asset_model.dart';

void main() {
  group('MediaAssetModel', () {
    test('parses from JSON correctly', () {
      final json = {
        'asset_id': 'abc-123',
        'youtube_video_id': 'yt-456',
        'original_filename': 'test.mp4',
        'file_size_bytes': 1048576,
        'media_type': 'VIDEO',
        'sync_status': 'COMPLETED',
        'created_at': '2026-03-25T10:00:00Z',
        'retry_count': 0,
      };

      final asset = MediaAssetModel.fromJson(json);

      expect(asset.assetId, 'abc-123');
      expect(asset.youtubeVideoId, 'yt-456');
      expect(asset.originalFilename, 'test.mp4');
      expect(asset.fileSizeBytes, 1048576);
      expect(asset.isCompleted, isTrue);
    });

    test('fileSizeFormatted returns readable size', () {
      final small = MediaAssetModel.fromJson({
        'asset_id': '1', 'original_filename': 'a', 'file_size_bytes': 500,
        'media_type': 'VIDEO', 'sync_status': 'PENDING',
        'created_at': '2026-03-25T10:00:00Z', 'retry_count': 0,
      });
      expect(small.fileSizeFormatted, '500 B');

      final mb = MediaAssetModel.fromJson({
        'asset_id': '2', 'original_filename': 'b', 'file_size_bytes': 5242880,
        'media_type': 'VIDEO', 'sync_status': 'PENDING',
        'created_at': '2026-03-25T10:00:00Z', 'retry_count': 0,
      });
      expect(mb.fileSizeFormatted, '5.0 MB');
    });

    test('status checkers work correctly', () {
      MediaAssetModel makeAsset(String status) => MediaAssetModel.fromJson({
        'asset_id': '1', 'original_filename': 'a', 'file_size_bytes': 100,
        'media_type': 'VIDEO', 'sync_status': status,
        'created_at': '2026-03-25T10:00:00Z', 'retry_count': 0,
      });

      expect(makeAsset('COMPLETED').isCompleted, isTrue);
      expect(makeAsset('FAILED').isFailed, isTrue);
      expect(makeAsset('UPLOADING').isUploading, isTrue);
      expect(makeAsset('PENDING').isPending, isTrue);
    });
  });

  group('MediaAssetModel thumbnailUrl', () {
    test('parses thumbnail_url from JSON', () {
      final json = {
        'asset_id': '1', 'original_filename': 'a.mp4',
        'file_size_bytes': 100, 'media_type': 'VIDEO',
        'sync_status': 'COMPLETED', 'created_at': '2026-03-25T10:00:00Z',
        'retry_count': 0,
        'thumbnail_url': 'https://img.youtube.com/vi/abc/hqdefault.jpg',
        'youtube_video_id': 'abc',
      };

      final asset = MediaAssetModel.fromJson(json);
      expect(asset.thumbnailUrl, 'https://img.youtube.com/vi/abc/hqdefault.jpg');
    });

    test('effectiveThumbnailUrl returns thumbnailUrl when present', () {
      final asset = MediaAssetModel.fromJson({
        'asset_id': '1', 'original_filename': 'a.mp4',
        'file_size_bytes': 100, 'media_type': 'VIDEO',
        'sync_status': 'COMPLETED', 'created_at': '2026-03-25T10:00:00Z',
        'retry_count': 0,
        'thumbnail_url': 'https://custom-thumb.jpg',
        'youtube_video_id': 'xyz',
      });

      expect(asset.effectiveThumbnailUrl, 'https://custom-thumb.jpg');
    });

    test('effectiveThumbnailUrl falls back to videoId URL', () {
      final asset = MediaAssetModel.fromJson({
        'asset_id': '1', 'original_filename': 'a.mp4',
        'file_size_bytes': 100, 'media_type': 'VIDEO',
        'sync_status': 'COMPLETED', 'created_at': '2026-03-25T10:00:00Z',
        'retry_count': 0,
        'youtube_video_id': 'myVideoId',
      });

      expect(asset.effectiveThumbnailUrl,
          'https://img.youtube.com/vi/myVideoId/hqdefault.jpg');
    });

    test('effectiveThumbnailUrl returns null when no videoId', () {
      final asset = MediaAssetModel.fromJson({
        'asset_id': '1', 'original_filename': 'a.mp4',
        'file_size_bytes': 100, 'media_type': 'VIDEO',
        'sync_status': 'PENDING', 'created_at': '2026-03-25T10:00:00Z',
        'retry_count': 0,
      });

      expect(asset.effectiveThumbnailUrl, isNull);
    });

    test('effectiveThumbnailUrl ignores empty strings', () {
      final asset = MediaAssetModel.fromJson({
        'asset_id': '1', 'original_filename': 'a.mp4',
        'file_size_bytes': 100, 'media_type': 'VIDEO',
        'sync_status': 'COMPLETED', 'created_at': '2026-03-25T10:00:00Z',
        'retry_count': 0,
        'thumbnail_url': '',
        'youtube_video_id': '',
      });

      expect(asset.effectiveThumbnailUrl, isNull);
    });
  });

  group('MediaAssetListResponse', () {
    test('parses paginated response', () {
      final json = {
        'assets': [
          {
            'asset_id': '1', 'original_filename': 'a.mp4',
            'file_size_bytes': 100, 'media_type': 'VIDEO',
            'sync_status': 'COMPLETED', 'created_at': '2026-03-25T10:00:00Z',
            'retry_count': 0,
          },
        ],
        'pagination': {
          'page': 1, 'limit': 50, 'total': 1, 'total_pages': 1,
        },
      };

      final result = MediaAssetListResponse.fromJson(json);

      expect(result.assets.length, 1);
      expect(result.page, 1);
      expect(result.total, 1);
    });
  });
}
