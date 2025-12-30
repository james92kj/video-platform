# Video Platform

A simple video sharing platform API built with Go.

## Stack

- Go 1.24
- PostgreSQL
- HTTP REST API

## Setup

```bash
# Run PostgreSQL
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=video-platform postgres

# Build & Run
go build -o api cmd/api/main.go
./api
```

## API Endpoints

**Health Check**
```bash
curl http://localhost:8080/health
```

**Create Video Metadata**
```bash
curl -X POST http://localhost:8080/api/v1/videos/metadata \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Video",
    "description": "Video description",
    "file_size": 1024000,
    "original_filename": "video.mp4"
  }'
```

**Get Video by ID**
```bash
curl http://localhost:8080/api/v1/videos/{video-id}
```

**List All Videos**
```bash
curl http://localhost:8080/api/v1/videos
```

## Configuration

- `PORT` - Server port (default: 8080)
- `ENV` - Environment (default: dev)

