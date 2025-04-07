package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileStorage implements Storage interface for local file system
type FileStorage struct {
	storageDir string
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(storageDir string) (*FileStorage, error) {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	return &FileStorage{
		storageDir: storageDir,
	}, nil
}

// Store stores a file in the local file system
func (s *FileStorage) Store(filename string, reader io.Reader) error {
	filepath := filepath.Join(s.storageDir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// Get retrieves a file from the local file system
func (s *FileStorage) Get(filename string) (io.ReadCloser, error) {
	filepath := filepath.Join(s.storageDir, filename)
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	return file, nil
}

// Delete removes a file from the local file system
func (s *FileStorage) Delete(filename string) error {
	filepath := filepath.Join(s.storageDir, filename)
	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}

// Exists checks if a file exists in the local file system
func (s *FileStorage) Exists(filename string) (bool, error) {
	filepath := filepath.Join(s.storageDir, filename)
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %v", err)
	}
	return true, nil
}

// List returns a list of all files in storage
func (s *FileStorage) List() ([]string, error) {
	// Read all files in the storage directory
	files, err := os.ReadDir(s.storageDir)
	if err != nil {
		return nil, fmt.Errorf("error reading storage directory: %v", err)
	}

	// Convert to string slice
	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
} 