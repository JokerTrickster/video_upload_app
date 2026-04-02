import 'package:flutter/foundation.dart';
import '../../../core/background/background_upload_service.dart';
import '../../../core/storage/settings_storage.dart';
import '../../../shared/models/upload_session_model.dart';
import '../data/upload_repository.dart';

class UploadFile {
  final String path;
  final String filename;
  final int size;
  String status; // 'pending', 'uploading', 'completed', 'failed'
  double progress;
  String? error;

  UploadFile({
    required this.path,
    required this.filename,
    required this.size,
    this.status = 'pending',
    this.progress = 0,
    this.error,
  });
}

class UploadProvider extends ChangeNotifier {
  final UploadRepository _uploadRepository;
  final BackgroundUploadService _backgroundService;

  final List<UploadFile> _files = [];
  UploadSessionModel? _session;
  bool _isUploading = false;
  String? _error;

  UploadProvider(this._uploadRepository)
      : _backgroundService = BackgroundUploadService();

  List<UploadFile> get files => _files;
  UploadSessionModel? get session => _session;
  bool get isUploading => _isUploading;
  String? get error => _error;
  int get completedCount => _files.where((f) => f.status == 'completed').length;
  int get failedCount => _files.where((f) => f.status == 'failed').length;

  double get overallProgress {
    if (_files.isEmpty) return 0;
    final totalProgress = _files.fold<double>(0, (sum, f) => sum + f.progress);
    return totalProgress / _files.length;
  }

  void addFiles(List<UploadFile> newFiles) {
    _files.addAll(newFiles);
    notifyListeners();
  }

  void removeFile(int index) {
    _files.removeAt(index);
    notifyListeners();
  }

  void clearFiles() {
    _files.removeRange(0, _files.length);
    _session = null;
    _error = null;
    notifyListeners();
  }

  Future<void> startUpload() async {
    if (_files.isEmpty) return;

    _isUploading = true;
    _error = null;
    notifyListeners();

    try {
      // Create upload session (always in foreground)
      final totalBytes = _files.fold<int>(0, (sum, f) => sum + f.size);
      _session = await _uploadRepository.initiateSession(
        totalFiles: _files.length,
        totalBytes: totalBytes,
      );
      notifyListeners();

      if (SettingsStorage.instance.isBackgroundUploadEnabled) {
        // Schedule background upload tasks
        await _backgroundService.scheduleUpload(
          sessionId: _session!.sessionId,
          files: _files,
        );
        for (final file in _files) {
          file.status = 'uploading';
        }
        notifyListeners();
      } else {
        // Fallback: foreground upload
        await _uploadForeground();
      }
    } catch (e) {
      _error = 'Upload failed: $e';
      _isUploading = false;
      notifyListeners();
    }
  }

  /// Foreground upload fallback (original logic)
  Future<void> _uploadForeground() async {
    for (int i = 0; i < _files.length; i++) {
      final file = _files[i];
      if (file.status == 'completed') continue;

      file.status = 'uploading';
      notifyListeners();

      try {
        await _uploadRepository.uploadVideo(
          sessionId: _session!.sessionId,
          filePath: file.path,
          filename: file.filename,
          fileSize: file.size,
          onProgress: (sent, total) {
            file.progress = total > 0 ? (sent / total) * 100 : 0;
            notifyListeners();
          },
        );
        file.status = 'completed';
        file.progress = 100;
      } catch (e) {
        file.status = 'failed';
        file.error = e.toString();
      }
      notifyListeners();
    }

    // Complete session
    try {
      await _uploadRepository.completeSession(_session!.sessionId);
    } catch (_) {}

    _isUploading = false;
    notifyListeners();
  }

  /// Sync state from background tasks when app returns to foreground
  Future<void> syncFromBackground() async {
    final state = await _backgroundService.syncState();
    if (state == null) return;

    if (!state.isActive) {
      _isUploading = false;
      notifyListeners();
      return;
    }

    // Reflect background task statuses in UI
    for (final task in state.tasks) {
      final fileIndex = _files.indexWhere((f) => f.path == task.filePath);
      if (fileIndex >= 0) {
        _files[fileIndex].status = task.status;
        _files[fileIndex].progress = task.progress;
        _files[fileIndex].error = task.error;
      }
    }

    _isUploading = state.tasks
        .any((t) => t.status == 'pending' || t.status == 'uploading');
    notifyListeners();
  }

  Future<void> cancelUpload() async {
    if (SettingsStorage.instance.isBackgroundUploadEnabled) {
      await _backgroundService.cancelAll();
    }
    if (_session != null) {
      try {
        await _uploadRepository.cancelSession(_session!.sessionId);
      } catch (_) {}
    }
    _isUploading = false;
    notifyListeners();
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }
}
