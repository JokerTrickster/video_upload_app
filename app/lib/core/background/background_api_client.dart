import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../constants/api_constants.dart';

class BackgroundApiClient {
  late final Dio _dio;
  final FlutterSecureStorage _storage = const FlutterSecureStorage();

  Future<void> initialize() async {
    _dio = Dio(BaseOptions(
      baseUrl: ApiConstants.baseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 10),
      headers: {'Content-Type': 'application/json'},
    ));

    final token = await _storage.read(key: 'access_token');
    if (token != null) {
      _dio.options.headers['Authorization'] = 'Bearer $token';
    }
  }

  Future<bool> refreshToken() async {
    try {
      final refreshToken = await _storage.read(key: 'refresh_token');
      if (refreshToken == null) return false;

      final response = await Dio().post(
        '${ApiConstants.baseUrl}${ApiConstants.authRefresh}',
        data: {'refresh_token': refreshToken},
      );

      if (response.statusCode == 200) {
        final newToken = response.data['data']['access_token'] as String;
        await _storage.write(key: 'access_token', value: newToken);
        _dio.options.headers['Authorization'] = 'Bearer $newToken';
        return true;
      }
    } catch (_) {}
    return false;
  }

  Future<void> uploadVideo({
    required String sessionId,
    required String filePath,
    required String filename,
    required int fileSize,
    void Function(int, int)? onProgress,
  }) async {
    final formData = FormData.fromMap({
      'session_id': sessionId,
      'title': filename,
      'description': '',
      'file': await MultipartFile.fromFile(filePath, filename: filename),
    });

    final uploadOptions = Options(
      headers: {'Content-Type': 'multipart/form-data'},
      receiveTimeout: const Duration(minutes: 30),
      sendTimeout: const Duration(minutes: 30),
    );

    try {
      await _dio.post(
        ApiConstants.mediaUploadVideo,
        data: formData,
        options: uploadOptions,
        onSendProgress: onProgress,
      );
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        final refreshed = await refreshToken();
        if (refreshed) {
          final retryFormData = FormData.fromMap({
            'session_id': sessionId,
            'title': filename,
            'description': '',
            'file':
                await MultipartFile.fromFile(filePath, filename: filename),
          });
          await _dio.post(
            ApiConstants.mediaUploadVideo,
            data: retryFormData,
            options: uploadOptions,
            onSendProgress: onProgress,
          );
          return;
        }
      }
      rethrow;
    }
  }

  Future<void> completeSession(String sessionId) async {
    await _dio.post(
      ApiConstants.mediaUploadComplete,
      data: {'session_id': sessionId},
    );
  }
}
