package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SBOMStorage defines the interface for SBOM storage operations
type SBOMStorage interface {
	Store(ctx context.Context, scanID int, document []byte) error
	Retrieve(ctx context.Context, scanID int) ([]byte, error)
	Delete(ctx context.Context, scanID int) error
	GetPresignedURL(ctx context.Context, scanID int, expiresIn time.Duration) (string, error)
	Exists(ctx context.Context, scanID int) (bool, error)
}

// S3Storage implements SBOMStorage using S3-compatible object storage
type S3Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

// NewS3Storage creates a new S3-based SBOM storage
func NewS3Storage(client *s3.Client, bucket string) *S3Storage {
	return &S3Storage{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        bucket,
	}
}

// computePath generates the S3 key for a given scan ID
// Pattern: scans/{scan_id}/sbom.json
func (s *S3Storage) computePath(scanID int) string {
	return fmt.Sprintf("scans/%d/sbom.json", scanID)
}

// Store uploads an SBOM document to S3
func (s *S3Storage) Store(ctx context.Context, scanID int, document []byte) error {
	path := s.computePath(scanID)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        bytes.NewReader(document),
		ContentType: aws.String("application/json"),
	})

	return err
}

// Retrieve downloads an SBOM document from S3
func (s *S3Storage) Retrieve(ctx context.Context, scanID int) ([]byte, error) {
	path := s.computePath(scanID)

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve SBOM from S3: %w", err)
	}
	defer result.Body.Close()

	document, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM document: %w", err)
	}

	return document, nil
}

// Delete removes an SBOM document from S3
func (s *S3Storage) Delete(ctx context.Context, scanID int) error {
	path := s.computePath(scanID)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	return err
}

// GetPresignedURL generates a pre-signed URL for direct SBOM download
func (s *S3Storage) GetPresignedURL(ctx context.Context, scanID int, expiresIn time.Duration) (string, error) {
	path := s.computePath(scanID)

	request, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// Exists checks if an SBOM document exists in S3
func (s *S3Storage) Exists(ctx context.Context, scanID int) (bool, error) {
	path := s.computePath(scanID)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		// Check if it's a "not found" error
		return false, nil
	}

	return true, nil
}
