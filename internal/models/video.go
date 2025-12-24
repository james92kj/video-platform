package models

import "time"

type Video struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	UserID           string    `json:"user_id"`
	Status           string    `json:"status"`
	FileSize         int64     `json:"file_size"`
	OriginalFileName string    `json:"original_filename"`
	Duration         int       `json:"duration"`
	CreatedAt        time.Time `json:created_at`
	UpdatedAt        time.Time `json:updated_at`
}

type UploadRequest struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	FileSize         int64  `json:"file_size"`
	OriginalFileName string `json:"original_filename"`
}

type VideoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *Video `json:"data,omitempty"`
}
