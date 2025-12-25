package storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/james92kj/video-platform/internal/config"
	"github.com/james92kj/video-platform/internal/logger"
	"time"
)

type S3Client struct {
	client *s3.Client
	bucket string
	log    *logger.Logger
}

func NewS3Client(log *logger.Logger) (*S3Client, error) {

	cfg := config.Load()

	// Create custom endpoint resolver for minio
	awsConfig := aws.Config{
		Region: cfg.S3Config.Region,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				cfg.S3Config.AccessKey, cfg.S3Config.SecretKey, ""),
		),
	}

	// Create s3 client with custom endpoint
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://" + cfg.S3Config.Endpoint)
		o.UsePathStyle = true
	})

	s3Client := &S3Client{
		client: client,
		bucket: cfg.S3Config.Bucket,
		log:    log,
	}

	log.Info("S3 client initialized successfully")
	return s3Client, nil
}

func (s *S3Client) GeneratePreSignedUrl(ctx context.Context, key string, expirationMinutes int) (string, error) {

	s.log.Info(fmt.Sprintf("Generating pre-signed url for %s with %d minutes", key, expirationMinutes))

	// Create presign client
	presignClient := s3.NewPresignClient(s.client)

	// Generate presigned put url
	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(o *s3.PresignOptions) {
		o.Expires = time.Duration(expirationMinutes) * time.Minute
	})

	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to generate pre-signed url: %v", err))
		return "", fmt.Errorf("failed to generate pre-signed url: %w", err)
	}

	s.log.Info(fmt.Sprintf("Generated pre-signed url: %s", request.URL))
	return request.URL, nil
}
