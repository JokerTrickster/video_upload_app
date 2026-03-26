---
name: youtube-gallery
description: YouTube 썸네일을 갤러리에 표시하고 탭하면 YouTube 앱으로 재생
status: in-progress
created: 2026-03-26T01:35:45Z
updated: 2026-03-26T01:35:45Z
---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | 갤러리에 아이콘만 보이면 어떤 영상인지 알 수 없어 사용성이 떨어짐 |
| **WHO** | 영상 업로드 후 관리하려는 앱 사용자 |
| **RISK** | YouTube API 썸네일 URL 변경 가능성 (낮음 — 표준 URL 패턴 사용) |
| **SUCCESS** | 갤러리 목록에서 썸네일이 보이고, 탭하면 YouTube 앱에서 재생됨 |
| **SCOPE** | Backend 6파일 수정 / Flutter 3파일 수정 + 테스트 |

## 1. Overview

YouTube에 업로드 완료된 영상의 썸네일을 갤러리에 표시하여 사용자가 영상을 시각적으로 식별할 수 있게 한다.

**Architecture: Pragmatic Balance**
- DB에 `thumbnail_url` 컬럼 추가 (새 업로드용)
- Flutter에서 `youtubeVideoId` 기반 fallback URL 생성 (기존 데이터용)
- 추가 패키지 불필요 (`Image.network` + `errorBuilder` 사용)

## 2. Data Flow

```
Upload Complete
  → YouTube API returns ThumbnailURL
  → upload_service / queue_service saves to MediaAsset.ThumbnailURL
  → API returns thumbnail_url in response
  → Flutter MediaAssetModel.thumbnailUrl receives it
  → _MediaCard displays Image.network(thumbnailUrl)

Fallback (existing data without thumbnail_url):
  → Flutter checks: thumbnailUrl ?? youtubeVideoId
  → If youtubeVideoId exists: generate https://img.youtube.com/vi/{id}/hqdefault.jpg
  → If neither: show video icon placeholder
```

## 3. Backend Changes

### 3.1 Domain Model (`internal/domain/media_asset.go`)

Add field to `MediaAsset` struct:
```go
ThumbnailURL *string `json:"thumbnail_url,omitempty" gorm:"type:varchar(1024)"`
```
Position: after `S3ObjectKey` field (line ~33).
GORM AutoMigrate handles column addition automatically.

### 3.2 Upload Service (`internal/service/upload_service.go`)

In `UploadVideo()` method, after successful YouTube upload and `MarkAsCompleted`:
```go
// Save thumbnail URL
if uploadResp.ThumbnailURL != "" {
    asset.ThumbnailURL = &uploadResp.ThumbnailURL
}
```

### 3.3 Queue Service (`internal/service/queue_service.go`)

In `processItem()` method, after successful YouTube upload:
```go
// Save thumbnail URL
if uploadResp.ThumbnailURL != "" {
    asset.ThumbnailURL = &uploadResp.ThumbnailURL
}
```

### 3.4 DTO (`internal/handler/dto.go`)

Add to `MediaAssetResponse`:
```go
ThumbnailURL *string `json:"thumbnail_url,omitempty"`
```

### 3.5 Handler (`internal/handler/media_handler.go`)

Add to both `MediaAssetResponse{}` mappings (list and get):
```go
ThumbnailURL: asset.ThumbnailURL,
```

## 4. Flutter Changes

### 4.1 Model (`lib/shared/models/media_asset_model.dart`)

Add field:
```dart
final String? thumbnailUrl;
```

Add to constructor and `fromJson`:
```dart
thumbnailUrl: json['thumbnail_url'] as String?,
```

Add computed property for fallback:
```dart
String? get effectiveThumbnailUrl {
  if (thumbnailUrl != null && thumbnailUrl!.isNotEmpty) return thumbnailUrl;
  if (youtubeVideoId != null && youtubeVideoId!.isNotEmpty) {
    return 'https://img.youtube.com/vi/$youtubeVideoId/hqdefault.jpg';
  }
  return null;
}
```

### 4.2 Gallery Card (`lib/features/media/presentation/media_list_screen.dart`)

