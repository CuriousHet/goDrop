package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"D-CAS/config"
)

type mockStorage struct {
	files map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		files: make(map[string][]byte),
	}
}

func (m *mockStorage) Store(ctx context.Context, reader io.Reader, filename string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	hash := "test-hash" // In a real implementation, this would be calculated
	m.files[hash] = data
	return hash, nil
}

func (m *mockStorage) Get(ctx context.Context, hash string) (io.ReadCloser, error) {
	data, exists := m.files[hash]
	if !exists {
		return nil, os.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *mockStorage) Delete(ctx context.Context, hash string) error {
	delete(m.files, hash)
	return nil
}

func (m *mockStorage) Exists(ctx context.Context, hash string) (bool, error) {
	_, exists := m.files[hash]
	return exists, nil
}

func TestUploadHandler(t *testing.T) {
	// Create test server
	store := newMockStorage()
	cfg := config.Config{Port: 8080}
	server := NewServer(cfg, store)

	// Create test file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("test content"))
	writer.Close()

	// Create request
	req := httptest.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	server.handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["hash"] != "test-hash" {
		t.Errorf("Expected hash 'test-hash', got %s", response["hash"])
	}
}

func TestSearchHandler(t *testing.T) {
	// Create test server
	store := newMockStorage()
	cfg := config.Config{Port: 8080}
	server := NewServer(cfg, store)

	// Create test file
	store.Store(context.Background(), bytes.NewReader([]byte("test content")), "test.txt")

	// Create request
	req := httptest.NewRequest("GET", "/api/search?q=test-hash", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["hash"] != "test-hash" {
		t.Errorf("Expected hash 'test-hash', got %s", response["hash"])
	}
}

func TestDownloadHandler(t *testing.T) {
	// Create test server
	store := newMockStorage()
	cfg := config.Config{Port: 8080}
	server := NewServer(cfg, store)

	// Create test file
	store.Store(context.Background(), bytes.NewReader([]byte("test content")), "test.txt")

	// Create request
	req := httptest.NewRequest("GET", "/api/download/test-hash", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test content" {
		t.Errorf("Expected content 'test content', got %s", w.Body.String())
	}
}

func TestCodeWordMapping(t *testing.T) {
	// Create test server
	store := newMockStorage()
	cfg := config.Config{Port: 8080}
	server := NewServer(cfg, store)

	// Create test file with code word
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("test content"))
	writer.WriteField("code_word", "test-code")
	writer.Close()

	// Upload file
	req := httptest.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	server.handler.ServeHTTP(w, req)

	// Search by code word
	req = httptest.NewRequest("GET", "/api/search?q=test-code", nil)
	w = httptest.NewRecorder()
	server.handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["hash"] != "test-hash" {
		t.Errorf("Expected hash 'test-hash', got %s", response["hash"])
	}
	if response["code_word"] != "test-code" {
		t.Errorf("Expected code word 'test-code', got %s", response["code_word"])
	}
} 