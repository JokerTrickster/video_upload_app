# Design: 백그라운드 업로드

**Created**: 2026-04-01T07:38:59Z
**Architecture**: Option C — Pragmatic Balance
**Plan Reference**: `docs/01-plan/features/background-upload.plan.md`

---

## Context Anchor

| 항목 | 내용 |
|------|------|
| **WHY** | 대용량 영상 업로드 중 앱을 계속 열어두어야 하는 UX 제약 해소 |
| **WHO** | 일상적으로 고화질 영상을 촬영하고 백업하는 사용자 |
| **RISK** | iOS 백그라운드 실행 시간 제한(~30초 기본, BGProcessingTask로 수분 확보), 배터리 소모 증가 |
| **SUCCESS** | 백그라운드 업로드 성공률 95%+, 앱 전환 후에도 업로드 중단 없음 |
| **SCOPE** | 수동 업로드 + 자동 큐 업로드의 백그라운드 실행, 로컬 알림 연동 |

---

## 1. Overview

### 선택된 아키텍처: Option C — Pragmatic Balance

백그라운드 서비스 레이어를 분리하되 과도한 추상화를 피함. SharedPreferences 기반 상태 브릿지로 포그라운드↔백그라운드 간 상태 동기화.

```
┌──────────────────────────────────────────────────────┐
│  Foreground (Main Isolate)                            │
│                                                       │
│  UploadProvider ──→ BackgroundUploadService            │
│       ↑                    │                          │
│       │                    ↓                          │
│  syncFromBackground()  workmanager.registerOneOffTask │
│       ↑                    │                          │
│       │                    ↓                          │
│  ┌────┴────────────────────────────────┐              │
│  │  UploadStatePersistence             │              │
│  │  (SharedPreferences - 상태 브릿지)    │              │
│  └────┬────────────────────────────────┘              │
└───────┼──────────────────────────────────────────────┘
        │
┌───────┼──────────────────────────────────────────────┐
│  Background (Separate Isolate)                        │
│       │                                               │
│  callbackDispatcher()                                 │
│       ↓                                               │
│  BackgroundTaskHandler                                │
│       ├─ BackgroundApiClient (Dio + 토큰 자동관리)      │
│       ├─ UploadStatePersistence (상태 기록)             │
│       └─ NotificationService (완료/실패 알림)           │
└──────────────────────────────────────────────────────┘
```

### 핵심 설계 원칙
1. **Isolate 독립성**: 백그라운드 태스크는 자체 Dio 인스턴스와 토큰 관리 보유
2. **상태 브릿지**: SharedPreferences를 통한 포그라운드↔백그라운드 통신
3. **폴백 지원**: 백그라운드 불가 시 기존 포그라운드 업로드로 자동 전환
4. **파일당 태스크**: iOS 시간 제한 대응을 위해 파일별 개별 태스크 스케줄링

---

## 2. Data Model

### 2.1 UploadTaskData (백그라운드 태스크 전달 데이터)

SharedPreferences에 JSON으로 저장되는 태스크 정보.

```dart
class UploadTaskData {
  final String taskId;           // 고유 태스크 ID (UUID)
  final String sessionId;       // 서버 업로드 세션 ID
  final String filePath;        // 파일 로컬 경로
  final String filename;        // 파일명
  final int fileSize;           // 파일 크기 (bytes)
  final String status;          // pending | uploading | completed | failed
  final double progress;        // 0.0 ~ 100.0
  final String? error;          // 실패 시 에러 메시지
  final int retryCount;         // 재시도 횟수
  final DateTime createdAt;     // 생성 시각
  final DateTime? completedAt;  // 완료 시각
}
```

### 2.2 BackgroundUploadState (전체 백그라운드 업로드 상태)

```dart
class BackgroundUploadState {
  final String sessionId;               // 현재 세션 ID
  final List<UploadTaskData> tasks;    // 태스크 목록
  final bool isActive;                  // 백그라운드 업로드 활성 여부
  final DateTime? lastUpdated;          // 마지막 상태 갱신 시각
}
```

### 2.3 Settings 확장

```dart
// SettingsStorage에 추가할 키
static const _keyBackgroundUpload = 'background_upload_enabled';   // bool
static const _keyWifiOnly = 'wifi_only_upload';                     // bool
static const _keyChargingOnly = 'charging_only_upload';             // bool
```