Replace `_buildStatusIcon()` in `_MediaCard`:
- If `asset.effectiveThumbnailUrl != null` → show thumbnail with status overlay
- If null → show current icon fallback

```dart
Widget _buildLeading(Responsive r) {
  final url = asset.effectiveThumbnailUrl;
  final size = r.iconLarge * 1.5; // Slightly larger for thumbnail

  if (url != null && asset.isCompleted) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: SizedBox(
        width: size,
        height: size,
        child: Image.network(
          url,
          fit: BoxFit.cover,
          errorBuilder: (_, __, ___) => _buildStatusIcon(r),
        ),
      ),
    );
  }
  return _buildStatusIcon(r);
}
```

### 4.3 Detail Screen (`lib/features/media/presentation/media_detail_screen.dart`)

Add thumbnail hero at top of Column (before _StatusCard):
```dart
if (asset.effectiveThumbnailUrl != null) ...[
  ClipRRect(
    borderRadius: BorderRadius.circular(12),
    child: AspectRatio(
      aspectRatio: 16 / 9,
      child: Image.network(
        asset.effectiveThumbnailUrl!,
        fit: BoxFit.cover,
        errorBuilder: (_, __, ___) => Container(
          color: Colors.grey[200],
          child: Icon(Icons.video_file, size: 64, color: Colors.grey),
        ),
      ),
    ),
  ),
  const SizedBox(height: 16),
],
```

## 5. Test Plan

### 5.1 Backend Tests

| Test | File | Description |
|------|------|-------------|
| MediaAsset ThumbnailURL field | `media_asset_test.go` | Validate field is optional, stored correctly |
| UploadService saves thumbnail | `upload_service_test.go` | Verify ThumbnailURL set after upload |
| Handler maps ThumbnailURL | `media_handler_test.go` | Response includes thumbnail_url |

### 5.2 Flutter Tests

| Test | File | Description |
|------|------|-------------|
| Model parses thumbnail_url | `models_test.dart` | fromJson with/without thumbnail_url |
| effectiveThumbnailUrl fallback | `models_test.dart` | Returns videoId-based URL when thumbnailUrl is null |
| MediaCard with thumbnail | `media_list_screen_test.dart` | Renders Image.network when URL present |
| MediaCard without thumbnail | `media_list_screen_test.dart` | Falls back to status icon |

## 6. File Change Summary

| File | Action | Lines Changed |
|------|--------|---------------|
| `backend/internal/domain/media_asset.go` | Modify | +1 (field) |
| `backend/internal/service/upload_service.go` | Modify | +3 (thumbnail save) |
| `backend/internal/service/queue_service.go` | Modify | +3 (thumbnail save) |
| `backend/internal/handler/dto.go` | Modify | +1 (field) |
| `backend/internal/handler/media_handler.go` | Modify | +2 (mapping x2) |
| `app/lib/shared/models/media_asset_model.dart` | Modify | +10 (field + getter) |
| `app/lib/features/media/presentation/media_list_screen.dart` | Modify | +20 (thumbnail card) |
| `app/lib/features/media/presentation/media_detail_screen.dart` | Modify | +15 (thumbnail hero) |
| Backend tests | Modify | +15 |
| Flutter tests | Modify | +20 |
| **Total** | | **~90 lines** |

## 7. Implementation Guide

### 7.1 Implementation Order

1. `media_asset.go` — Add ThumbnailURL field
2. `upload_service.go` — Save thumbnail after upload
3. `queue_service.go` — Save thumbnail after queue upload
4. `dto.go` — Add to response DTO
5. `media_handler.go` — Map field in responses
6. `media_asset_model.dart` — Add field + effectiveThumbnailUrl
7. `media_list_screen.dart` — Thumbnail in card
8. `media_detail_screen.dart` — Thumbnail hero image
9. Backend tests
10. Flutter tests

### 7.2 Dependencies

None — no new packages required.

### 7.3 Session Guide

**Module Map:**
| Module | Files | Description |
|--------|-------|-------------|
| module-1 | Backend (steps 1-5) | Domain + Service + Handler |
| module-2 | Flutter (steps 6-8) | Model + UI |
| module-3 | Tests (steps 9-10) | Backend + Flutter tests |

**Recommended**: Single session (all modules, ~30 min)
