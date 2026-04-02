import 'dart:convert';

class UploadTaskData {
  final String taskId;
  final String sessionId;
  final String filePath;
  final String filename;
  final int fileSize;
  final String status; // pending | uploading | completed | failed
  final double progress;
  final String? error;
  final int retryCount;
  final DateTime createdAt;
  final DateTime? completedAt;

  UploadTaskData({
    required this.taskId,
    required this.sessionId,
    required this.filePath,
    required this.filename,
    required this.fileSize,
    this.status = 'pending',
    this.progress = 0,
    this.error,
    this.retryCount = 0,
    required this.createdAt,
    this.completedAt,
  });

  UploadTaskData copyWith({
    String? taskId,
    String? sessionId,
    String? filePath,
    String? filename,
    int? fileSize,
    String? status,
    double? progress,
    String? error,
    int? retryCount,
    DateTime? createdAt,
    DateTime? completedAt,
  }) {
    return UploadTaskData(
      taskId: taskId ?? this.taskId,
      sessionId: sessionId ?? this.sessionId,
      filePath: filePath ?? this.filePath,
      filename: filename ?? this.filename,
      fileSize: fileSize ?? this.fileSize,
      status: status ?? this.status,
      progress: progress ?? this.progress,
      error: error ?? this.error,
      retryCount: retryCount ?? this.retryCount,
      createdAt: createdAt ?? this.createdAt,
      completedAt: completedAt ?? this.completedAt,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'taskId': taskId,
      'sessionId': sessionId,
      'filePath': filePath,
      'filename': filename,
      'fileSize': fileSize,
      'status': status,
      'progress': progress,
      'error': error,
      'retryCount': retryCount,
      'createdAt': createdAt.toIso8601String(),
      'completedAt': completedAt?.toIso8601String(),
    };
  }

  factory UploadTaskData.fromJson(Map<String, dynamic> json) {
    return UploadTaskData(
      taskId: json['taskId'] as String,
      sessionId: json['sessionId'] as String,
      filePath: json['filePath'] as String,
      filename: json['filename'] as String,
      fileSize: json['fileSize'] as int,
      status: json['status'] as String? ?? 'pending',
      progress: (json['progress'] as num?)?.toDouble() ?? 0,
      error: json['error'] as String?,
      retryCount: json['retryCount'] as int? ?? 0,
      createdAt: DateTime.parse(json['createdAt'] as String),
      completedAt: json['completedAt'] != null
          ? DateTime.parse(json['completedAt'] as String)
          : null,
    );
  }

  String toJsonString() => jsonEncode(toJson());

  factory UploadTaskData.fromJsonString(String jsonString) {
    return UploadTaskData.fromJson(
        jsonDecode(jsonString) as Map<String, dynamic>);
  }
}

class BackgroundUploadState {
  final String sessionId;
  final List<UploadTaskData> tasks;
  final bool isActive;
  final DateTime? lastUpdated;

  BackgroundUploadState({
    required this.sessionId,
    required this.tasks,
    this.isActive = true,
    this.lastUpdated,
  });

  Map<String, dynamic> toJson() {
    return {
      'sessionId': sessionId,
      'tasks': tasks.map((t) => t.toJson()).toList(),
      'isActive': isActive,
      'lastUpdated': lastUpdated?.toIso8601String(),
    };
  }

  factory BackgroundUploadState.fromJson(Map<String, dynamic> json) {
    return BackgroundUploadState(
      sessionId: json['sessionId'] as String,
      tasks: (json['tasks'] as List<dynamic>)
          .map((t) => UploadTaskData.fromJson(t as Map<String, dynamic>))
          .toList(),
      isActive: json['isActive'] as bool? ?? false,
      lastUpdated: json['lastUpdated'] != null
          ? DateTime.parse(json['lastUpdated'] as String)
          : null,
    );
  }
}
