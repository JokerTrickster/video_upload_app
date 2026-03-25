import '../../../core/api/api_client.dart';
import '../../../core/constants/api_constants.dart';
import '../../../shared/models/queue_model.dart';

class QueueRepository {
  final ApiClient _apiClient;

  QueueRepository(this._apiClient);

  Future<QueueItemModel> addToQueue({
    required String filePath,
    required String filename,
    required int fileSizeBytes,
    String? title,
    String? description,
  }) async {
    final response = await _apiClient.post(
      ApiConstants.queueAdd,
      data: {
        'file_path': filePath,
        'filename': filename,
        'file_size_bytes': fileSizeBytes,
        'title': title ?? filename,
        'description': description ?? '',
      },
    );
    return QueueItemModel.fromJson(
      response.data['data'] as Map<String, dynamic>,
    );
  }

  Future<List<QueueItemModel>> getQueue({int page = 1, int limit = 50}) async {
    final response = await _apiClient.get(
      ApiConstants.queueList,
      queryParameters: {'page': page, 'limit': limit},
    );
    final data = response.data['data'] as Map<String, dynamic>;
    final items = data['items'] as List<dynamic>;
    return items
        .map((i) => QueueItemModel.fromJson(i as Map<String, dynamic>))
        .toList();
  }

  Future<void> removeFromQueue(String queueId) async {
    await _apiClient.delete('${ApiConstants.queueRemove}/$queueId');
  }

  Future<QuotaModel> getQuota() async {
    final response = await _apiClient.get(ApiConstants.queueQuota);
    return QuotaModel.fromJson(
      response.data['data'] as Map<String, dynamic>,
    );
  }
}
