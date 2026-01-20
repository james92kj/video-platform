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
	S3Key            *string   `json:"s3_key,omitempty"`
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

type GetUploadUrlRequest struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}
type UploadURLResponse struct {
	Success   bool   `json:"success"`
	UploadURL string `json:"upload_url"`
	VideoID   string `json:"video_id"`
	Key       string `json:"key"`
	ExpiresIn int64  `json:"expires_in"`
	Message   string `json:"message"`
}

type ProcessingJob struct {
	VideoID       string         `json:"video_id"`
	OriginalS3Key string         `json:"original_s3_key"`
	Status        string         `json:"status"`
	Segments      []VideoSegment `json:"segments"`
	Resolutions   []string       `json:"resolutions"`
	ManifestURL   string         `json:"manifest_url"`
	Error         string         `json:"error,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type VideoSegment struct {
	Index         int                          `json:"index"`
	Duration      float64                      `json:"duration"`
	Originals3key string                       `json:"original_s3_key"`
	Size          int64                        `json:"size"`
	Transcoded    map[string]TranscodingResult `json:"transcoded"`
}

type TranscodingResult struct {
	Resolution string `json:"resolution"`
	Bitrate    int    `json:"bitrate"`
	S3Key      string `json:"s3_key"`
	Size       int64  `json:"size"`
	Success    bool   `json:"success"`
	Error      string `json:"error"`
}

type Resolution struct {
	Name    string `json:"name"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Bitrate uint   `json:"bitrate"`
	Codec   string `json:"codec"`
}

/*

	type S3PathConfig struct {
    	UploadsPrefix   string // "uploads/"
    	SegmentsPrefix  string // "segments/"
    	ManifestsPrefix string // "manifests/"
	}
```

**Example S3 Structure:**
```
my-video-bucket/
├── uploads/
│   └── video-123.mp4                    (original upload)
├── segments/
│   └── video-123/
│       ├── 240p/
│       │   ├── segment_0.ts
│       │   ├── segment_1.ts
│       │   └── segment_2.ts
│       ├── 480p/
│       │   ├── segment_0.ts
│       │   ├── segment_1.ts
│       │   └── segment_2.ts
│       └── 720p/...
└── manifests/
    └── video-123/
        ├── master.m3u8                  (lists all qualities)
        ├── 240p.m3u8                    (240p playlist)
        ├── 480p.m3u8                    (480p playlist)
        └── 720p.m3u8                    (720p playlist)
*/

// Define predefine resolutions
var (
	Resolution240p = Resolution{
		Name:    "240p",
		Width:   426,
		Height:  240,
		Bitrate: 400,
		Codec:   "libx264",
	}

	Resolution480p = Resolution{
		Name:    "480p",
		Width:   854,
		Height:  480,
		Bitrate: 1000,
		Codec:   "libx264",
	}
	Resolution720p = Resolution{
		Name:    "720p",
		Width:   1280,
		Height:  720,
		Bitrate: 2500,
		Codec:   "libx264",
	}

	Resolution1080p = Resolution{
		Name:    "1080p",
		Width:   1920,
		Height:  1080,
		Bitrate: 5000,
		Codec:   "libx264",
	}

	AllResolutions = []Resolution{
		Resolution240p,
		Resolution480p,
		Resolution720p,
		Resolution1080p,
	}
)

func (j *ProcessingJob) UpdateStatus(status string) {
	j.Status = status
	j.UpdatedAt = time.Now()
}

func (j *ProcessingJob) SetError(err error) {
	j.Status = "error"
	j.Error = err.Error()
	j.UpdatedAt = time.Now()
}

func (j *ProcessingJob) AddSegment(segment VideoSegment) {
	j.Segments = append(j.Segments, segment)
	j.UpdatedAt = time.Now()
}
