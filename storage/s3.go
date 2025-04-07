package storage

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Storage implements Storage interface for S3-compatible storage
type S3Storage struct {
	client *s3.S3
	bucket string
}

// NewS3Storage creates a new S3Storage instance for R2
func NewS3Storage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3Storage, error) {
	log.Printf("Initializing S3 storage with endpoint: %s, bucket: %s", endpoint, bucket)

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("auto"),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true), // Required for Cloudflare R2
		DisableSSL:       aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	// Create S3 client
	client := s3.New(sess)

	// Test bucket access
	_, err = client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to access bucket: %v", err)
	}

	log.Printf("Successfully initialized S3 storage")
	return &S3Storage{
		client: client,
		bucket: bucket,
	}, nil
}

// Store stores a file in S3
func (s *S3Storage) Store(filename string, reader io.Reader) error {
	// Get the file extension from the original filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".bin" // Default extension for files without one
	}

	// Create the storage filename with extension
	storageFilename := filename
	if !strings.HasSuffix(storageFilename, ext) {
		storageFilename = filename + ext
	}

	log.Printf("Storing file: %s with extension: %s", storageFilename, ext)

	// Convert io.Reader to io.ReadSeeker by reading all data into memory
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read file data: %v", err)
	}

	// Determine content type based on extension
	contentType := "application/octet-stream"
	switch strings.ToLower(ext) {
	case ".pdf":
		contentType = "application/pdf"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".txt":
		contentType = "text/plain"
	case ".html", ".htm":
		contentType = "text/html"
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "application/javascript"
	case ".json":
		contentType = "application/json"
	case ".xml":
		contentType = "application/xml"
	case ".zip":
		contentType = "application/zip"
	case ".doc", ".docx":
		contentType = "application/msword"
	case ".xls", ".xlsx":
		contentType = "application/vnd.ms-excel"
	case ".ppt", ".pptx":
		contentType = "application/vnd.ms-powerpoint"
	}

	// Create a bytes.Reader from the data
	readSeeker := bytes.NewReader(data)

	_, err = s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(storageFilename),
		Body:        readSeeker,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to store file: %v", err)
	}
	log.Printf("Successfully stored file: %s", storageFilename)
	return nil
}

// Get retrieves a file from S3
func (s *S3Storage) Get(filename string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %v", err)
	}

	return result.Body, nil
}

// Delete removes a file from S3
func (s *S3Storage) Delete(filename string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}

// Exists checks if a file exists in S3
func (s *S3Storage) Exists(filename string) (bool, error) {
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %v", err)
	}

	return true, nil
}

// List returns a list of all files in storage
func (s *S3Storage) List() ([]string, error) {
	log.Printf("Listing objects in bucket: %s", s.bucket)
	
	var files []string
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	}

	// List objects in the bucket
	resp, err := s.client.ListObjectsV2(input)
	if err != nil {
		log.Printf("Error listing objects: %v", err)
		return nil, fmt.Errorf("error listing objects in bucket: %v", err)
	}

	// Collect all file names
	for _, obj := range resp.Contents {
		if obj.Key != nil {
			// Remove bucket prefix if present
			key := *obj.Key
			if strings.HasPrefix(key, s.bucket+"/") {
				key = strings.TrimPrefix(key, s.bucket+"/")
			}
			files = append(files, key)
		}
	}

	log.Printf("Found %d files in bucket", len(files))
	return files, nil
} 