---

## 3. Component Design

### 3.1 BackgroundUploadService

백그라운드 업로드 태스크 관리의 진입점. 포그라운드에서 사용.

```dart
// lib/core/background/background_upload_service.dart

class BackgroundUploadService {
  static const String _uploadTaskName = 'com.app.backgroundUpload';

  /// workmanager 초기화 (main.dart에서 호출)
  static Future<void> initialize() async {
    await Workmanager().initialize(callbackDispatcher, isInDebugMode: false);
  }

  /// 파일 업로드 태스크 등록
  Future<void> scheduleUpload({
    required String sessionId,
    required List<UploadFile> files,
  }) async {
    final persistence = UploadStatePersistence();

    // 각 파일을 개별 태스크로 저장
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

      // workmanager에 태스크 등록
      await Workmanager().registerOneOffTask(
        taskId,
        _uploadTaskName,
        inputData: {'taskId': taskId},
        constraints: Constraints(
          networkType:
            SettingsStorage.instance.isWifiOnly
              ? NetworkType.unmetered
              : NetworkType.connected,
          requiresCharging: SettingsStorage.instance.isChargingOnly,
        ),
        backoffPolicy: BackoffPolicy.exponential,
        initialDelay: Duration.zero,
        existingWorkPolicy: ExistingWorkPolicy.keep,
      );
    }

    // 전체 상태 저장
    await persistence.saveState(BackgroundUploadState(
      sessionId: sessionId,
      tasks: await persistence.getAllTasks(sessionId),
      isActive: true,
      lastUpdated: DateTime.now(),
    ));
  }

  /// 진행 중인 백그라운드 업로드 취소
  Future<void> cancelAll() async {
    await Workmanager().cancelAll();
    await UploadStatePersistence().clearAll();
  }

  /// 포그라운드 복귀 시 상태 동기화
  Future<BackgroundUploadState?> syncState() async {
    return await UploadStatePersistence().loadState();
  }
}
```

### 3.2 BackgroundTaskHandler

백그라운드 isolate에서 실행되는 태스크 핸들러.

```dart
// lib/core/background/background_task_handler.dart

/// workmanager의 최상위 콜백 (isolate 진입점)
@pragma('vm:entry-point')
void callbackDispatcher() {
  Workmanager().executeTask((taskName, inputData) async {
    if (taskName != BackgroundUploadService._uploadTaskName) {
      return Future.value(true);
    }

    final taskId = inputData?['taskId'] as String?;
    if (taskId == null) return Future.value(false);

    final handler = BackgroundTaskHandler();
    return await handler.execute(taskId);
  });
}

class BackgroundTaskHandler {
  late final BackgroundApiClient _apiClient;
  late final UploadStatePersistence _persistence;

  BackgroundTaskHandler() {
    _apiClient = BackgroundApiClient();
    _persistence = UploadStatePersistence();
  }

  Future<bool> execute(String taskId) async {
    final task = await _persistence.loadTask(taskId);
    if (task == null) return true; // 이미 완료된 태스크

    // 상태를 uploading으로 갱신
    await _persistence.updateTaskStatus(taskId, 'uploading');

    try {
      // API 클라이언트 초기화 (토큰 로드)
      await _apiClient.initialize();

      // 파일 업로드 실행
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

      // 성공 처리
      await _persistence.updateTaskStatus(taskId, 'completed',
          completedAt: DateTime.now());

      // 성공 알림
      await NotificationService().init();
      await NotificationService().showUploadComplete(task.filename);

      // 모든 파일 완료 확인 → 세션 완료
      await _checkAndCompleteSession(task.sessionId);

      return true;
    } catch (e) {
      // 실패 처리
      final newRetryCount = task.retryCount + 1;
      if (newRetryCount >= 3) {
        await _persistence.updateTaskStatus(taskId, 'failed',
            error: e.toString());
        await NotificationService().init();
        await NotificationService().showUploadFailed(task.filename, e.toString());
      } else {
        await _persistence.updateTaskRetry(taskId, newRetryCount);
        return false; // workmanager가 자동 재시도
      }
      return true;
    }
  }

  Future<void> _checkAndCompleteSession(String sessionId) async {
    final tasks = await _persistence.getAllTasks(sessionId);
    final allDone = tasks.every(
        (t) => t.status == 'completed' || t.status == 'failed');

    if (allDone) {
      final completedCount = tasks.where((t) => t.status == 'completed').length;
      final totalCount = tasks.length;

      try {
        await _apiClient.completeSession(sessionId);
      } catch (_) {}

      await _persistence.updateStateActive(sessionId, false);

      // 전체 완료 요약 알림
      if (completedCount == totalCount) {
        await NotificationService().init();
        await NotificationService().showUploadComplete(
            '$completedCount files uploaded successfully');
      }
    }
  }
}
```

