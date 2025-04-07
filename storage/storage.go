package storage

import "io"

// Storage defines the interface for file storage operations
type Storage interface {
	Store(filename string, reader io.Reader) error
	Get(filename string) (io.ReadCloser, error)
	Delete(filename string) error
	Exists(filename string) (bool, error)
	List() ([]string, error)
} 