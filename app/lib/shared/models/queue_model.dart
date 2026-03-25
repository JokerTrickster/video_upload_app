class QueueItemModel {
  final String queueId;
  final String filename;
  final int fileSizeBytes;
  final String title;
  final String status;
  final int retryCount;
  final String? errorMessage;
  final DateTime createdAt;
  final DateTime? processedAt;

  QueueItemModel({
    required this.queueId,
    required this.filename,
    required this.fileSizeBytes,
    required this.title,
    required this.status,
    required this.retryCount,
    this.errorMessage,
    required this.createdAt,
    this.processedAt,
  });

  factory QueueItemModel.fromJson(Map<String, dynamic> json) {
    return QueueItemModel(
      queueId: json['queue_id'] as String,
      filename: json['filename'] as String,
      fileSizeBytes: json['file_size_bytes'] as int? ?? 0,
      title: json['title'] as String? ?? '',
      status: json['status'] as String,
      retryCount: json['retry_count'] as int? ?? 0,
      errorMessage: json['error_message'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
      processedAt: json['processed_at'] != null
          ? DateTime.parse(json['processed_at'] as String)
          : null,
    );
  }

  bool get isPending => status == 'PENDING';
  bool get isProcessing => status == 'PROCESSING';
  bool get isCompleted => status == 'COMPLETED';
  bool get isFailed => status == 'FAILED';

  String get fileSizeFormatted {
    if (fileSizeBytes < 1024 * 1024) {
      return '${(fileSizeBytes / 1024).toStringAsFixed(1)} KB';
    }
    if (fileSizeBytes < 1024 * 1024 * 1024) {
      return '${(fileSizeBytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(fileSizeBytes / (1024 * 1024 * 1024)).toStringAsFixed(2)} GB';
  }
}

class QuotaModel {
  final String date;
  final int unitsUsed;
  final int unitsMax;
  final int uploadsToday;
  final int remainingUploads;
  final bool canUpload;

  QuotaModel({
    required this.date,
    required this.unitsUsed,
    required this.unitsMax,
    required this.uploadsToday,
    required this.remainingUploads,
    required this.canUpload,
  });

  factory QuotaModel.fromJson(Map<String, dynamic> json) {
    return QuotaModel(
      date: json['date'] as String,
      unitsUsed: json['units_used'] as int,
      unitsMax: json['units_max'] as int,
      uploadsToday: json['uploads_today'] as int,
      remainingUploads: json['remaining_uploads'] as int,
      canUpload: json['can_upload'] as bool,
    );
  }

  double get usagePercent => unitsMax > 0 ? (unitsUsed / unitsMax) * 100 : 0;
}
