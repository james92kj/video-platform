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
	store    storage.VideoStore
	log      *logger.Logger
	s3client *storage.S3Client
}

func NewVideoHandler(store storage.VideoStore,
	log *logger.Logger,
	s3client *storage.S3Client,
) *VideoHandler {

	return &VideoHandler{
		store:    store,
		log:      log,
		s3client: s3client,
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
		h.log.Error(fmt.Sprintf("Failed to decode request body: %v", err))
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
		h.log.Error(fmt.Sprintf("Failed to decode request body: %v", err))
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

func (h *VideoHandler) GetUploadUrl(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	h.log.Info("===== GetUploadUrl Handler Started ====")

	// Step 2: Parse request body
	var req models.GetUploadUrlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error(fmt.Sprintf("Failed to decode request body: %v", err))
		h.sendError(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	h.log.Info(fmt.Sprintf("Decoded Request: Fileame=%s, FileSize=%d", req.FileName, req.FileSize))

	// Next generate the Pre-signed Url
	videoID := uuid.New().String()
	s3Key := fmt.Sprintf("upload/%s/%s", videoID, req.FileName)
	h.log.Info(fmt.Sprintf("Generated VideoID=%s, S3Key=%s", videoID, s3Key))

	// Generate the Pre-signed URL
	ctx := r.Context()
	uploadURL, err := h.s3client.GeneratePreSignedUrl(ctx, s3Key, 60)
	if err != nil {
		h.log.Error(fmt.Sprintf("Failed to generate upload url: %v", err))
		h.sendError(w, "Failed to generate upload url", http.StatusInternalServerError)
		return
	}
	h.log.Info(fmt.Sprintf("Generated upload url: %s", uploadURL))

	// Build the response object
	response := models.UploadURLResponse{
		Success:   true,
		UploadURL: uploadURL,
		VideoID:   videoID,
		Key:       s3Key,
		ExpiresIn: 60 * 60,
		Message:   "Upload URL Generated Successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.log.Info(fmt.Sprintf("Response Sent For VideoID: %s", videoID))
}
