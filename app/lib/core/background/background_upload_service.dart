import 'package:workmanager/workmanager.dart';
import '../../shared/models/upload_task_data.dart';
import '../../features/upload/presentation/upload_provider.dart';
import '../storage/settings_storage.dart';
import 'background_task_handler.dart';
import 'upload_state_persistence.dart';

class BackgroundUploadService {
  static Future<void> initialize() async {
    await Workmanager().initialize(callbackDispatcher, isInDebugMode: false);
  }

  Future<void> scheduleUpload({
    required String sessionId,
    required List<UploadFile> files,
  }) async {
    final persistence = UploadStatePersistence();

    for (int i = 0; i < files.length; i++) {
      final file = files[i];
      final taskId = '${sessionId}_$i';

      final taskData = UploadTaskData(
        taskId: taskId,
        sessionId: sessionId,
        filePath: file.path,
        filename: file.filename,
        fileSize: file.size,
        status: 'pending',
        progress: 0,
        retryCount: 0,
        createdAt: DateTime.now(),
      );

      await persistence.saveTask(taskData);

      await Workmanager().registerOneOffTask(
        taskId,
        backgroundUploadTaskName,
        inputData: {'taskId': taskId},
        constraints: Constraints(
          networkType: SettingsStorage.instance.isWifiOnly
              ? NetworkType.unmetered
              : NetworkType.connected,
          requiresCharging: SettingsStorage.instance.isChargingOnly,
        ),
        backoffPolicy: BackoffPolicy.exponential,
        initialDelay: Duration.zero,
        existingWorkPolicy: ExistingWorkPolicy.keep,
      );
    }

    await persistence.saveState(BackgroundUploadState(
      sessionId: sessionId,
      tasks: await persistence.getAllTasks(sessionId),
      isActive: true,
      lastUpdated: DateTime.now(),
    ));
  }

  Future<void> cancelAll() async {
    await Workmanager().cancelAll();
    await UploadStatePersistence().clearAll();
  }

  Future<BackgroundUploadState?> syncState() async {
    return await UploadStatePersistence().loadState();
  }
}
