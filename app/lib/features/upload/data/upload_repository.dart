import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/constants/api_constants.dart';
import '../../../shared/models/upload_session_model.dart';

class UploadRepository {
  final ApiClient _apiClient;

  UploadRepository(this._apiClient);

  Future<UploadSessionModel> initiateSession({
    required int totalFiles,
    required int totalBytes,
  }) async {
    final response = await _apiClient.post(
      ApiConstants.mediaUploadInitiate,
      data: {'total_files': totalFiles, 'total_bytes': totalBytes},
    );
    final data = response.data['data'] as Map<String, dynamic>;
    // Backend returns session_id + started_at; construct minimal model
    return UploadSessionModel(
      sessionId: data['session_id'] as String,
      userId: '',
      totalFiles: totalFiles,
      completedFiles: 0,
      failedFiles: 0,
      totalBytes: totalBytes,
      uploadedBytes: 0,
      sessionStatus: 'ACTIVE',
      startedAt: DateTime.parse(data['started_at'] as String),
    );
  }

  Future<Map<String, dynamic>> uploadVideo({
    required String sessionId,
    required String filePath,
    required String filename,
    required int fileSize,
    String? title,
    String? description,
    void Function(int, int)? onProgress,
  }) async {
    final formData = FormData.fromMap({
      'session_id': sessionId,
      'title': title ?? filename,
      'description': description ?? '',
      'file': await MultipartFile.fromFile(filePath, filename: filename),
    });

    final response = await _apiClient.upload(
      ApiConstants.mediaUploadVideo,
      formData: formData,
      onSendProgress: onProgress,
    );
    return response.data['data'] as Map<String, dynamic>;
  }

  Future<UploadSessionModel> getSessionStatus(String sessionId) async {
    final response = await _apiClient.get(
      '${ApiConstants.mediaUploadStatus}/$sessionId',
    );
    return UploadSessionModel.fromJson(
      response.data['data'] as Map<String, dynamic>,
    );
  }

  Future<void> completeSession(String sessionId) async {
    await _apiClient.post(
      ApiConstants.mediaUploadComplete,
      data: {'session_id': sessionId},
    );
  }

  Future<void> cancelSession(String sessionId) async {
    await _apiClient.post(
      ApiConstants.mediaUploadCancel,
      data: {'session_id': sessionId},
    );
  }
}
