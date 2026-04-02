import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:video_upload_app/core/background/upload_state_persistence.dart';
import 'package:video_upload_app/shared/models/upload_task_data.dart';

void main() {
  late UploadStatePersistence persistence;
  late DateTime now;

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    persistence = UploadStatePersistence();
    now = DateTime(2026, 4, 1, 12, 0, 0);
  });

  UploadTaskData createTask({
    String taskId = 'session1_0',
    String sessionId = 'session1',
    String status = 'pending',
    double progress = 0,
    int retryCount = 0,
    String? error,
  }) {
    return UploadTaskData(
      taskId: taskId,
      sessionId: sessionId,
      filePath: '/path/$taskId.mp4',
      filename: '$taskId.mp4',
      fileSize: 1024,
      status: status,
      progress: progress,
      retryCount: retryCount,
      error: error,
      createdAt: now,
    );
  }

  group('UploadStatePersistence - Task CRUD', () {
    test('saveTask and loadTask roundtrip', () async {
      final task = createTask();
      await persistence.saveTask(task);

      final loaded = await persistence.loadTask('session1_0');
      expect(loaded, isNotNull);
      expect(loaded!.taskId, 'session1_0');
      expect(loaded.sessionId, 'session1');
      expect(loaded.filePath, '/path/session1_0.mp4');
      expect(loaded.filename, 'session1_0.mp4');
      expect(loaded.fileSize, 1024);
      expect(loaded.status, 'pending');
      expect(loaded.progress, 0);
      expect(loaded.retryCount, 0);
    });

    test('loadTask returns null for non-existent task', () async {
      final loaded = await persistence.loadTask('nonexistent');
      expect(loaded, isNull);
    });

    test('saveTask overwrites existing task', () async {
      final task = createTask(status: 'pending');
      await persistence.saveTask(task);

      final updated = task.copyWith(status: 'uploading', progress: 50.0);
      await persistence.saveTask(updated);

      final loaded = await persistence.loadTask('session1_0');
      expect(loaded!.status, 'uploading');
      expect(loaded.progress, 50.0);
    });

    test('getAllTasks filters by sessionId', () async {
      await persistence.saveTask(
          createTask(taskId: 's1_0', sessionId: 'session1'));
      await persistence.saveTask(
          createTask(taskId: 's1_1', sessionId: 'session1'));
      await persistence.saveTask(
          createTask(taskId: 's2_0', sessionId: 'session2'));

      final session1Tasks = await persistence.getAllTasks('session1');
      expect(session1Tasks.length, 2);
      expect(session1Tasks.every((t) => t.sessionId == 'session1'), true);

      final session2Tasks = await persistence.getAllTasks('session2');
      expect(session2Tasks.length, 1);
      expect(session2Tasks.first.sessionId, 'session2');
    });

    test('getAllTasks returns empty list for unknown session', () async {
      await persistence
          .saveTask(createTask(taskId: 's1_0', sessionId: 'session1'));

      final tasks = await persistence.getAllTasks('unknown');
      expect(tasks, isEmpty);
    });
  });

  group('UploadStatePersistence - Status Updates', () {
    test('updateTaskStatus changes status', () async {
      await persistence.saveTask(createTask(status: 'pending'));

      await persistence.updateTaskStatus('session1_0', 'uploading');
      final loaded = await persistence.loadTask('session1_0');
      expect(loaded!.status, 'uploading');
    });

    test('updateTaskStatus with completedAt', () async {
      await persistence.saveTask(createTask());
      final completedAt = DateTime(2026, 4, 1, 13, 0, 0);

      await persistence.updateTaskStatus('session1_0', 'completed',
          completedAt: completedAt);
      final loaded = await persistence.loadTask('session1_0');
      expect(loaded!.status, 'completed');
      expect(loaded.completedAt, completedAt);
    });

    test('updateTaskStatus with error', () async {
      await persistence.saveTask(createTask());

      await persistence.updateTaskStatus('session1_0', 'failed',
          error: 'Network timeout');
      final loaded = await persistence.loadTask('session1_0');
      expect(loaded!.status, 'failed');
      expect(loaded.error, 'Network timeout');
    });

    test('updateTaskStatus does nothing for non-existent task', () async {
      // Should not throw
      await persistence.updateTaskStatus('nonexistent', 'uploading');
    });

    test('updateTaskProgress updates progress value', () async {
      await persistence.saveTask(createTask(progress: 0));

      await persistence.updateTaskProgress('session1_0', 75.5);
      final loaded = await persistence.loadTask('session1_0');
      expect(loaded!.progress, 75.5);
    });

    test('updateTaskProgress does nothing for non-existent task', () async {
      await persistence.updateTaskProgress('nonexistent', 50.0);
    });

    test('updateTaskRetry increments retry and resets to pending', () async {
      await persistence
          .saveTask(createTask(status: 'uploading', retryCount: 0));

      await persistence.updateTaskRetry('session1_0', 1);
      final loaded = await persistence.loadTask('session1_0');
      expect(loaded!.retryCount, 1);
      expect(loaded.status, 'pending');
    });

    test('updateTaskRetry does nothing for non-existent task', () async {
      await persistence.updateTaskRetry('nonexistent', 1);
    });
  });

  group('UploadStatePersistence - State Management', () {
    test('saveState and loadState roundtrip', () async {
      final tasks = [
        createTask(taskId: 's1_0', status: 'completed', progress: 100),
        createTask(taskId: 's1_1', status: 'uploading', progress: 50),
      ];
      final state = BackgroundUploadState(
        sessionId: 'session1',
        tasks: tasks,
        isActive: true,
        lastUpdated: now,
      );

      await persistence.saveState(state);
      final loaded = await persistence.loadState();

      expect(loaded, isNotNull);
      expect(loaded!.sessionId, 'session1');
      expect(loaded.isActive, true);
      expect(loaded.tasks.length, 2);
      expect(loaded.lastUpdated, now);
    });

    test('loadState returns null when no state saved', () async {
      final loaded = await persistence.loadState();
      expect(loaded, isNull);
    });

    test('updateStateActive changes isActive flag', () async {
      final task = createTask(taskId: 's1_0');
      await persistence.saveTask(task);

      final state = BackgroundUploadState(
        sessionId: 'session1',
        tasks: [task],
        isActive: true,
        lastUpdated: now,
      );
      await persistence.saveState(state);

      await persistence.updateStateActive('session1', false);
      final loaded = await persistence.loadState();
      expect(loaded!.isActive, false);
    });

    test('updateStateActive ignores mismatched sessionId', () async {
      final state = BackgroundUploadState(
        sessionId: 'session1',
        tasks: [],
        isActive: true,
        lastUpdated: now,
      );
      await persistence.saveState(state);

      await persistence.updateStateActive('session2', false);
      final loaded = await persistence.loadState();
      expect(loaded!.isActive, true); // unchanged
    });
  });

  group('UploadStatePersistence - clearAll', () {
    test('clearAll removes all background upload data', () async {
      await persistence.saveTask(createTask(taskId: 's1_0'));
      await persistence.saveTask(createTask(taskId: 's1_1'));
      await persistence.saveState(BackgroundUploadState(
        sessionId: 'session1',
        tasks: [],
        isActive: true,
      ));

      await persistence.clearAll();

      expect(await persistence.loadTask('s1_0'), isNull);
      expect(await persistence.loadTask('s1_1'), isNull);
      expect(await persistence.loadState(), isNull);
    });

    test('clearAll does not affect non-background keys', () async {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setBool('auto_upload_enabled', true);

      await persistence.saveTask(createTask());
      await persistence.clearAll();

      expect(prefs.getBool('auto_upload_enabled'), true);
    });
  });
}
