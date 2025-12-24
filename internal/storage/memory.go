package storage 


import (
	"sync"
	"errors"

	"github.com/james92kj/video-platform/internal/models"
)


type MemoryStore struct {
	mu 		sync.RWMutex
	videos 	map[string]*models.Video
}


func NewMemoryStore() *MemoryStore {
	return &MemoryStore {
		videos: make(map[string]*models.Video),
	}
}


func (s *MemoryStore) Create(video *models.Video) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.videos[video.ID] = video
	return nil
}


func (s *MemoryStore) GetByID(id string) (*models.Video, error){
	s.mu.RLock()
	defer s.mu.RUnlock()

	video, exists := s.videos[id]

	if !exists {
		return nil, errors.New("video not found") 
	}
	
	return video, nil
}


func (s *MemoryStore) List() ([]*models.Video, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	videos := make([]*models.Video, 0, len(s.videos))

	for _, v := range s.videos {
		videos = append(videos,v)
	}
	
	return videos, nil
}
