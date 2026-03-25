import '../../../core/api/api_client.dart';
import '../../../core/constants/api_constants.dart';
import '../../../shared/models/media_asset_model.dart';

class MediaRepository {
  final ApiClient _apiClient;

  MediaRepository(this._apiClient);

  Future<MediaAssetListResponse> listAssets({
    int page = 1,
    int limit = 50,
    String? mediaType,
    String? syncStatus,
    String sort = 'created_at_desc',
  }) async {
    final queryParams = <String, dynamic>{
      'page': page,
      'limit': limit,
      'sort': sort,
    };
    if (mediaType != null) queryParams['media_type'] = mediaType;
    if (syncStatus != null) queryParams['sync_status'] = syncStatus;

    final response = await _apiClient.get(
      ApiConstants.mediaList,
      queryParameters: queryParams,
    );
    return MediaAssetListResponse.fromJson(
      response.data['data'] as Map<String, dynamic>,
    );
  }

  Future<MediaAssetModel> getAsset(String assetId) async {
    final response = await _apiClient.get(
      '${ApiConstants.mediaDetail}/$assetId',
    );
    return MediaAssetModel.fromJson(
      response.data['data'] as Map<String, dynamic>,
    );
  }

  Future<void> deleteAsset(String assetId) async {
    await _apiClient.delete('${ApiConstants.mediaDetail}/$assetId');
  }
}
