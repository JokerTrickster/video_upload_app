---
name: youtube-gallery
description: YouTube 썸네일을 갤러리에 표시하고 탭하면 YouTube 앱으로 재생
status: in-progress
created: 2026-03-25T22:31:05Z
updated: 2026-03-25T22:31:05Z
---

## Executive Summary

| 항목 | 내용 |
|------|------|
| Feature | YouTube Gallery Display |
| 시작일 | 2026-03-25 |
| 예상 규모 | Small (DB 1필드 추가 + API 수정 + Flutter UI 수정) |

| 관점 | 설명 |
|------|------|
| **Problem** | 갤러리에 업로드된 영상이 아이콘으로만 표시되어 어떤 영상인지 식별 불가 |
| **Solution** | 업로드 완료 시 YouTube 썸네일 URL을 DB에 저장, 갤러리에 이미지로 표시 |
| **Function UX Effect** | 갤러리가 시각적으로 풍부해지고, 탭 한번으로 YouTube 재생 가능 |
| **Core Value** | 사용자가 업로드한 영상을 빠르게 식별하고 바로 시청 가능 |

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | 갤러리에 아이콘만 보이면 어떤 영상인지 알 수 없어 사용성이 떨어짐 |
| **WHO** | 영상 업로드 후 관리하려는 앱 사용자 |
| **RISK** | YouTube API 썸네일 URL 변경 가능성 (낮음 — 표준 URL 패턴 사용) |
| **SUCCESS** | 갤러리 목록에서 썸네일이 보이고, 탭하면 YouTube 앱에서 재생됨 |
| **SCOPE** | Backend: DB 마이그레이션 1개 + API 응답 수정 / Flutter: 모델 1필드 + UI 2화면 수정 |

## 1. Requirements

### 1.1 Backend Changes

1. **DB Migration**: `media_assets` 테이블에 `thumbnail_url VARCHAR(1024)` 컬럼 추가
2. **Upload Flow 수정**: YouTube 업로드 완료 시 `UploadVideoResponse.ThumbnailURL`을 `MediaAsset.ThumbnailURL`에 저장
   - `upload_service.go`: `MarkAsCompleted` 후 thumbnail 저장
   - `queue_service.go`: `processItem` 내 동일 처리
3. **API Response 수정**: `MediaAssetResponse` DTO에 `thumbnail_url` 필드 추가
4. **기존 데이터 처리**: 이미 업로드된 영상은 `https://img.youtube.com/vi/{videoId}/hqdefault.jpg` URL로 역채움 가능 (optional migration)

### 1.2 Flutter Changes

1. **Model 수정**: `MediaAssetModel`에 `thumbnailUrl` 필드 추가
2. **MediaListScreen**: 갤러리 카드에 `Image.network(thumbnailUrl)` 표시
   - 썸네일 없으면 기존 아이콘 표시 (fallback)
   - `CachedNetworkImage` 또는 기본 `Image.network` + `errorBuilder`
3. **MediaDetailScreen**: 상단에 썸네일 이미지 크게 표시 + "Open in YouTube" 버튼 (이미 존재)

### 1.3 Constraints

- YouTube 썸네일 URL 패턴: `https://img.youtube.com/vi/{videoId}/hqdefault.jpg` (480x360)
- API 썸네일이 없는 경우 videoId로 직접 구성 가능 (fallback)
- 네트워크 없을 때 썸네일 로딩 실패 → placeholder 아이콘

## 2. Implementation Plan

### Phase 1: Backend (DB + Service)
1. `media_asset.go`: `ThumbnailURL *string` 필드 추가
2. GORM AutoMigrate가 자동으로 컬럼 추가
3. `upload_service.go`: 업로드 성공 시 `asset.ThumbnailURL = &uploadResp.ThumbnailURL`
4. `queue_service.go`: 동일 처리
5. `dto.go`: `MediaAssetResponse`에 `ThumbnailURL` 추가
6. `media_handler.go`: 응답 매핑에 `ThumbnailURL` 추가

### Phase 2: Flutter (Model + UI)
1. `media_asset_model.dart`: `thumbnailUrl` 필드 추가
2. `media_list_screen.dart`: 카드 위젯에 썸네일 이미지 표시
3. `media_detail_screen.dart`: 상단에 썸네일 크게 표시

### Phase 3: Tests
1. Backend: `MediaAsset.ThumbnailURL` 저장 테스트
2. Backend: API 응답에 thumbnail_url 포함 테스트
3. Flutter: `MediaAssetModel.thumbnailUrl` 파싱 테스트
4. Flutter: 썸네일 있는/없는 UI 렌더링 테스트

## 3. Risk Analysis

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| YouTube 썸네일 URL 변경 | Medium | Low | videoId 기반 표준 URL 패턴 사용 |
| 썸네일 로딩 실패 (네트워크) | Low | Medium | errorBuilder로 fallback 아이콘 표시 |
| 기존 데이터 thumbnail 없음 | Low | Certain | videoId로 URL 역생성 (Flutter에서 처리) |

## 4. Success Criteria

- [ ] 새로 업로드된 영상의 썸네일이 DB에 저장됨
- [ ] 갤러리 목록에 썸네일 이미지가 표시됨
- [ ] 기존 영상도 videoId로 썸네일 표시됨 (fallback)
- [ ] 썸네일 탭 → MediaDetail → "Open in YouTube" → YouTube 앱 재생
- [ ] 네트워크 없을 때 graceful fallback (아이콘)
- [ ] 모든 관련 테스트 통과
