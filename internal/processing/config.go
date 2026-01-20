package processing

import (
	"fmt"
	"path/filepath"
)
import "github.com/james92kj/video-platform/internal/models"

type ProcessingConfig struct {
	FFmpegPath      string
	TempDir         string
	SegmentDuration int
	Resolutions     []models.Resolution
	S3Path          S3PathConfig
}

type S3PathConfig struct {
	UploadsPrefix   string // The original uploads are stored
	SegmentsPrefix  string // Processed segments are stored
	ManifestsPrefix string // HLS manifests are stored
}

func DefaultProcessingConfig() *ProcessingConfig {
	return &ProcessingConfig{
		FFmpegPath:      "/opt/homebrew/bin/ffmpeg",
		TempDir:         filepath.Join("/tmp", "video-processing"),
		SegmentDuration: 10,
		Resolutions:     models.AllResolutions,
		S3Path: S3PathConfig{
			UploadsPrefix:   "uploads",
			SegmentsPrefix:  "segments",
			ManifestsPrefix: "manifests",
		},
	}
}

func (cfg *ProcessingConfig) GetSegmentPath(videoID string, segmentIndex int, resolution string) string {
	filename := fmt.Sprintf("segment-%d.ts", segmentIndex)
	return filepath.Join(
		cfg.S3Path.SegmentsPrefix,
		videoID,
		resolution,
		filename)

}

func (cfg *ProcessingConfig) GetManifestPath(videoID string, resolution string) string {
	filename := fmt.Sprintf("%s.m3u8", resolution)
	if resolution == "" {
		return filepath.Join(
			cfg.S3Path.ManifestsPrefix,
			videoID,
			"master.m3u8",
		)
	}
	return filepath.Join(cfg.S3Path.ManifestsPrefix, videoID, filename)
}

func (cfg *ProcessingConfig) GetTempVideoPath(videoID string) string {
	return filepath.Join(cfg.TempDir, videoID, "original.mp4")
}

func (cfg *ProcessingConfig) GetTempSegmentPath(videoID string, segmentIndex int) string {
	filename := fmt.Sprintf("segment_%d.ts", segmentIndex)
	return filepath.Join(cfg.TempDir, videoID, "segments", filename)
}

func (cfg *ProcessingConfig) GetTempTranscodedPath(videoID string, segmentIndex int, resolution string) string {
	filename := fmt.Sprintf("segment_%d_%s.ts", segmentIndex, resolution)
	return filepath.Join(cfg.TempDir, videoID, "transcoded", resolution, filename)
}
