package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/james92kj/video-platform/internal/logger"
	"github.com/james92kj/video-platform/internal/models"
	storage2 "github.com/james92kj/video-platform/internal/storage"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type VideoHandler struct {
	store    storage2.VideoStore
	log      *logger.Logger
	s3client *storage2.S3Client
}

func NewVideoHandler(store storage2.VideoStore,
	log *logger.Logger,
	s3client *storage2.S3Client,
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

	err := json.NewEncoder(w).Encode(models.VideoResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
	if err != nil {
		return
	}
}

func (h *VideoHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(models.VideoResponse{
		Success: false,
		Message: message,
	})
	if err != nil {
		return
	}

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

func (h *VideoHandler) HandleUploadComplete(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Webhook received")); err != nil {
		h.log.Error(fmt.Sprintf("Failed to write response: %v", err))
	}

	h.log.Info("----- Upload Complete Webhook Request Received -----")

	// Parse the webhook payload
	var event struct {
		VideoID string `json:"video_id"`
		S3Key   string `json:"s3_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.log.Error(fmt.Sprintf("Failed to decode webhook payload: %v", err))
		h.sendError(w, "Failed to decode webhook payload", http.StatusBadRequest)
		return
	}

	h.log.Info(fmt.Sprintf("Upload Complete for VideoID: %s, S3Key: %s", event.VideoID, event.S3Key))

	// Get the video from database
	video, err := h.store.GetByID(event.VideoID)
	if err != nil {
		h.log.Error(fmt.Sprintf("No video found under the VideoID: %s", event.VideoID))
		h.sendError(w, "Video Not Found", http.StatusBadRequest)
		return
	}

	// 2. Update video status to "uploaded"
	video.Status = "uploaded"
	video.S3Key = &event.S3Key
	video.UpdatedAt = time.Now()

	// Save to database
	if _, err := h.store.Update(video); err != nil {
		h.log.Error(fmt.Sprintf("Failed to update video: %v", err))
		h.sendError(w, "Issue in saving the video", http.StatusBadRequest)
		return
	}
	h.sendSuccess(w, "Video Uploaded", video)

	h.log.Info(fmt.Sprintf("Ready to start processing video:%s", video.ID))

	// Respond with Success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"video_id": video.ID,
		"status":   "Upload Acknowledged",
		"message":  "Video uploaded complete. Will start processing the video shortly",
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

	// Build the video object
	video := &models.Video{
		ID:               uuid.New().String(),
		Title:            req.FileName,
		Description:      "",
		UserID:           "550e8400-e29b-41d4-a716-446655440000",
		Status:           "pending",
		S3Key:            &s3Key,
		OriginalFileName: req.FileName,
		FileSize:         req.FileSize,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if _, err := h.store.Create(video); err != nil {
		h.log.Error(fmt.Sprintf("Failed to store the video data: %v", err))
		h.sendError(w, "Failed to store the video data", http.StatusInternalServerError)
		return
	}

	h.log.Info(fmt.Sprintf("Video Created: %s", video.ID))

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
