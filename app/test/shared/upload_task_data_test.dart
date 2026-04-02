import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:video_upload_app/shared/models/upload_task_data.dart';

void main() {
  group('UploadTaskData', () {
    late UploadTaskData task;
    late DateTime now;

    setUp(() {
      now = DateTime(2026, 4, 1, 12, 0, 0);
      task = UploadTaskData(
        taskId: 'session1_0',
        sessionId: 'session1',
        filePath: '/path/to/video.mp4',
        filename: 'video.mp4',
        fileSize: 1024000,
        status: 'pending',
        progress: 0,
        retryCount: 0,
        createdAt: now,
      );
    });

    test('creates with correct default values', () {
      expect(task.taskId, 'session1_0');
      expect(task.sessionId, 'session1');
      expect(task.filePath, '/path/to/video.mp4');
      expect(task.filename, 'video.mp4');
      expect(task.fileSize, 1024000);
      expect(task.status, 'pending');
      expect(task.progress, 0);
      expect(task.error, isNull);
      expect(task.retryCount, 0);
      expect(task.createdAt, now);
      expect(task.completedAt, isNull);
    });

    test('copyWith updates only specified fields', () {
      final updated = task.copyWith(
        status: 'uploading',
        progress: 50.0,
      );

      expect(updated.status, 'uploading');
      expect(updated.progress, 50.0);
      expect(updated.taskId, task.taskId);
      expect(updated.sessionId, task.sessionId);
      expect(updated.filePath, task.filePath);
      expect(updated.filename, task.filename);
      expect(updated.fileSize, task.fileSize);
      expect(updated.retryCount, 0);
    });

    test('copyWith preserves all fields when no arguments given', () {
      final copy = task.copyWith();

      expect(copy.taskId, task.taskId);
      expect(copy.sessionId, task.sessionId);
      expect(copy.filePath, task.filePath);
      expect(copy.filename, task.filename);
      expect(copy.fileSize, task.fileSize);
      expect(copy.status, task.status);
      expect(copy.progress, task.progress);
      expect(copy.error, task.error);
      expect(copy.retryCount, task.retryCount);
      expect(copy.createdAt, task.createdAt);
      expect(copy.completedAt, task.completedAt);
    });

    test('copyWith with completedAt and error', () {
      final completedAt = DateTime(2026, 4, 1, 12, 30, 0);
      final updated = task.copyWith(
        status: 'failed',
        error: 'Network timeout',
        completedAt: completedAt,
        retryCount: 3,
      );

      expect(updated.status, 'failed');
      expect(updated.error, 'Network timeout');
      expect(updated.completedAt, completedAt);
      expect(updated.retryCount, 3);
    });

    test('toJson serializes all fields correctly', () {
      final json = task.toJson();

      expect(json['taskId'], 'session1_0');
      expect(json['sessionId'], 'session1');
      expect(json['filePath'], '/path/to/video.mp4');
      expect(json['filename'], 'video.mp4');
      expect(json['fileSize'], 1024000);
      expect(json['status'], 'pending');
      expect(json['progress'], 0);
      expect(json['error'], isNull);
      expect(json['retryCount'], 0);
      expect(json['createdAt'], now.toIso8601String());
      expect(json['completedAt'], isNull);
    });

    test('fromJson deserializes all fields correctly', () {
      final json = task.toJson();
      final restored = UploadTaskData.fromJson(json);

      expect(restored.taskId, task.taskId);
      expect(restored.sessionId, task.sessionId);
      expect(restored.filePath, task.filePath);
      expect(restored.filename, task.filename);
      expect(restored.fileSize, task.fileSize);
      expect(restored.status, task.status);
      expect(restored.progress, task.progress);
      expect(restored.error, task.error);
      expect(restored.retryCount, task.retryCount);
      expect(restored.createdAt, task.createdAt);
      expect(restored.completedAt, task.completedAt);
    });

    test('fromJson handles optional completedAt', () {
      final completedAt = DateTime(2026, 4, 1, 13, 0, 0);
      final taskWithCompletion = task.copyWith(
        status: 'completed',
        completedAt: completedAt,
      );
      final json = taskWithCompletion.toJson();
      final restored = UploadTaskData.fromJson(json);

      expect(restored.completedAt, completedAt);
      expect(restored.status, 'completed');
    });

    test('fromJson handles missing optional fields with defaults', () {
      final minimalJson = {
        'taskId': 'test_0',
        'sessionId': 'test',
        'filePath': '/path/file.mp4',
        'filename': 'file.mp4',
        'fileSize': 500,
        'createdAt': now.toIso8601String(),
      };

      final restored = UploadTaskData.fromJson(minimalJson);
      expect(restored.status, 'pending');
      expect(restored.progress, 0);
      expect(restored.retryCount, 0);
      expect(restored.error, isNull);
      expect(restored.completedAt, isNull);
    });

    test('toJsonString and fromJsonString roundtrip', () {
      final jsonString = task.toJsonString();
      final restored = UploadTaskData.fromJsonString(jsonString);

      expect(restored.taskId, task.taskId);
      expect(restored.sessionId, task.sessionId);
      expect(restored.filename, task.filename);
    });

    test('JSON string is valid JSON', () {
      final jsonString = task.toJsonString();
      expect(() => jsonDecode(jsonString), returnsNormally);
    });
  });

  group('BackgroundUploadState', () {
    late DateTime now;
    late List<UploadTaskData> tasks;

    setUp(() {
      now = DateTime(2026, 4, 1, 12, 0, 0);
      tasks = [
        UploadTaskData(
          taskId: 'session1_0',
          sessionId: 'session1',
          filePath: '/path/video1.mp4',
          filename: 'video1.mp4',
          fileSize: 1000,
          status: 'completed',
          progress: 100,
          createdAt: now,
        ),
        UploadTaskData(
          taskId: 'session1_1',
          sessionId: 'session1',
          filePath: '/path/video2.mp4',
          filename: 'video2.mp4',
          fileSize: 2000,
          status: 'uploading',
          progress: 50,
          createdAt: now,
        ),
      ];
    });

    test('creates with correct values', () {
      final state = BackgroundUploadState(
        sessionId: 'session1',
        tasks: tasks,
        isActive: true,
        lastUpdated: now,
      );

      expect(state.sessionId, 'session1');
      expect(state.tasks.length, 2);
      expect(state.isActive, true);
      expect(state.lastUpdated, now);
    });

    test('default isActive is true', () {
      final state = BackgroundUploadState(
        sessionId: 'session1',
        tasks: tasks,
      );
      expect(state.isActive, true);
    });

    test('toJson serializes tasks array', () {
      final state = BackgroundUploadState(
        sessionId: 'session1',
        tasks: tasks,
        isActive: true,
        lastUpdated: now,
      );
      final json = state.toJson();

      expect(json['sessionId'], 'session1');
      expect(json['isActive'], true);
      expect(json['tasks'], isList);
      expect((json['tasks'] as List).length, 2);
      expect(json['lastUpdated'], now.toIso8601String());
    });

    test('fromJson restores state with tasks', () {
      final state = BackgroundUploadState(
        sessionId: 'session1',
        tasks: tasks,
        isActive: true,
        lastUpdated: now,
      );
      final json = state.toJson();
      final restored = BackgroundUploadState.fromJson(json);

      expect(restored.sessionId, 'session1');
      expect(restored.isActive, true);
      expect(restored.tasks.length, 2);
      expect(restored.tasks[0].taskId, 'session1_0');
      expect(restored.tasks[0].status, 'completed');
      expect(restored.tasks[1].taskId, 'session1_1');
      expect(restored.tasks[1].status, 'uploading');
      expect(restored.lastUpdated, now);
    });

    test('fromJson handles null lastUpdated', () {
      final json = {
        'sessionId': 'session1',
        'tasks': [],
        'isActive': false,
      };
      final restored = BackgroundUploadState.fromJson(json);

      expect(restored.lastUpdated, isNull);
      expect(restored.isActive, false);
      expect(restored.tasks, isEmpty);
    });

    test('JSON roundtrip preserves all data', () {
      final state = BackgroundUploadState(
        sessionId: 'session1',
        tasks: tasks,
        isActive: true,
        lastUpdated: now,
      );

      final jsonString = jsonEncode(state.toJson());
      final restoredJson = jsonDecode(jsonString) as Map<String, dynamic>;
      final restored = BackgroundUploadState.fromJson(restoredJson);

      expect(restored.sessionId, state.sessionId);
      expect(restored.isActive, state.isActive);
      expect(restored.tasks.length, state.tasks.length);
      for (int i = 0; i < state.tasks.length; i++) {
        expect(restored.tasks[i].taskId, state.tasks[i].taskId);
        expect(restored.tasks[i].status, state.tasks[i].status);
        expect(restored.tasks[i].progress, state.tasks[i].progress);
      }
    });
  });
}
