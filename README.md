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

- `GET /health` - Health check
- `POST /api/v1/videos/metadata` - Create video metadata
- `GET /api/v1/videos/{id}` - Get video by ID
- `GET /api/v1/videos` - List all videos
- `GET /api/v1/videos/GetUploadURL` - Generate the Pre-signed URL

## Configuration

- `PORT` - Server port (default: 8080)
- `ENV` - Environment (default: dev)

