import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/core/constants/api_constants.dart';

void main() {
  group('ApiConstants', () {
    test('baseUrl should point to localhost API', () {
      expect(ApiConstants.baseUrl, contains('localhost'));
      expect(ApiConstants.baseUrl, contains('/api/v1'));
    });

    group('Auth endpoints', () {
      test('authGoogleUrl starts with /auth', () {
        expect(ApiConstants.authGoogleUrl, startsWith('/auth'));
      });

      test('authGoogleCallback starts with /auth', () {
        expect(ApiConstants.authGoogleCallback, startsWith('/auth'));
      });

      test('authRefresh starts with /auth', () {
        expect(ApiConstants.authRefresh, startsWith('/auth'));
      });

      test('authMe starts with /auth', () {
        expect(ApiConstants.authMe, startsWith('/auth'));
      });

      test('authLogout starts with /auth', () {
        expect(ApiConstants.authLogout, startsWith('/auth'));
      });
    });

    group('Media endpoints', () {
      test('mediaUploadInitiate starts with /media', () {
        expect(ApiConstants.mediaUploadInitiate, startsWith('/media'));
      });

      test('mediaUploadVideo starts with /media', () {
        expect(ApiConstants.mediaUploadVideo, startsWith('/media'));
      });

      test('mediaUploadStatus starts with /media', () {
        expect(ApiConstants.mediaUploadStatus, startsWith('/media'));
      });

      test('mediaUploadComplete starts with /media', () {
        expect(ApiConstants.mediaUploadComplete, startsWith('/media'));
      });

      test('mediaUploadCancel starts with /media', () {
        expect(ApiConstants.mediaUploadCancel, startsWith('/media'));
      });

      test('mediaList starts with /media', () {
        expect(ApiConstants.mediaList, startsWith('/media'));
      });

      test('mediaDetail starts with /media', () {
        expect(ApiConstants.mediaDetail, startsWith('/media'));
      });
    });

    test('all endpoints are unique', () {
      final endpoints = [
        ApiConstants.authGoogleUrl,
        ApiConstants.authGoogleCallback,
        ApiConstants.authRefresh,
        ApiConstants.authMe,
        ApiConstants.authLogout,
        ApiConstants.mediaUploadInitiate,
        ApiConstants.mediaUploadVideo,
        ApiConstants.mediaUploadStatus,
        ApiConstants.mediaUploadComplete,
        ApiConstants.mediaUploadCancel,
        ApiConstants.mediaList,
        ApiConstants.mediaDetail,
      ];
      expect(endpoints.toSet().length, equals(endpoints.length));
    });

    test('no endpoint contains baseUrl (relative paths only)', () {
      final endpoints = [
        ApiConstants.authGoogleUrl,
        ApiConstants.authGoogleCallback,
        ApiConstants.authRefresh,
        ApiConstants.authMe,
        ApiConstants.authLogout,
        ApiConstants.mediaUploadInitiate,
        ApiConstants.mediaUploadVideo,
        ApiConstants.mediaUploadStatus,
        ApiConstants.mediaUploadComplete,
        ApiConstants.mediaUploadCancel,
        ApiConstants.mediaList,
        ApiConstants.mediaDetail,
      ];
      for (final endpoint in endpoints) {
        expect(endpoint, isNot(contains('http')),
            reason: '$endpoint should be a relative path');
      }
    });
  });
}
