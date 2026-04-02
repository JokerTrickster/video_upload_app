import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import '../../shared/models/upload_task_data.dart';

class UploadStatePersistence {
  static const _keyPrefix = 'bg_upload_';
  static const _keyState = '${_keyPrefix}state';
  static const _keyTaskPrefix = '${_keyPrefix}task_';

  Future<void> saveTask(UploadTaskData task) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(
        '$_keyTaskPrefix${task.taskId}', jsonEncode(task.toJson()));
  }

  Future<UploadTaskData?> loadTask(String taskId) async {
    final prefs = await SharedPreferences.getInstance();
    final json = prefs.getString('$_keyTaskPrefix$taskId');
    if (json == null) return null;
    return UploadTaskData.fromJson(
        jsonDecode(json) as Map<String, dynamic>);
  }

  Future<List<UploadTaskData>> getAllTasks(String sessionId) async {
    final prefs = await SharedPreferences.getInstance();
    final keys =
        prefs.getKeys().where((k) => k.startsWith(_keyTaskPrefix)).toList();
    final tasks = <UploadTaskData>[];

    for (final key in keys) {
      final json = prefs.getString(key);
      if (json != null) {
        final task = UploadTaskData.fromJson(
            jsonDecode(json) as Map<String, dynamic>);
        if (task.sessionId == sessionId) {
          tasks.add(task);
        }
      }
    }
    return tasks;
  }

  Future<void> updateTaskStatus(
    String taskId,
    String status, {
    DateTime? completedAt,
    String? error,
  }) async {
    final task = await loadTask(taskId);
    if (task == null) return;

    final updated = task.copyWith(
      status: status,
      completedAt: completedAt,
      error: error,
    );
    await saveTask(updated);
  }

  Future<void> updateTaskProgress(String taskId, double progress) async {
    final task = await loadTask(taskId);
    if (task == null) return;
    await saveTask(task.copyWith(progress: progress));
  }

  Future<void> updateTaskRetry(String taskId, int retryCount) async {
    final task = await loadTask(taskId);
    if (task == null) return;
    await saveTask(task.copyWith(retryCount: retryCount, status: 'pending'));
  }

  Future<void> saveState(BackgroundUploadState state) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyState, jsonEncode(state.toJson()));
  }

  Future<BackgroundUploadState?> loadState() async {
    final prefs = await SharedPreferences.getInstance();
    final json = prefs.getString(_keyState);
    if (json == null) return null;
    return BackgroundUploadState.fromJson(
        jsonDecode(json) as Map<String, dynamic>);
  }

  Future<void> updateStateActive(String sessionId, bool isActive) async {
    final state = await loadState();
    if (state == null || state.sessionId != sessionId) return;
    await saveState(BackgroundUploadState(
      sessionId: sessionId,
      tasks: await getAllTasks(sessionId),
      isActive: isActive,
      lastUpdated: DateTime.now(),
    ));
  }

  Future<void> clearAll() async {
    final prefs = await SharedPreferences.getInstance();
    final keys =
        prefs.getKeys().where((k) => k.startsWith(_keyPrefix)).toList();
    for (final key in keys) {
      await prefs.remove(key);
    }
  }
}
