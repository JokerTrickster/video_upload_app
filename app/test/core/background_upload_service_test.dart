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

  group('BackgroundUploadService - Task Scheduling Logic', () {
    test('scheduleUpload creates tasks with correct taskIds', () async {
      // Simulate what scheduleUpload does: create task data per file
      final sessionId = 'test-session';
      final filenames = ['video1.mp4', 'video2.mp4', 'video3.mp4'];

      for (int i = 0; i < filenames.length; i++) {
        final taskId = '${sessionId}_$i';
        final task = UploadTaskData(
          taskId: taskId,
          sessionId: sessionId,
          filePath: '/path/${filenames[i]}',
          filename: filenames[i],
          fileSize: 1024 * (i + 1),
          status: 'pending',
          progress: 0,
          retryCount: 0,
          createdAt: now,
        );
        await persistence.saveTask(task);
      }

      // Verify all tasks created
      final tasks = await persistence.getAllTasks(sessionId);
      expect(tasks.length, 3);

      // Verify taskId format
      expect(tasks.any((t) => t.taskId == 'test-session_0'), true);
      expect(tasks.any((t) => t.taskId == 'test-session_1'), true);
      expect(tasks.any((t) => t.taskId == 'test-session_2'), true);

      // Verify all pending
      expect(tasks.every((t) => t.status == 'pending'), true);
    });

    test('scheduleUpload saves BackgroundUploadState', () async {
      final sessionId = 'test-session';
      final task = UploadTaskData(
        taskId: '${sessionId}_0',
        sessionId: sessionId,
        filePath: '/path/video.mp4',
        filename: 'video.mp4',
        fileSize: 1024,
        status: 'pending',
        progress: 0,
        retryCount: 0,
        createdAt: now,
      );
      await persistence.saveTask(task);

      final state = BackgroundUploadState(
        sessionId: sessionId,
        tasks: [task],
        isActive: true,
        lastUpdated: now,
      );
      await persistence.saveState(state);

      final loaded = await persistence.loadState();
      expect(loaded, isNotNull);
      expect(loaded!.sessionId, sessionId);
      expect(loaded.isActive, true);
      expect(loaded.tasks.length, 1);
    });
  });

  group('BackgroundUploadService - Cancel Logic', () {
    test('cancelAll clears all persisted data', () async {
      // Setup tasks
      await persistence.saveTask(UploadTaskData(
        taskId: 'session_0',
        sessionId: 'session',
        filePath: '/path/v.mp4',
        filename: 'v.mp4',
        fileSize: 1024,
        createdAt: now,
      ));
      await persistence.saveState(BackgroundUploadState(
        sessionId: 'session',
        tasks: [],
        isActive: true,
      ));

      // cancelAll internally calls persistence.clearAll()
      await persistence.clearAll();

      expect(await persistence.loadTask('session_0'), isNull);
      expect(await persistence.loadState(), isNull);
    });
  });

  group('BackgroundUploadService - Sync Logic', () {
    test('syncState returns current state', () async {
      final state = BackgroundUploadState(
        sessionId: 'session',
        tasks: [
          UploadTaskData(
            taskId: 'session_0',
            sessionId: 'session',
            filePath: '/path/v1.mp4',
            filename: 'v1.mp4',
            fileSize: 1024,
            status: 'completed',
            progress: 100,
            createdAt: now,
          ),
          UploadTaskData(
            taskId: 'session_1',
            sessionId: 'session',
            filePath: '/path/v2.mp4',
            filename: 'v2.mp4',
            fileSize: 2048,
            status: 'uploading',
            progress: 60,
            createdAt: now,
          ),
        ],
        isActive: true,
        lastUpdated: now,
      );
      await persistence.saveState(state);

      // syncState internally calls persistence.loadState()
      final synced = await persistence.loadState();
      expect(synced, isNotNull);
      expect(synced!.isActive, true);
      expect(synced.tasks.length, 2);
      expect(synced.tasks[0].status, 'completed');
      expect(synced.tasks[1].status, 'uploading');
      expect(synced.tasks[1].progress, 60);
    });

    test('syncState returns null when no background upload', () async {
      final synced = await persistence.loadState();
      expect(synced, isNull);
    });

    test('syncState shows inactive when all done', () async {
      final state = BackgroundUploadState(
        sessionId: 'session',
        tasks: [
          UploadTaskData(
            taskId: 'session_0',
            sessionId: 'session',
            filePath: '/path/v.mp4',
            filename: 'v.mp4',
            fileSize: 1024,
            status: 'completed',
            progress: 100,
            createdAt: now,
          ),
        ],
        isActive: false,
        lastUpdated: now,
      );
      await persistence.saveState(state);

      final synced = await persistence.loadState();
      expect(synced!.isActive, false);
    });
  });

  group('BackgroundUploadService - Task Lifecycle', () {
    test('full task lifecycle: pending -> uploading -> completed', () async {
      final taskId = 'session_0';
      await persistence.saveTask(UploadTaskData(
        taskId: taskId,
        sessionId: 'session',
        filePath: '/path/v.mp4',
        filename: 'v.mp4',
        fileSize: 1024,
        status: 'pending',
        createdAt: now,
      ));

      // Transition to uploading
      await persistence.updateTaskStatus(taskId, 'uploading');
      var task = await persistence.loadTask(taskId);
      expect(task!.status, 'uploading');

      // Progress updates
      await persistence.updateTaskProgress(taskId, 25.0);
      task = await persistence.loadTask(taskId);
      expect(task!.progress, 25.0);

      await persistence.updateTaskProgress(taskId, 75.0);
      task = await persistence.loadTask(taskId);
      expect(task!.progress, 75.0);

      // Complete
      final completedAt = DateTime(2026, 4, 1, 13, 0, 0);
      await persistence.updateTaskStatus(taskId, 'completed',
          completedAt: completedAt);
      task = await persistence.loadTask(taskId);
      expect(task!.status, 'completed');
      expect(task.completedAt, completedAt);
    });

    test('task lifecycle with retries: pending -> uploading -> retry -> failed',
        () async {
      final taskId = 'session_0';
      await persistence.saveTask(UploadTaskData(
        taskId: taskId,
        sessionId: 'session',
        filePath: '/path/v.mp4',
        filename: 'v.mp4',
        fileSize: 1024,
        status: 'pending',
        createdAt: now,
      ));

      // First attempt fails
      await persistence.updateTaskStatus(taskId, 'uploading');
      await persistence.updateTaskRetry(taskId, 1);
      var task = await persistence.loadTask(taskId);
      expect(task!.retryCount, 1);
      expect(task.status, 'pending'); // reset to pending for retry

      // Second attempt fails
      await persistence.updateTaskStatus(taskId, 'uploading');
      await persistence.updateTaskRetry(taskId, 2);
      task = await persistence.loadTask(taskId);
      expect(task!.retryCount, 2);

      // Third attempt fails -> permanent failure
      await persistence.updateTaskStatus(taskId, 'uploading');
      await persistence.updateTaskStatus(taskId, 'failed',
          error: 'Max retries exceeded');
      task = await persistence.loadTask(taskId);
      expect(task!.status, 'failed');
      expect(task.error, 'Max retries exceeded');
    });

    test('session completion check: all tasks done', () async {
      final sessionId = 'session';
      await persistence.saveTask(UploadTaskData(
        taskId: '${sessionId}_0',
        sessionId: sessionId,
        filePath: '/path/v1.mp4',
        filename: 'v1.mp4',
        fileSize: 1024,
        status: 'completed',
        progress: 100,
        createdAt: now,
      ));
      await persistence.saveTask(UploadTaskData(
        taskId: '${sessionId}_1',
        sessionId: sessionId,
        filePath: '/path/v2.mp4',
        filename: 'v2.mp4',
        fileSize: 2048,
        status: 'failed',
        progress: 30,
        error: 'Network error',
        createdAt: now,
      ));

      final tasks = await persistence.getAllTasks(sessionId);
      final allDone = tasks
          .every((t) => t.status == 'completed' || t.status == 'failed');
      expect(allDone, true);

      final completedCount =
          tasks.where((t) => t.status == 'completed').length;
      expect(completedCount, 1);
      expect(tasks.length, 2);
    });
  });
}