### 3.3 BackgroundApiClient

백그라운드 isolate 전용 경량 API 클라이언트.

```dart
// lib/core/background/background_api_client.dart

class BackgroundApiClient {
  late final Dio _dio;
  final FlutterSecureStorage _storage = const FlutterSecureStorage();

  Future<void> initialize() async {
    _dio = Dio(BaseOptions(
      baseUrl: ApiConstants.baseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 10),
      headers: {'Content-Type': 'application/json'},
    ));

    // 토큰 설정
    final token = await _storage.read(key: 'access_token');
    if (token != null) {
      _dio.options.headers['Authorization'] = 'Bearer $token';
    }
  }

  /// 토큰 리프레시 (401 시 호출)
  Future<bool> refreshToken() async {
    try {
      final refreshToken = await _storage.read(key: 'refresh_token');
      if (refreshToken == null) return false;

      final response = await Dio().post(
        '${ApiConstants.baseUrl}${ApiConstants.authRefresh}',
        data: {'refresh_token': refreshToken},
      );

      if (response.statusCode == 200) {
        final newToken = response.data['data']['access_token'];
        await _storage.write(key: 'access_token', value: newToken);
        _dio.options.headers['Authorization'] = 'Bearer $newToken';
        return true;
      }
    } catch (_) {}
    return false;
  }

  /// 비디오 업로드
  Future<void> uploadVideo({
    required String sessionId,
    required String filePath,
    required String filename,
    required int fileSize,
    void Function(int, int)? onProgress,
  }) async {
    final formData = FormData.fromMap({
      'session_id': sessionId,
      'title': filename,
      'description': '',
      'file': await MultipartFile.fromFile(filePath, filename: filename),
    });

    try {
      await _dio.post(
        ApiConstants.mediaUploadVideo,
        data: formData,
        options: Options(
          headers: {'Content-Type': 'multipart/form-data'},
          receiveTimeout: const Duration(minutes: 30),
          sendTimeout: const Duration(minutes: 30),
        ),
        onSendProgress: onProgress,
      );
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        final refreshed = await refreshToken();
        if (refreshed) {
          // 재시도
          await _dio.post(
            ApiConstants.mediaUploadVideo,
            data: formData,
            options: Options(
              headers: {'Content-Type': 'multipart/form-data'},
              receiveTimeout: const Duration(minutes: 30),
              sendTimeout: const Duration(minutes: 30),
            ),
            onSendProgress: onProgress,
          );
          return;
        }
      }
      rethrow;
    }
  }

  /// 세션 완료
  Future<void> completeSession(String sessionId) async {
    await _dio.post(
      ApiConstants.mediaUploadComplete,
      data: {'session_id': sessionId},
    );
  }
}
```

### 3.4 UploadStatePersistence

SharedPreferences 기반 포그라운드↔백그라운드 상태 브릿지.

