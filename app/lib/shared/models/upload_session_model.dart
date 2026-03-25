class UploadSessionModel {
  final String sessionId;
  final String userId;
  final int totalFiles;
  final int completedFiles;
  final int failedFiles;
  final int totalBytes;
  final int uploadedBytes;
  final String sessionStatus;
  final DateTime startedAt;
  final DateTime? completedAt;

  UploadSessionModel({
    required this.sessionId,
    required this.userId,
    required this.totalFiles,
    required this.completedFiles,
    required this.failedFiles,
    required this.totalBytes,
    required this.uploadedBytes,
    required this.sessionStatus,
    required this.startedAt,
    this.completedAt,
  });

  factory UploadSessionModel.fromJson(Map<String, dynamic> json) {
    return UploadSessionModel(
      sessionId: json['session_id'] as String,
      userId: json['user_id'] as String,
      totalFiles: json['total_files'] as int,
      completedFiles: json['completed_files'] as int,
      failedFiles: json['failed_files'] as int,
      totalBytes: json['total_bytes'] as int,
      uploadedBytes: json['uploaded_bytes'] as int,
      sessionStatus: json['session_status'] as String,
      startedAt: DateTime.parse(json['started_at'] as String),
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
    );
  }

  double get progress =>
      totalBytes > 0 ? (uploadedBytes / totalBytes) * 100 : 0;

  int get pendingFiles => totalFiles - completedFiles - failedFiles;
  bool get isActive => sessionStatus == 'ACTIVE';
  bool get isCompleted => sessionStatus == 'COMPLETED';
}
