package database

import (
	"fmt"
	"github.com/james92kj/video-platform/internal/models"
	"github.com/james92kj/video-platform/internal/storage"
)

var _ storage.VideoStore = (*VideoRepository)(nil)

type VideoRepository struct {
	db *DB
}

func NewVideoRespository(db *DB) *VideoRepository {
	return &VideoRepository{
		db: db,
	}
}

func (r *VideoRepository) Create(video *models.Video) (*models.Video, error) {
	query := `
		INSERT INTO (id, user_id, title, description, status, file_size, duration, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.conn.Exec(query,
		video.ID,
		video.UserID,
		video.Title,
		video.Description,
		video.Status,
		video.FileSize,
		video.Duration,
		video.CreatedAt,
		video.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create video: %w", err)
	}
	return video, nil
}

func (r *VideoRepository) GetByID(id string) (*models.Video, error) {
	query := `
		SELECT id, user_id, title, description, status, file_size, duration, created_at, updated_at
		FROM videos WHERE id = $1
	`

	video := &models.Video{}
	err := r.db.conn.QueryRow(query, id).Scan(
		&video.ID,
		&video.UserID,
		&video.Title,
		&video.Description,
		&video.Status,
		&video.FileSize,
		&video.Duration,
		&video.CreatedAt,
		&video.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get video by id: %w", err)
	}
	return video, nil
}

func (r *VideoRepository) List() ([]*models.Video, error) {
	query := `
		SELECT id, user_id, title, description, status, file_size, duration, created_at, updated_at
		FROM videos
	`

	videos := []*models.Video{}
	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list videos: %w", err)
	}

	for rows.Next() {
		video := &models.Video{}
		err := rows.Scan(
			&video.ID,
			&video.UserID,
			&video.Title,
			&video.Description,
			&video.Status,
			&video.FileSize,
			&video.Duration,
			&video.CreatedAt,
			&video.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch the videos: %w", err)
		}
		videos = append(videos, video)
	}

	return videos, nil
}