```dart
// lib/core/background/upload_state_persistence.dart

class UploadStatePersistence {
  static const _keyPrefix = 'bg_upload_';
  static const _keyState = '${_keyPrefix}state';
  static const _keyTaskPrefix = '${_keyPrefix}task_';

  /// 태스크 저장
  Future<void> saveTask(UploadTaskData task) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('$_keyTaskPrefix${task.taskId}', jsonEncode(task.toJson()));
  }

  /// 태스크 로드
  Future<UploadTaskData?> loadTask(String taskId) async {
    final prefs = await SharedPreferences.getInstance();
    final json = prefs.getString('$_keyTaskPrefix$taskId');
    if (json == null) return null;
    return UploadTaskData.fromJson(jsonDecode(json));
  }

  /// 특정 세션의 모든 태스크 로드
  Future<List<UploadTaskData>> getAllTasks(String sessionId) async {
    final prefs = await SharedPreferences.getInstance();
    final keys = prefs.getKeys().where(
        (k) => k.startsWith(_keyTaskPrefix));
    final tasks = <UploadTaskData>[];

    for (final key in keys) {
      final json = prefs.getString(key);
      if (json != null) {
        final task = UploadTaskData.fromJson(jsonDecode(json));
        if (task.sessionId == sessionId) {
          tasks.add(task);
        }
      }
    }
    return tasks;
  }

  /// 태스크 상태 업데이트
  Future<void> updateTaskStatus(String taskId, String status, {
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

  /// 진행률 업데이트
  Future<void> updateTaskProgress(String taskId, double progress) async {
    final task = await loadTask(taskId);
    if (task == null) return;
    await saveTask(task.copyWith(progress: progress));
  }

  /// 재시도 횟수 업데이트
  Future<void> updateTaskRetry(String taskId, int retryCount) async {
    final task = await loadTask(taskId);
    if (task == null) return;
    await saveTask(task.copyWith(retryCount: retryCount, status: 'pending'));
  }

  /// 전체 상태 저장/로드
  Future<void> saveState(BackgroundUploadState state) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyState, jsonEncode(state.toJson()));
  }

  Future<BackgroundUploadState?> loadState() async {
    final prefs = await SharedPreferences.getInstance();
    final json = prefs.getString(_keyState);
    if (json == null) return null;
    return BackgroundUploadState.fromJson(jsonDecode(json));
  }

  /// 세션 활성 상태 업데이트
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

  /// 전체 클리어
  Future<void> clearAll() async {
    final prefs = await SharedPreferences.getInstance();
    final keys = prefs.getKeys().where((k) => k.startsWith(_keyPrefix));
    for (final key in keys) {
      await prefs.remove(key);
    }
  }
}
```

---

## 4. Modified Components

### 4.1 UploadProvider 수정

기존 포그라운드 순차 업로드를 백그라운드 스케줄링으로 전환.

```dart
// 변경 전: startUpload() 내에서 직접 for 루프 업로드
// 변경 후: BackgroundUploadService를 통해 태스크 스케줄링

class UploadProvider extends ChangeNotifier {
  final UploadRepository _uploadRepository;
  final BackgroundUploadService _backgroundService;

  // ... 기존 필드 유지 ...

  UploadProvider(this._uploadRepository)
    : _backgroundService = BackgroundUploadService();

  Future<void> startUpload() async {
    if (_files.isEmpty) return;

    _isUploading = true;
    _error = null;
    notifyListeners();

    try {
      // 1. 세션 생성 (포그라운드에서 실행)
      final totalBytes = _files.fold<int>(0, (sum, f) => sum + f.size);
      _session = await _uploadRepository.initiateSession(
        totalFiles: _files.length,
        totalBytes: totalBytes,
      );
      notifyListeners();

      // 2. 백그라운드 업로드 설정 확인
      if (SettingsStorage.instance.isBackgroundUploadEnabled) {
        // 백그라운드 태스크로 스케줄링
        await _backgroundService.scheduleUpload(
          sessionId: _session!.sessionId,
          files: _files,
        );
        // UI 상태: 백그라운드로 전환됨을 표시
        for (final file in _files) {
          file.status = 'uploading';
        }
        notifyListeners();
      } else {
        // 기존 포그라운드 업로드 (폴백)
        await _uploadForeground();
      }
    } catch (e) {
      _error = 'Upload failed: $e';
      _isUploading = false;
      notifyListeners();
    }
  }

  /// 기존 포그라운드 업로드 로직 (폴백)
  Future<void> _uploadForeground() async {
    // ... 기존 startUpload()의 for 루프 로직 이동 ...
  }

  /// 포그라운드 복귀 시 백그라운드 상태 동기화
  Future<void> syncFromBackground() async {
    final state = await _backgroundService.syncState();
    if (state == null || !state.isActive) {
      _isUploading = false;
      notifyListeners();
      return;
    }

    // 백그라운드 태스크 상태를 UI에 반영
    for (final task in state.tasks) {
      final fileIndex = _files.indexWhere((f) => f.path == task.filePath);
      if (fileIndex >= 0) {
        _files[fileIndex].status = task.status;
        _files[fileIndex].progress = task.progress;
        _files[fileIndex].error = task.error;
      }
    }

    _isUploading = state.tasks.any(
        (t) => t.status == 'pending' || t.status == 'uploading');
    notifyListeners();
  }

  // ... cancelUpload, clearError 등 기존 메서드 유지 ...
}
```

