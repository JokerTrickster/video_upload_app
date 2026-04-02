# Plan: 백그라운드 업로드

## Executive Summary

| 관점 | 내용 |
|------|------|
| **Problem** | 앱이 백그라운드로 전환되면 업로드가 즉시 중단되어, 대용량 영상 업로드 시 사용자가 앱을 수분~수십분간 열어두어야 함 |
| **Solution** | Flutter workmanager + iOS BGProcessingTask를 활용해 앱 상태와 무관하게 업로드 지속, 완료/실패 시 로컬 알림 |
| **기능/UX 효과** | 업로드 시작 후 다른 앱 사용 가능, 수동 업로드와 자동 큐 업로드 모두 백그라운드 지원 |
| **핵심 가치** | 사용자의 시간을 되찾아주는 "Set and Forget" 업로드 경험 |

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

## 1. Problem Statement

### 현재 상황
- `UploadProvider.startUpload()`는 포그라운드에서만 동작하는 순차 업로드 루프
- 앱이 백그라운드로 전환되면 iOS가 ~30초 후 프로세스를 일시정지
- 사용자는 1GB 영상 업로드(약 10분) 동안 앱을 열어두어야 함
- `QueueProvider`의 자동 큐도 같은 제약을 가짐

### 영향
- 사용자 경험 저하: 업로드 중 다른 작업 불가
- 업로드 실패 증가: 실수로 앱 전환 시 업로드 중단
- 자동 백업 기능의 실효성 저하

---

## 2. Goals & Success Metrics

### Primary Goals
1. 앱이 백그라운드에 있어도 업로드가 계속 진행됨
2. 업로드 완료/실패 시 로컬 알림으로 사용자에게 통보
3. 수동 업로드와 자동 큐 업로드 모두 백그라운드 지원
4. 배터리/데이터 사용량에 대한 사용자 제어 가능

### Success Metrics
| 지표 | 목표 |
|------|------|
| 백그라운드 업로드 성공률 | 95% 이상 |
| 앱 전환 후 업로드 지속률 | 100% (전환 직후 중단 없음) |
| 배터리 영향 | 포그라운드 대비 10% 이내 추가 소모 |
| 사용자 인지 지연 | 완료 후 알림까지 30초 이내 |

### Success Criteria
- SC-1: 앱을 백그라운드로 전환해도 업로드가 중단되지 않음
- SC-2: 업로드 완료 시 로컬 알림이 정상적으로 표시됨
- SC-3: 업로드 실패 시 재시도 또는 에러 알림이 표시됨
- SC-4: Settings에서 백그라운드 업로드 ON/OFF 토글 가능
- SC-5: WiFi 전용 모드 설정 시 셀룰러에서 백그라운드 업로드 중지

---

## 3. Scope

### In Scope
- Flutter workmanager 패키지 통합
- iOS BGProcessingTask 등록 및 설정
- Android WorkManager 통합 (Flutter workmanager가 자동 처리)
- 업로드 작업의 백그라운드 태스크화
- 로컬 알림 연동 (기존 `NotificationService` 확장)
- 백그라운드 업로드 상태 영속화 (SharedPreferences)
- Settings 화면에 백그라운드 업로드 설정 추가
- WiFi 전용 모드 지원

### Out of Scope
- 청크 업로드 (별도 feature로 진행)
- 네트워크 상태 감지 및 자동 재시도 (별도 feature)
- 업로드 진행률 실시간 알림 표시 (로컬 알림만, 상태바 진행률은 제외)
- iOS URLSession background transfer (BGProcessingTask로 충분)

---

## 4. Technical Analysis

### 현재 아키텍처
```
User Action → UploadProvider (ChangeNotifier)
                → UploadRepository.uploadVideo() (Dio HTTP)
                    → API Server
```
- 모든 로직이 Flutter 앱 프로세스 내에서 실행
- 앱 생명주기에 완전히 종속

### 목표 아키텍처
```
User Action → UploadProvider → BackgroundUploadService
                                   ↓
                              workmanager.registerOneOffTask()
                                   ↓
                              BackgroundTaskHandler (isolate)
                                   ↓
                              UploadRepository.uploadVideo()
                                   ↓
                              NotificationService (완료/실패 알림)
```

### 핵심 기술 선택

#### workmanager (Flutter 패키지)
- iOS: BGProcessingTask 래핑
- Android: Jetpack WorkManager 래핑
- 크로스플랫폼 단일 API
- 제약 조건 설정 가능: 네트워크 타입, 배터리 상태, 충전 상태

#### BGProcessingTask 특성 (iOS)
- 최대 실행 시간: 수 분 (시스템이 동적 결정)
- 충전 중 + WiFi 시 더 긴 실행 시간 부여
- 시스템이 실행 시점을 결정 (즉시 실행 보장 안 됨)
- `Info.plist`에 `BGTaskSchedulerPermittedIdentifiers` 등록 필요

### 기술적 제약 및 대응

| 제약 | 대응 |
|------|------|
| BGProcessingTask 실행 시간 제한 | 파일당 하나의 태스크로 분리, 완료 후 다음 태스크 스케줄링 |
| Dart isolate에서 Provider 접근 불가 | SharedPreferences로 상태 영속화, 포그라운드 복귀 시 동기화 |
| 시스템이 태스크 취소 가능 | 업로드 중단 지점 기록, 재스케줄링 시 이어서 진행 |
| Dio 인스턴스 isolate 공유 불가 | BackgroundTaskHandler에서 별도 Dio 인스턴스 생성 |

---

## 5. Dependencies

