package database

import (
	"database/sql"
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

	var s3KeyValue interface{}
	if video.S3Key != nil {
		s3KeyValue = *video.S3Key
	} else {
		s3KeyValue = nil
	}

	query := `
		INSERT INTO videos(id, user_id, title, description, status, file_size, duration, original_filename,s3_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,$11)
	`

	_, err := r.db.conn.Exec(query,
		video.ID,
		video.UserID,
		video.Title,
		video.Description,
		video.Status,
		video.FileSize,
		video.Duration,
		video.OriginalFileName,
		s3KeyValue,
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
		SELECT id, user_id, title, description, status, file_size, duration,original_filename,s3_key, created_at, updated_at
		FROM videos WHERE id = $1
	`

	var s3Key sql.NullString
	video := &models.Video{}
	err := r.db.conn.QueryRow(query, id).Scan(
		&video.ID,
		&video.UserID,
		&video.Title,
		&video.Description,
		&video.Status,
		&video.FileSize,
		&video.Duration,
		&video.OriginalFileName,
		&s3Key,
		&video.CreatedAt,
		&video.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get video by id: %w", err)
	}

	if s3Key.Valid {
		video.S3Key = &s3Key.String
	}

	return video, nil
}

func (r *VideoRepository) List() ([]*models.Video, error) {
	query := `
		SELECT id, user_id, title, description, status, file_size, duration, original_filename,s3_key, created_at, updated_at
		FROM videos
	`

	videos := []*models.Video{}
	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list videos: %w", err)
	}

	for rows.Next() {

		var s3Key sql.NullString
		video := &models.Video{}
		err := rows.Scan(
			&video.ID,
			&video.UserID,
			&video.Title,
			&video.Description,
			&video.Status,
			&video.FileSize,
			&video.Duration,
			&video.OriginalFileName,
			&s3Key,
			&video.CreatedAt,
			&video.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch the videos: %w", err)
		}

		if s3Key.Valid {
			video.S3Key = &s3Key.String
		}

		videos = append(videos, video)
	}

	return videos, nil
}

func (r *VideoRepository) Update(video *models.Video) (*models.Video, error) {

	query := `
      UPDATE videos
      SET title = $1,
          description = $2,
          status = $3,
          file_size = $4,
          duration = $5,
          s3__key = $6,
          updated_at = $7
      WHERE id = $8
    `

	var s3KeyValue interface{}
	if video.S3Key != nil {
		s3KeyValue = *video.S3Key
	} else {
		s3KeyValue = nil
	}

	_, err := r.db.conn.Exec(query,
		video.Title,
		video.Description,
		video.Status,
		video.FileSize,
		video.Duration,
		s3KeyValue,
		video.UpdatedAt,
		video.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update video: %w", err)
	}
	return video, nil
}