### 4.2 QueueProvider 수정

큐 처리를 백그라운드 서비스에 위임.

```dart
// QueueProvider에 추가할 메서드

/// 큐 아이템을 백그라운드 업로드로 스케줄링
Future<void> processQueueInBackground() async {
  if (!SettingsStorage.instance.isBackgroundUploadEnabled) return;

  final pendingItems = _items.where((i) => i.isPending).toList();
  if (pendingItems.isEmpty) return;

  final backgroundService = BackgroundUploadService();
  // 각 큐 아이템을 개별 백그라운드 태스크로 등록
  for (final item in pendingItems) {
    final uploadFile = UploadFile(
      path: item.filePath,
      filename: item.filename,
      size: item.fileSizeBytes,
    );
    await backgroundService.scheduleUpload(
      sessionId: item.queueId, // 큐 아이템 ID를 세션으로 사용
      files: [uploadFile],
    );
  }
}
```

### 4.3 main.dart 수정

```dart
void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await NotificationService().init();
  await SettingsStorage.instance.init();
  await BackgroundUploadService.initialize(); // 추가
  runApp(const MyApp());
}
```

### 4.4 SettingsStorage 확장

```dart
// 추가할 키와 메서드

static const _keyBackgroundUpload = 'background_upload_enabled';
static const _keyWifiOnly = 'wifi_only_upload';
static const _keyChargingOnly = 'charging_only_upload';

bool get isBackgroundUploadEnabled {
  return _prefs?.getBool(_keyBackgroundUpload) ?? true; // 기본 활성화
}

Future<void> setBackgroundUploadEnabled(bool value) async {
  await init();
  await _prefs!.setBool(_keyBackgroundUpload, value);
}

bool get isWifiOnly {
  return _prefs?.getBool(_keyWifiOnly) ?? true; // 기본 WiFi만
}

Future<void> setWifiOnly(bool value) async {
  await init();
  await _prefs!.setBool(_keyWifiOnly, value);
}

bool get isChargingOnly {
  return _prefs?.getBool(_keyChargingOnly) ?? false;
}

Future<void> setChargingOnly(bool value) async {
  await init();
  await _prefs!.setBool(_keyChargingOnly, value);
}
```

### 4.5 Settings Screen 확장

```dart
// settings_screen.dart에 추가할 위젯

// 백그라운드 업로드 섹션
_buildSectionTitle('Background Upload'),
SwitchListTile(
  title: const Text('Background Upload'),
  subtitle: const Text('Continue uploads when app is in background'),
  value: SettingsStorage.instance.isBackgroundUploadEnabled,
  onChanged: (value) async {
    await SettingsStorage.instance.setBackgroundUploadEnabled(value);
    setState(() {});
  },
),
SwitchListTile(
  title: const Text('WiFi Only'),
  subtitle: const Text('Only upload when connected to WiFi'),
  value: SettingsStorage.instance.isWifiOnly,
  onChanged: (value) async {
    await SettingsStorage.instance.setWifiOnly(value);
    setState(() {});
  },
),
SwitchListTile(
  title: const Text('Charging Only'),
  subtitle: const Text('Only upload when device is charging'),
  value: SettingsStorage.instance.isChargingOnly,
  onChanged: (value) async {
    await SettingsStorage.instance.setChargingOnly(value);
    setState(() {});
  },
),
```

---

## 5. Platform Configuration

### 5.1 iOS Info.plist

```xml
<!-- ios/Runner/Info.plist에 추가 -->
<key>BGTaskSchedulerPermittedIdentifiers</key>
<array>
    <string>com.app.backgroundUpload</string>
</array>
<key>UIBackgroundModes</key>
<array>
    <string>fetch</string>
    <string>processing</string>
</array>
```

### 5.2 iOS AppDelegate.swift