### 새로운 패키지
| 패키지 | 버전 | 용도 |
|--------|------|------|
| workmanager | ^0.5.2 | 백그라운드 태스크 스케줄링 |
| connectivity_plus | ^6.1.4 | WiFi/셀룰러 감지 (WiFi 전용 모드) |

### 기존 패키지 활용
| 패키지 | 활용 |
|--------|------|
| flutter_local_notifications | 완료/실패 알림 (이미 구현됨) |
| shared_preferences | 백그라운드 상태 영속화 (이미 사용중) |
| flutter_secure_storage | 토큰 접근 (이미 사용중) |
| dio | HTTP 업로드 (이미 사용중) |

### 플랫폼 설정
- iOS: `Info.plist`에 `BGTaskSchedulerPermittedIdentifiers` 추가
- iOS: `AppDelegate.swift`에 BGTask 등록 코드 추가
- Android: 별도 설정 불필요 (workmanager가 자동 처리)

---

## 6. Implementation Strategy

### Phase 1: 백그라운드 태스크 인프라 (Core)
1. workmanager 패키지 추가 및 초기화
2. `BackgroundUploadService` 클래스 생성
3. 백그라운드 태스크 핸들러 (callbackDispatcher) 구현
4. iOS Info.plist / AppDelegate 설정

### Phase 2: 업로드 로직 백그라운드화
5. 업로드 상태 영속화 레이어 (SharedPreferences 기반)
6. `UploadProvider` 수정: 백그라운드 태스크 스케줄링 방식으로 전환
7. 포그라운드 복귀 시 상태 동기화 로직
8. 백그라운드 isolate용 API 클라이언트 생성

### Phase 3: 큐 업로드 백그라운드화
9. `QueueProvider` 수정: 큐 처리를 백그라운드 태스크로 위임
10. 큐 아이템별 백그라운드 태스크 스케줄링

### Phase 4: 알림 및 설정
11. 백그라운드 업로드 완료/실패 알림 연동
12. Settings 화면에 백그라운드 업로드 토글 추가
13. WiFi 전용 모드 설정 추가
14. 배터리 최적화 설정 (충전 중만 업로드 옵션)

---

## 7. Risk Assessment

| 리스크 | 확률 | 영향 | 대응 |
|--------|------|------|------|
| iOS가 BGProcessingTask를 장시간 미실행 | 중 | 높음 | 파일당 태스크 분리 + 즉시 실행 불가 시 알림 |
| Dart isolate에서 토큰 만료 | 중 | 중 | flutter_secure_storage에서 토큰 읽기 + 리프레시 로직 포함 |
| 대용량 파일이 태스크 시간 내 완료 안 됨 | 높음 | 높음 | 업로드 진행 지점 기록, 다음 태스크에서 이어서 진행 (청크 업로드 feature와 연계) |
| Android 제조사별 배터리 최적화 차이 | 중 | 중 | workmanager 공식 가이드 따르기, 사용자에게 배터리 최적화 해제 안내 |

---

## 8. File Impact Analysis

### 새로 생성할 파일
| 파일 | 설명 |
|------|------|
| `lib/core/background/background_upload_service.dart` | 백그라운드 업로드 서비스 (태스크 등록/관리) |
| `lib/core/background/background_task_handler.dart` | callbackDispatcher + 태스크 핸들러 |
| `lib/core/background/upload_state_persistence.dart` | 업로드 상태 영속화 (SharedPreferences) |

### 수정할 파일
| 파일 | 변경 내용 |
|------|-----------|
| `lib/main.dart` | workmanager 초기화 추가 |
| `lib/features/upload/presentation/upload_provider.dart` | 백그라운드 태스크 스케줄링으로 전환 |
| `lib/features/queue/presentation/queue_provider.dart` | 큐 처리 백그라운드 위임 |
| `lib/core/notifications/notification_service.dart` | 백그라운드 완료/실패 알림 추가 |
| `lib/features/settings/presentation/settings_screen.dart` | 백그라운드 설정 UI 추가 |
| `lib/core/storage/settings_storage.dart` | 백그라운드 관련 설정 키 추가 |
| `pubspec.yaml` | workmanager, connectivity_plus 추가 |
| `ios/Runner/Info.plist` | BGTaskScheduler 권한 추가 |
| `ios/Runner/AppDelegate.swift` | BGTask 등록 |

### 영향 범위 추정
- 새 파일: 3개
- 수정 파일: 9개
- 추정 변경량: ~600 라인

---

## 9. Testing Strategy

### Unit Tests
- `BackgroundUploadService` 태스크 등록/취소 로직
- `UploadStatePersistence` 상태 저장/복원
- `UploadProvider` 백그라운드 전환 시나리오
- WiFi 전용 모드 제약 조건 검증

### Integration Tests
- 앱 백그라운드 전환 → 업로드 지속 확인
- 업로드 완료 → 알림 표시 확인
- 포그라운드 복귀 → 상태 동기화 확인
- 네트워크 타입 변경 → WiFi 전용 모드 반응 확인

### Manual Test Scenarios
- 업로드 시작 → 홈 버튼 → 다른 앱 사용 → 알림 확인
- 대용량 파일 업로드 → 앱 전환 → 완료까지 대기
- WiFi → 셀룰러 전환 시 동작 확인
- 배터리 저전력 모드에서의 동작 확인

---

## 10. Rollback Plan

- workmanager 태스크 등록 실패 시 기존 포그라운드 업로드로 폴백
- `BackgroundUploadService`를 옵셔널 의존성으로 주입하여 비활성화 가능
- Settings에서 백그라운드 업로드 OFF 시 기존 방식으로 동작
