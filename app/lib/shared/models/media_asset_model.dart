class MediaAssetModel {
  final String assetId;
  final String? youtubeVideoId;
  final String? s3ObjectKey;
  final String originalFilename;
  final int fileSizeBytes;
  final String mediaType;
  final String syncStatus;
  final DateTime createdAt;
  final DateTime? uploadStartedAt;
  final DateTime? uploadCompletedAt;
  final String? errorMessage;
  final int retryCount;

  MediaAssetModel({
    required this.assetId,
    this.youtubeVideoId,
    this.s3ObjectKey,
    required this.originalFilename,
    required this.fileSizeBytes,
    required this.mediaType,
    required this.syncStatus,
    required this.createdAt,
    this.uploadStartedAt,
    this.uploadCompletedAt,
    this.errorMessage,
    required this.retryCount,
  });

  factory MediaAssetModel.fromJson(Map<String, dynamic> json) {
    return MediaAssetModel(
      assetId: json['asset_id'] as String,
      youtubeVideoId: json['youtube_video_id'] as String?,
      s3ObjectKey: json['s3_object_key'] as String?,
      originalFilename: json['original_filename'] as String,
      fileSizeBytes: json['file_size_bytes'] as int,
      mediaType: json['media_type'] as String,
      syncStatus: json['sync_status'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      uploadStartedAt: json['upload_started_at'] != null
          ? DateTime.parse(json['upload_started_at'] as String)
          : null,
      uploadCompletedAt: json['upload_completed_at'] != null
          ? DateTime.parse(json['upload_completed_at'] as String)
          : null,
      errorMessage: json['error_message'] as String?,
      retryCount: json['retry_count'] as int? ?? 0,
    );
  }

  String get fileSizeFormatted {
    if (fileSizeBytes < 1024) return '$fileSizeBytes B';
    if (fileSizeBytes < 1024 * 1024) {
      return '${(fileSizeBytes / 1024).toStringAsFixed(1)} KB';
    }
    if (fileSizeBytes < 1024 * 1024 * 1024) {
      return '${(fileSizeBytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(fileSizeBytes / (1024 * 1024 * 1024)).toStringAsFixed(2)} GB';
  }

  bool get isCompleted => syncStatus == 'COMPLETED';
  bool get isFailed => syncStatus == 'FAILED';
  bool get isUploading => syncStatus == 'UPLOADING';
  bool get isPending => syncStatus == 'PENDING';
}

class MediaAssetListResponse {
  final List<MediaAssetModel> assets;
  final int page;
  final int limit;
  final int total;
  final int totalPages;

  MediaAssetListResponse({
    required this.assets,
    required this.page,
    required this.limit,
    required this.total,
    required this.totalPages,
  });

  factory MediaAssetListResponse.fromJson(Map<String, dynamic> json) {
    final pagination = json['pagination'] as Map<String, dynamic>;
    final assetsList = json['assets'] as List<dynamic>;
    return MediaAssetListResponse(
      assets: assetsList
          .map((a) => MediaAssetModel.fromJson(a as Map<String, dynamic>))
          .toList(),
      page: pagination['page'] as int,
      limit: pagination['limit'] as int,
      total: pagination['total'] as int,
      totalPages: pagination['total_pages'] as int,
    );
  }
}