```swift
// ios/Runner/AppDelegate.swift 수정

import UIKit
import Flutter
import workmanager  // 추가

@main
@objc class AppDelegate: FlutterAppDelegate {
  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
    GeneratedPluginRegistrant.register(with: self)

    // workmanager BGTask 등록
    WorkmanagerPlugin.registerTask(withIdentifier: "com.app.backgroundUpload")

    // BGProcessingTask 설정
    if #available(iOS 13.0, *) {
      BGTaskScheduler.shared.register(
        forTaskWithIdentifier: "com.app.backgroundUpload",
        using: nil
      ) { task in
        // workmanager가 처리
      }
    }

    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }
}
```

---

## 6. State Flow Diagrams

### 6.1 업로드 시작 Flow

```
User taps "Upload"
    ↓
UploadProvider.startUpload()
    ↓
세션 생성 (API call - 포그라운드)
    ↓
isBackgroundUploadEnabled?
    ├── YES → BackgroundUploadService.scheduleUpload()
    │            ↓
    │         파일별 workmanager 태스크 등록
    │            ↓
    │         UI: "Uploading in background..."
    │
    └── NO  → UploadProvider._uploadForeground()
                 ↓
              기존 순차 업로드 (포그라운드)
```

### 6.2 백그라운드 태스크 실행 Flow

```
workmanager triggers task
    ↓
callbackDispatcher()
    ↓
BackgroundTaskHandler.execute(taskId)
    ↓
UploadStatePersistence.loadTask(taskId)
    ↓
BackgroundApiClient.initialize() (토큰 로드)
    ↓
uploadVideo() 실행
    ├── 성공 → 상태='completed' → 알림 → 세션 완료 체크
    └── 실패 → retryCount < 3?
                  ├── YES → return false (workmanager 자동 재시도)
                  └── NO  → 상태='failed' → 실패 알림
```

### 6.3 포그라운드 복귀 Flow

```
App comes to foreground
    ↓
UploadProvider.syncFromBackground()
    ↓
UploadStatePersistence.loadState()
    ↓
Update UI with latest task statuses
    ↓
isActive?
    ├── YES → 진행 중 표시
    └── NO  → 완료/실패 결과 표시
```

---

## 7. Error Handling

| 시나리오 | 처리 |
|----------|------|
| 토큰 만료 (401) | BackgroundApiClient가 refreshToken() 시도 → 성공 시 재시도, 실패 시 task failed |
| 네트워크 끊김 | workmanager가 네트워크 조건 충족 시 자동 재스케줄링 |
| iOS가 태스크 종료 | workmanager가 자동으로 재스케줄링 (backoff policy) |
| SharedPreferences 읽기 실패 | 빈 상태 반환, 포그라운드 폴백 |
| 파일 삭제됨 (업로드 전) | FileSystemException 캐치 → task failed → 알림 |
| 3회 재시도 초과 | 상태 'failed' → 실패 알림 → 사용자 수동 재시도 가능 |

---

## 8. Dependencies

### 새로 추가할 패키지

```yaml
# pubspec.yaml에 추가
workmanager: ^0.5.2
connectivity_plus: ^6.1.4
```

### 기존 패키지 활용 (변경 없음)
- `dio: ^5.9.2`
- `flutter_secure_storage: ^10.0.0`
- `shared_preferences: ^2.5.3`
- `flutter_local_notifications: ^19.5.0`

---

## 9. Testing Strategy

### 9.1 Unit Tests

| 테스트 | 대상 | 검증 |
|--------|------|------|
| `upload_state_persistence_test.dart` | UploadStatePersistence | 태스크 CRUD, 상태 저장/복원, JSON 직렬화 |
| `upload_task_data_test.dart` | UploadTaskData | 모델 생성, copyWith, JSON 변환 |
| `background_upload_service_test.dart` | BackgroundUploadService | 태스크 스케줄링, 취소, 상태 동기화 |
| `settings_storage_test.dart` | SettingsStorage (확장) | 새 키 읽기/쓰기, 기본값 |

### 9.2 Widget Tests

| 테스트 | 대상 | 검증 |
|--------|------|------|
| `settings_screen_test.dart` | Settings Screen | 백그라운드 토글 렌더링, WiFi/충전 스위치 |

### 9.3 Integration Tests (Manual)

