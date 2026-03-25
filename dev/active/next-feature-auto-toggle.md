# Next Feature: Auto/Manual Upload Toggle + Global Progress

**Requested**: 2026-03-25
**Status**: Not started — to implement in next session

## Requirements

1. **앨범 접근 → 수동 업로드**: 이미 구현됨 (image_picker → UploadScreen)
2. **자동/수동 토글**:
   - 설정 화면에 "Auto Upload" 스위치 추가
   - ON: 영상 선택 시 자동으로 큐에 추가 → 서버 스케줄러가 자동 업로드
   - OFF: 기존대로 수동 업로드 (Upload 화면에서 직접 업로드)
   - 토글 상태는 SharedPreferences에 저장
3. **글로벌 업로드 진행률 게이지**:
   - 앱 하단 또는 상단에 항상 보이는 미니 progress bar
   - 현재 업로드 중인 파일명 + 진행률(%)
   - 탭하면 상세 화면으로 이동

## Implementation Plan

### Flutter 변경사항
1. **SettingsScreen** (새로 생성)
   - Auto Upload 토글 (SharedPreferences 저장)
   - 할당량 표시
   - 라우트: /settings

2. **GlobalUploadBanner** (새 위젯)
   - 업로드 진행 중일 때 화면 하단에 고정 표시
   - 파일명 + LinearProgressIndicator + % 표시
   - 모든 화면에서 보이도록 MaterialApp 레벨에 Overlay 또는 bottomSheet

3. **UploadScreen 수정**
   - 자동 모드 ON이면: 파일 선택 → 바로 큐에 추가 → "큐에 추가됨" 알림
   - 자동 모드 OFF이면: 기존 동작 (수동 업로드)

4. **MediaListScreen**
   - AppBar 아래에 진행률 바 표시 (업로드 중일 때만)

### Backend 변경사항
- 없음 (이미 큐 시스템 + 수동 업로드 둘 다 지원)

### Files to Create/Modify
- CREATE: `lib/features/settings/presentation/settings_screen.dart`
- CREATE: `lib/shared/widgets/upload_progress_banner.dart`
- CREATE: `lib/core/storage/settings_storage.dart` (SharedPreferences wrapper)
- MODIFY: `lib/features/upload/presentation/upload_screen.dart` (auto mode branch)
- MODIFY: `lib/core/router/app_router.dart` (add /settings route)
- MODIFY: `lib/main.dart` (add settings route, global progress overlay)
- MODIFY: `lib/features/media/presentation/media_list_screen.dart` (progress bar + settings icon)

### Estimated Effort
- 약 6개 파일 생성/수정
- Flutter only (백엔드 변경 없음)
