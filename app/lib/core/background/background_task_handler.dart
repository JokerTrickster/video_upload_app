import 'package:workmanager/workmanager.dart';
import '../notifications/notification_service.dart';
import 'background_api_client.dart';
import 'upload_state_persistence.dart';

const String backgroundUploadTaskName = 'com.app.backgroundUpload';

@pragma('vm:entry-point')
void callbackDispatcher() {
  Workmanager().executeTask((taskName, inputData) async {
    if (taskName != backgroundUploadTaskName) {
      return Future.value(true);
    }

    final taskId = inputData?['taskId'] as String?;
    if (taskId == null) return Future.value(false);

    final handler = BackgroundTaskHandler();
    return await handler.execute(taskId);
  });
}

class BackgroundTaskHandler {
  final BackgroundApiClient _apiClient = BackgroundApiClient();
  final UploadStatePersistence _persistence = UploadStatePersistence();

  Future<bool> execute(String taskId) async {
    final task = await _persistence.loadTask(taskId);
    if (task == null) return true;

    await _persistence.updateTaskStatus(taskId, 'uploading');

    try {
      await _apiClient.initialize();

      await _apiClient.uploadVideo(
        sessionId: task.sessionId,
        filePath: task.filePath,
        filename: task.filename,
        fileSize: task.fileSize,
        onProgress: (sent, total) async {
          final progress = total > 0 ? (sent / total) * 100 : 0.0;
          await _persistence.updateTaskProgress(taskId, progress);
        },
      );

      await _persistence.updateTaskStatus(taskId, 'completed',
          completedAt: DateTime.now());

      await NotificationService().init();
      await NotificationService().showUploadComplete(task.filename);

      await _checkAndCompleteSession(task.sessionId);

      return true;
    } catch (e) {
      final newRetryCount = task.retryCount + 1;
      if (newRetryCount >= 3) {
        await _persistence.updateTaskStatus(taskId, 'failed',
            error: e.toString());
        await NotificationService().init();
        await NotificationService()
            .showUploadFailed(task.filename, e.toString());
      } else {
        await _persistence.updateTaskRetry(taskId, newRetryCount);
        return false; // workmanager auto-retries
      }
      return true;
    }
  }

  Future<void> _checkAndCompleteSession(String sessionId) async {
    final tasks = await _persistence.getAllTasks(sessionId);
    final allDone = tasks
        .every((t) => t.status == 'completed' || t.status == 'failed');

    if (allDone) {
      final completedCount =
          tasks.where((t) => t.status == 'completed').length;
      final totalCount = tasks.length;

      try {
        await _apiClient.completeSession(sessionId);
      } catch (_) {}

      await _persistence.updateStateActive(sessionId, false);

      if (completedCount == totalCount) {
        await NotificationService().init();
        await NotificationService()
            .showUploadComplete('$completedCount files uploaded successfully');
      }
    }
  }
}