| 시나리오 | 검증 |
|----------|------|
| 업로드 시작 → 홈 버튼 | 업로드 지속, 알림 수신 |
| 백그라운드 업로드 중 앱 복귀 | 진행률 동기화 |
| WiFi → 셀룰러 전환 (WiFi 전용 모드) | 업로드 일시중지 |
| 배터리 저전력 모드 | 태스크 스케줄링 지연 허용 |

---

## 10. Migration Strategy

### 10.1 하위 호환성

- `SettingsStorage.isBackgroundUploadEnabled` 기본값 `true` → 업데이트 후 자동 활성화
- 기존 포그라운드 업로드 코드는 `_uploadForeground()` 메서드로 보존
- 백그라운드 비활성화 시 기존 동작과 동일

### 10.2 점진적 적용

1. 패키지 추가 + 플랫폼 설정 → 빌드 확인
2. 상태 영속화 레이어 추가 → 단위 테스트
3. BackgroundTaskHandler + BackgroundApiClient → 통합 테스트
4. UploadProvider 수정 → 기존 테스트 통과 확인
5. Settings UI → 수동 테스트
6. QueueProvider 연동 → 최종 테스트

---

## 11. Implementation Guide

### 11.1 Implementation Order

| 순서 | 파일 | 작업 | 의존성 |
|------|------|------|--------|
| 1 | `pubspec.yaml` | workmanager, connectivity_plus 추가 | 없음 |
| 2 | `ios/Runner/Info.plist` | BGTask 권한 추가 | 없음 |
| 3 | `ios/Runner/AppDelegate.swift` | BGTask 등록 | 1 |
| 4 | `lib/shared/models/upload_task_data.dart` | UploadTaskData 모델 | 없음 |
| 5 | `lib/core/background/upload_state_persistence.dart` | 상태 영속화 | 4 |
| 6 | `lib/core/background/background_api_client.dart` | 백그라운드 API 클라이언트 | 없음 |
| 7 | `lib/core/background/background_task_handler.dart` | 태스크 핸들러 | 5, 6 |
| 8 | `lib/core/background/background_upload_service.dart` | 서비스 진입점 | 5, 7 |
| 9 | `lib/core/storage/settings_storage.dart` | 설정 키 확장 | 없음 |
| 10 | `lib/main.dart` | 초기화 코드 추가 | 8 |
| 11 | `lib/features/upload/presentation/upload_provider.dart` | 백그라운드 전환 | 8, 9 |
| 12 | `lib/features/queue/presentation/queue_provider.dart` | 큐 백그라운드화 | 8 |
| 13 | `lib/features/settings/presentation/settings_screen.dart` | 설정 UI | 9 |
| 14 | 테스트 파일들 | 단위/위젯 테스트 | 4-13 |

### 11.2 File Summary

| 구분 | 파일 수 | 추정 라인 |
|------|---------|-----------|
| 새 파일 (Dart) | 5개 | ~400 라인 |
| 수정 파일 (Dart) | 4개 | ~150 라인 |
| 플랫폼 설정 | 2개 | ~20 라인 |
| 테스트 파일 | 4개 | ~300 라인 |
| **합계** | **15개** | **~870 라인** |

### 11.3 Session Guide

#### Module Map

| Module | 파일 | 설명 | 추정 시간 |
|--------|------|------|-----------|
| **module-1** | 순서 1-3 | 패키지 추가 + 플랫폼 설정 | 15분 |
| **module-2** | 순서 4-5 | 데이터 모델 + 상태 영속화 | 30분 |
| **module-3** | 순서 6-8 | 백그라운드 태스크 핵심 로직 | 45분 |
| **module-4** | 순서 9-13 | 기존 코드 수정 + UI | 30분 |
| **module-5** | 순서 14 | 테스트 작성 | 30분 |

#### Recommended Session Plan

- **Session 1**: module-1 + module-2 (설정 + 모델) → 빌드 검증
- **Session 2**: module-3 (백그라운드 핵심) → 단독 테스트
- **Session 3**: module-4 + module-5 (통합 + 테스트) → 전체 검증

```
/pdca do background-upload --scope module-1,module-2   # Session 1
/pdca do background-upload --scope module-3             # Session 2
/pdca do background-upload --scope module-4,module-5   # Session 3
```
