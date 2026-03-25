import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/shared/models/user_model.dart';
import 'package:video_upload_app/shared/models/upload_session_model.dart';

void main() {
  group('UserModel', () {
    test('parses from JSON', () {
      final json = {
        'id': 'user-123',
        'email': 'test@example.com',
        'google_id': 'google-456',
        'youtube_channel_id': 'UC123',
        'youtube_channel_name': 'Test Channel',
        'profile_image_url': 'https://example.com/pic.jpg',
        'created_at': '2026-03-25T10:00:00Z',
      };

      final user = UserModel.fromJson(json);

      expect(user.id, 'user-123');
      expect(user.email, 'test@example.com');
      expect(user.googleId, 'google-456');
      expect(user.youtubeChannelId, 'UC123');
      expect(user.youtubeChannelName, 'Test Channel');
    });

    test('handles null optional fields', () {
      final json = {
        'id': 'user-123',
        'email': 'test@example.com',
        'google_id': 'google-456',
        'created_at': '2026-03-25T10:00:00Z',
      };

      final user = UserModel.fromJson(json);

      expect(user.youtubeChannelId, isNull);
      expect(user.youtubeChannelName, isNull);
      expect(user.profileImageUrl, isNull);
    });
  });

  group('AuthTokens', () {
    test('parses from JSON', () {
      final json = {
        'access_token': 'abc',
        'refresh_token': 'xyz',
        'expires_in': 900,
        'token_type': 'Bearer',
      };

      final tokens = AuthTokens.fromJson(json);

      expect(tokens.accessToken, 'abc');
      expect(tokens.refreshToken, 'xyz');
      expect(tokens.expiresIn, 900);
      expect(tokens.tokenType, 'Bearer');
    });
  });

  group('UploadSessionModel', () {
    test('parses from JSON', () {
      final json = {
        'session_id': 'sess-123',
        'user_id': 'user-456',
        'total_files': 10,
        'completed_files': 5,
        'failed_files': 1,
        'total_bytes': 1048576,
        'uploaded_bytes': 524288,
        'session_status': 'ACTIVE',
        'started_at': '2026-03-25T10:00:00Z',
      };

      final session = UploadSessionModel.fromJson(json);

      expect(session.sessionId, 'sess-123');
      expect(session.totalFiles, 10);
      expect(session.completedFiles, 5);
      expect(session.failedFiles, 1);
      expect(session.isActive, isTrue);
    });

    test('calculates progress correctly', () {
      final session = UploadSessionModel(
        sessionId: '1', userId: '2', totalFiles: 10,
        completedFiles: 5, failedFiles: 0, totalBytes: 1000,
        uploadedBytes: 500, sessionStatus: 'ACTIVE',
        startedAt: DateTime.now(),
      );

      expect(session.progress, 50.0);
      expect(session.pendingFiles, 5);
    });

    test('handles zero total bytes', () {
      final session = UploadSessionModel(
        sessionId: '1', userId: '2', totalFiles: 0,
        completedFiles: 0, failedFiles: 0, totalBytes: 0,
        uploadedBytes: 0, sessionStatus: 'ACTIVE',
        startedAt: DateTime.now(),
      );

      expect(session.progress, 0);
    });
  });
}
