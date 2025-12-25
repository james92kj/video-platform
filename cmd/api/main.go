package main

import (
	"fmt"
	"github.com/james92kj/video-platform/internal/database"
	"github.com/james92kj/video-platform/internal/storage"
	"net/http"

	"github.com/james92kj/video-platform/internal/config"
	"github.com/james92kj/video-platform/internal/handlers"
	"github.com/james92kj/video-platform/internal/logger"
)

func main() {

	cfg := config.Load()
	log := logger.New()

	log.Info("Video Platform sharing..")

	// Connect to Postgres
	connection_str := "postgresql://postgres:postgres@localhost:5432/video-platform?sslmode=disable"
	db, err := database.New(connection_str)

	if err != nil {
		log.Fatal("Error connecting to database", err)
	}
	defer db.Close()
	log.Info("Connected to database")

	// Initialize the s3 client
	s3_client, err := storage.NewS3Client(log)
	if err != nil {
		log.Fatal("Error connecting to s3 client", err)
	}

	videoRepo := database.NewVideoRespository(db)
	videoHandler := handlers.NewVideoHandler(videoRepo, log, s3_client)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/v1/videos/metadata", videoHandler.CreateMetadata)
	http.HandleFunc("/api/v1/videos/upload-url", videoHandler.GetUploadUrl)
	http.HandleFunc("/api/v1/videos/", videoHandler.GetVideo)
	http.HandleFunc("/api/v1/videos", videoHandler.ListVideos)

	port := ":" + cfg.Port
	log.Info(fmt.Sprintf("Server running on http://localhost%s\n", port))

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Error(fmt.Sprintf("Server failed: %v", err))
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Welcome to video platform\n")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\n")
}
