package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/james92kj/video-platform/internal/logger"
	"github.com/james92kj/video-platform/internal/models"
	"github.com/james92kj/video-platform/internal/storage"

	"github.com/google/uuid"
)

type VideoHandler struct {
	store storage.VideoStore
	log   *logger.Logger
}

func NewVideoHandler(store storage.VideoStore, log *logger.Logger) *VideoHandler {

	return &VideoHandler{
		store: store,
		log:   log,
	}
}

func (h *VideoHandler) sendSuccess(w http.ResponseWriter, message string, data *models.Video) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(models.VideoResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func (h *VideoHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(models.VideoResponse{
		Success: false,
		Message: message,
	})

}

func (h *VideoHandler) CreateMetadata(w http.ResponseWriter, r *http.Request) {

	// validate if it is the right method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	video := &models.Video{
		ID:               uuid.New().String(),
		Title:            req.Title,
		Description:      req.Description,
		UserID:           "550e8400-e29b-41d4-a716-446655440000",
		Status:           "pending",
		OriginalFileName: req.OriginalFileName,
		FileSize:         req.FileSize,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if _, err := h.store.Create(video); err != nil {
		h.log.Error(fmt.Sprintf("Failed to decode request body: %v", err))
		h.sendError(w, "Failed to create Metadata", http.StatusInternalServerError)
		return
	}

	h.log.Info("Video Metadata Created: " + video.ID)
	h.sendSuccess(w, "Video Metadata Created", video)
}

func (h *VideoHandler) GetVideo(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the ID from url
	id := r.URL.Path[len("/api/v1/videos/"):]
	if id == "" {
		h.sendError(w, "Invalid Video ID", http.StatusBadRequest)
		return
	}

	video, err := h.store.GetByID(id)
	if err != nil {
		h.sendError(w, "Video Not Found", http.StatusBadRequest)
		return
	}

	h.sendSuccess(w, "Video Retrieved", video)
}

func (h *VideoHandler) ListVideos(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	videos, err := h.store.List()
	if err != nil {
		h.sendError(w, "No videos Found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(videos),
		"data":    videos,
	})
}
