package main

import (
	"fmt"
	"github.com/james92kj/video-platform/internal/config"
	database2 "github.com/james92kj/video-platform/internal/database"
	"github.com/james92kj/video-platform/internal/handlers"
	"github.com/james92kj/video-platform/internal/logger"
	"github.com/james92kj/video-platform/internal/storage"
	"net/http"
)

func main() {

	cfg := config.Load()
	log := logger.New()

	log.Info("Video Platform sharing..")

	// Connect to Postgres
	connection_str := "postgresql://postgres:postgres@localhost:5432/video-platform?sslmode=disable"
	db, err := database2.New(connection_str)

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

	videoRepo := database2.NewVideoRespository(db)
	videoHandler := handlers.NewVideoHandler(videoRepo, log, s3_client)

	// Add CORS middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/videos/metadata", videoHandler.CreateMetadata)
	mux.HandleFunc("/api/v1/videos/upload-url", videoHandler.GetUploadUrl)
	mux.HandleFunc("/api/v1/videos/", videoHandler.GetVideo)
	mux.HandleFunc("/api/v1/videos", videoHandler.ListVideos)

	// Wrap with CORS
	handler := enableCORS(mux)

	port := ":" + cfg.Port
	log.Info(fmt.Sprintf("Server running on http://localhost%s\n", port))

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Error(fmt.Sprintf("Server failed: %v", err))
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Welcome to video platform\n")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\n")
}
