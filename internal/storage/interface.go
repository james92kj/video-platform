package storage

import (
	"github.com/james92kj/video-platform/internal/models"
)

type VideoStore interface {
	Create(video *models.Video) (*models.Video, error)
	GetByID(id string) (*models.Video, error)
	List() ([]*models.Video, error)
}
