# API Endpoint Flows

This document describes the complete flow for each API endpoint in the Video Platform.

---

## Table of Contents

1. [GET /](#1-get---home)
2. [GET /health](#2-get-health---health-check)
3. [POST /api/v1/videos/metadata](#3-post-apiv1videosmetadata---create-video-metadata)
4. [POST /api/v1/videos/upload-url](#4-post-apiv1videosupload-url---get-pre-signed-upload-url)
5. [GET /api/v1/videos/{id}](#5-get-apiv1videosid---get-video-by-id)
6. [GET /api/v1/videos](#6-get-apiv1videos---list-all-videos)

---

## 1. GET `/` - Home

### Purpose
Simple welcome endpoint to verify the server is running.

### Flow Diagram
```
┌──────────┐         ┌──────────┐
│  Client  │ ──GET──▶│  Server  │
└──────────┘         └──────────┘
                          │
                          ▼
                    Return welcome
                       message
                          │
                          ▼
               ┌─────────────────────┐
               │ "Welcome to video   │
               │      platform"      │
               └─────────────────────┘
```

### Request
```
GET /
```

### Response
```
Status: 200 OK
Body: "Welcome to video platform\n"
```

### Code Location
- Handler: `cmd/api/main.go` → `homeHandler()`

---

## 2. GET `/health` - Health Check

### Purpose
Health check endpoint for monitoring and load balancers to verify service is alive.

### Flow Diagram
```
┌──────────┐         ┌──────────┐
│  Client  │ ──GET──▶│  Server  │
│(or LB)   │         └──────────┘
└──────────┘              │
                          ▼
                    Return "OK"
                          │
                          ▼
               ┌─────────────────────┐
               │    Status: 200      │
               │    Body: "OK"       │
               └─────────────────────┘
```

### Request
```
GET /health
```

### Response
```
Status: 200 OK
Body: "OK\n"
```

### Code Location
- Handler: `cmd/api/main.go` → `healthHandler()`

### Notes
- Currently does NOT check database connectivity
- Future improvement: Add DB ping check

---

## 3. POST `/api/v1/videos/metadata` - Create Video Metadata

### Purpose
Create video metadata record in the database WITHOUT uploading the actual video file. This is useful when you want to register a video's information before uploading.

### Flow Diagram
```
┌──────────┐                    ┌──────────┐                    ┌────────────┐
│  Client  │                    │  Server  │                    │ PostgreSQL │
└────┬─────┘                    └────┬─────┘                    └─────┬──────┘
     │                               │                                │
     │  POST /api/v1/videos/metadata │                                │
     │  {title, description,         │                                │
     │   file_size, original_filename}                                │
     │ ─────────────────────────────▶│                                │
     │                               │                                │
     │                               │  Validate request body         │
     │                               │  ────────────────────          │
     │                               │                                │
     │                               │  Generate UUID                 │
     │                               │  ──────────────                │
     │                               │                                │
     │                               │  Build Video object            │
     │                               │  - ID: new UUID                │
     │                               │  - Status: "pending"           │
     │                               │  - UserID: hardcoded           │
     │                               │  - CreatedAt: now              │
     │                               │  - UpdatedAt: now              │
     │                               │                                │
     │                               │  INSERT INTO videos            │
     │                               │ ──────────────────────────────▶│
     │                               │                                │
     │                               │◀─────────── Success ───────────│
     │                               │                                │
     │◀───────── Response ───────────│                                │
     │  {success, message, data}     │                                │
     │                               │                                │
```

### Request
```http
POST /api/v1/videos/metadata
Content-Type: application/json

{
    "title": "My Video Title",
    "description": "Video description here",
    "file_size": 1024000,
    "original_filename": "video.mp4"
}
```

### Request Fields
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | Yes | Video title |
| `description` | string | No | Video description |
| `file_size` | int64 | Yes | File size in bytes |
| `original_filename` | string | Yes | Original file name |

### Response (Success)
```json
{
    "success": true,
    "message": "Video Metadata Created",
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "title": "My Video Title",
        "description": "Video description here",
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "status": "pending",
        "file_size": 1024000,
        "original_filename": "video.mp4",
        "duration": 0,
        "created_at": "2026-01-26T10:00:00Z",
        "updated_at": "2026-01-26T10:00:00Z"
    }
}
```

### Response (Error - Invalid Request)
```json
{
    "success": false,
    "message": "Invalid Request"
}
```

### Response (Error - DB Failure)
```json
{
    "success": false,
    "message": "Failed to create Metadata"
}
```

### Code Location
- Handler: `internal/handlers/video.go` → `CreateMetadata()`
- Model: `internal/models/video.go` → `UploadRequest`, `Video`
- Repository: `internal/database/video_repository.go` → `Create()`

### Database Operation
```sql
INSERT INTO videos(id, user_id, title, description, status, file_size, 
                   duration, original_filename, s3_key, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
```

### Notes
- ⚠️ UserID is currently HARDCODED: `550e8400-e29b-41d4-a716-446655440000`
- ⚠️ No S3 key is generated (this is metadata only, no file upload)
- Status is set to "pending"

---

## 4. POST `/api/v1/videos/upload-url` - Get Pre-signed Upload URL

### Purpose
Generate a pre-signed S3 URL that allows the client to upload a video file directly to S3, bypassing the server. This is the **recommended** upload flow.

### Flow Diagram
```
┌──────────┐                    ┌──────────┐                    ┌────────────┐          ┌─────┐
│  Client  │                    │  Server  │                    │ PostgreSQL │          │ S3  │
└────┬─────┘                    └────┬─────┘                    └─────┬──────┘          └──┬──┘
     │                               │                                │                    │
     │  POST /api/v1/videos/upload-url                                │                    │
     │  {file_name, file_size}       │                                │                    │
     │ ─────────────────────────────▶│                                │                    │
     │                               │                                │                    │
     │                               │  1. Validate request           │                    │
     │                               │  ───────────────────           │                    │
     │                               │                                │                    │
     │                               │  2. Generate Video UUID        │                    │
     │                               │  ──────────────────────        │                    │
     │                               │                                │                    │
     │                               │  3. Build S3 Key               │                    │
     │                               │  "upload/{videoID}/{fileName}" │                    │
     │                               │                                │                    │
     │                               │  4. Request Pre-signed URL     │                    │
     │                               │ ───────────────────────────────────────────────────▶│
     │                               │                                │                    │
     │                               │◀─────────── Pre-signed URL ────────────────────────│
     │                               │                                │                    │
     │                               │  5. Save metadata to DB        │                    │
     │                               │ ──────────────────────────────▶│                    │
     │                               │                                │                    │
     │                               │◀─────────── Success ───────────│                    │
     │                               │                                │                    │
     │◀───────── Response ───────────│                                │                    │
     │  {upload_url, video_id, key}  │                                │                    │
     │                               │                                │                    │
     │                               │                                │                    │
     │  6. Client uploads directly to S3 using pre-signed URL         │                    │
     │ ────────────────────────────────────────────────────────────────────────────────────▶│
     │                               │                                │                    │
     │◀─────────────────────────────────────────────────── Upload Complete ────────────────│
     │                               │                                │                    │
```

### Request
```http
POST /api/v1/videos/upload-url
Content-Type: application/json

{
    "file_name": "my-video.mp4",
    "file_size": 52428800
}
```

### Request Fields
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file_name` | string | Yes | Name of the file to upload |
| `file_size` | int64 | Yes | File size in bytes |

### Response (Success)
```json
{
    "success": true,
    "upload_url": "https://bucket.s3.region.amazonaws.com/upload/abc-123/my-video.mp4?X-Amz-...",
    "video_id": "abc-123-def-456",
    "key": "upload/abc-123-def-456/my-video.mp4",
    "expires_in": 3600,
    "message": "Upload URL Generated Successfully"
}
```

### Response (Error - Invalid Request)
```json
{
    "success": false,
    "message": "Invalid Request"
}
```

### Response (Error - S3 Failure)
```json
{
    "success": false,
    "message": "Failed to generate upload url"
}
```

### Code Location
- Handler: `internal/handlers/video.go` → `GetUploadUrl()`
- Model: `internal/models/video.go` → `GetUploadUrlRequest`, `UploadURLResponse`
- S3 Client: `internal/storage/s3_client.go` → `GeneratePreSignedUrl()`
- Repository: `internal/database/video_repository.go` → `Create()`

### S3 Key Format
```
upload/{videoID}/{fileName}

Example: upload/550e8400-e29b-41d4-a716-446655440001/my-video.mp4
```

### Client Upload Flow (After Getting URL)
```bash
# Using curl to upload directly to S3
curl -X PUT "${upload_url}" \
  -H "Content-Type: video/mp4" \
  --data-binary @my-video.mp4
```

### Notes
- ⚠️ UserID is currently HARDCODED: `550e8400-e29b-41d4-a716-446655440000`
- ⚠️ Two different UUIDs are generated (one for videoID in S3 key, one for DB record) - potential bug
- Pre-signed URL expires in 60 minutes
- Video status is set to "pending"
- Title is set to filename (no separate title field in request)

---

## 5. GET `/api/v1/videos/{id}` - Get Video by ID

### Purpose
Retrieve a single video's metadata by its UUID.

### Flow Diagram
```
┌──────────┐                    ┌──────────┐                    ┌────────────┐
│  Client  │                    │  Server  │                    │ PostgreSQL │
└────┬─────┘                    └────┬─────┘                    └─────┬──────┘
     │                               │                                │
     │  GET /api/v1/videos/{id}      │                                │
     │ ─────────────────────────────▶│                                │
     │                               │                                │
     │                               │  1. Extract ID from URL path   │
     │                               │  ───────────────────────────   │
     │                               │                                │
     │                               │  2. Validate ID not empty      │
     │                               │  ─────────────────────────     │
     │                               │                                │
     │                               │  3. Query database             │
     │                               │  SELECT * FROM videos          │
     │                               │  WHERE id = $1                 │
     │                               │ ──────────────────────────────▶│
     │                               │                                │
     │                               │◀─────────── Video Row ─────────│
     │                               │                                │
     │◀───────── Response ───────────│                                │
     │  {success, message, data}     │                                │
     │                               │                                │
```

### Request
```http
GET /api/v1/videos/550e8400-e29b-41d4-a716-446655440001
```

### Response (Success)
```json
{
    "success": true,
    "message": "Video Retrieved",
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "title": "My Video Title",
        "description": "Video description here",
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "status": "pending",
        "file_size": 1024000,
        "original_filename": "video.mp4",
        "s3_key": "upload/550e8400.../video.mp4",
        "duration": 0,
        "created_at": "2026-01-26T10:00:00Z",
        "updated_at": "2026-01-26T10:00:00Z"
    }
}
```

### Response (Error - Not Found)
```json
{
    "success": false,
    "message": "Video Not Found"
}
```

### Response (Error - Invalid ID)
```json
{
    "success": false,
    "message": "Invalid Video ID"
}
```

### Code Location
- Handler: `internal/handlers/video.go` → `GetVideo()`
- Repository: `internal/database/video_repository.go` → `GetByID()`

### Database Query
```sql
SELECT id, user_id, title, description, status, file_size, 
       duration, original_filename, s3_key, created_at, updated_at
FROM videos 
WHERE id = $1
```

### Notes
- ID is extracted from URL path using string slicing
- S3 key is optional (nullable in DB)

---

## 6. GET `/api/v1/videos` - List All Videos

### Purpose
Retrieve all videos in the database.

### Flow Diagram
```
┌──────────┐                    ┌──────────┐                    ┌────────────┐
│  Client  │                    │  Server  │                    │ PostgreSQL │
└────┬─────┘                    └────┬─────┘                    └─────┬──────┘
     │                               │                                │
     │  GET /api/v1/videos           │                                │
     │ ─────────────────────────────▶│                                │
     │                               │                                │
     │                               │  Query all videos              │
     │                               │  SELECT * FROM videos          │
     │                               │ ──────────────────────────────▶│
     │                               │                                │
     │                               │◀────────── All Rows ───────────│
     │                               │                                │
     │                               │  Build response array          │
     │                               │  ─────────────────────         │
     │                               │                                │
     │◀───────── Response ───────────│                                │
     │  {success, count, data: [...]}│                                │
     │                               │                                │
```

### Request
```http
GET /api/v1/videos
```

### Response (Success)
```json
{
    "success": true,
    "count": 2,
    "data": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440001",
            "title": "First Video",
            "description": "Description 1",
            "user_id": "550e8400-e29b-41d4-a716-446655440000",
            "status": "pending",
            "file_size": 1024000,
            "original_filename": "video1.mp4",
            "s3_key": "upload/550e8400.../video1.mp4",
            "duration": 0,
            "created_at": "2026-01-26T10:00:00Z",
            "updated_at": "2026-01-26T10:00:00Z"
        },
        {
            "id": "550e8400-e29b-41d4-a716-446655440002",
            "title": "Second Video",
            "description": "Description 2",
            "user_id": "550e8400-e29b-41d4-a716-446655440000",
            "status": "ready",
            "file_size": 2048000,
            "original_filename": "video2.mp4",
            "s3_key": "upload/550e8400.../video2.mp4",
            "duration": 120,
            "created_at": "2026-01-26T11:00:00Z",
            "updated_at": "2026-01-26T11:00:00Z"
        }
    ]
}
```

### Response (Error)
```json
{
    "success": false,
    "message": "No videos Found"
}
```

### Code Location
- Handler: `internal/handlers/video.go` → `ListVideos()`
- Repository: `internal/database/video_repository.go` → `List()`

### Database Query
```sql
SELECT id, user_id, title, description, status, file_size, 
       duration, original_filename, s3_key, created_at, updated_at
FROM videos
```

### Notes
- ⚠️ No pagination implemented - returns ALL videos
- ⚠️ No filtering by user_id
- Response format differs from other endpoints (uses `count` instead of `message`)

---

## Complete Upload Flow (End-to-End)

Here's how a typical video upload works from start to finish:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         COMPLETE UPLOAD FLOW                                 │
└─────────────────────────────────────────────────────────────────────────────┘

Step 1: Request Upload URL
──────────────────────────
    Client                          Server                          S3
       │                               │                             │
       │  POST /api/v1/videos/upload-url                             │
       │  {file_name, file_size}       │                             │
       │ ─────────────────────────────▶│                             │
       │                               │──── Generate Pre-signed ───▶│
       │                               │◀─────── URL ────────────────│
       │◀─── {upload_url, video_id} ───│                             │


Step 2: Upload to S3
────────────────────
    Client                          Server                          S3
       │                               │                             │
       │                               │                             │
       │────────── PUT video file ─────────────────────────────────▶│
       │                               │                             │
       │◀────────────────────────────────────── 200 OK ─────────────│


Step 3: (TODO) Confirm Upload
─────────────────────────────
    Client                          Server                          S3
       │                               │                             │
       │  POST /api/v1/videos/upload-complete                        │
       │  {video_id}                   │                             │
       │ ─────────────────────────────▶│                             │
       │                               │──── Verify object exists ──▶│
       │                               │◀──────── Confirmed ─────────│
       │                               │                             │
       │                               │  Update status to "ready"   │
       │◀──────── Success ─────────────│                             │


Step 4: (TODO) Video Processing
───────────────────────────────
    Server                          Processing                       S3
       │                               │                             │
       │──── Start transcode job ─────▶│                             │
       │                               │◀─── Download original ──────│
       │                               │                             │
       │                               │  Transcode to 240p, 480p,   │
       │                               │  720p, 1080p                │
       │                               │                             │
       │                               │───── Upload segments ──────▶│
       │                               │───── Upload manifests ─────▶│
       │                               │                             │
       │◀───── Job Complete ───────────│                             │
```

---

## Video Status States

| Status | Description |
|--------|-------------|
| `pending` | Metadata created, awaiting upload or processing |
| `uploading` | (Future) Upload in progress |
| `processing` | (Future) Video is being transcoded |
| `ready` | (Future) Video is ready for playback |
| `error` | (Future) Something went wrong |

---

## Known Issues & TODOs

1. **Hardcoded User ID** - All videos use the same user ID
2. **No Authentication** - No JWT/session validation
3. **Double UUID in GetUploadUrl** - Creates two different UUIDs
4. **No Pagination** - List endpoint returns all videos
5. **No Upload Confirmation** - No webhook/callback when S3 upload completes
6. **Missing Endpoints**:
   - `DELETE /api/v1/videos/{id}` - Delete a video
   - `PUT /api/v1/videos/{id}` - Update video metadata
   - `POST /api/v1/videos/upload-complete` - Confirm upload
   - `GET /api/v1/videos/{id}/stream` - Get streaming URL

---

*Last Updated: January 26, 2026*

