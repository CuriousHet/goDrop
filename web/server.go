package web

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"dcas/config"
	"dcas/storage"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB
)

type Server struct {
	config *config.Config
	store  storage.Storage
	mux    *http.ServeMux
}

func NewServer(cfg config.Config, store storage.Storage) *Server {
	s := &Server{
		config: &cfg,
		store:  store,
		mux:    http.NewServeMux(),
	}

	// Register routes
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/upload", s.handleUpload)
	s.mux.HandleFunc("/receive", s.handleReceive)
	s.mux.HandleFunc("/download/", s.handleDownload)

	// Serve static files
	fs := http.FileServer(http.Dir("web/static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	return s
}

func (s *Server) Start(addr string) error {
	if addr == "" {
		addr = fmt.Sprintf(":%d", s.config.Port)
	}
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with 10MB max size
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error getting file from form: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > 10<<20 { // 10MB
		http.Error(w, "File too large. Maximum size is 10MB", http.StatusBadRequest)
		return
	}

	// Get code word from form
	codeWord := r.FormValue("code_word")

	// Generate file code (hash or filename)
	var fileCode string
	// Always generate SHA-256 hash of the file
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Printf("Error generating hash: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fileCode = hex.EncodeToString(hash.Sum(nil))

	// Reset file reader for storage
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// Store file with code word if provided
	var storageFilename string
	fileExt := filepath.Ext(header.Filename)
	if codeWord != "" {
		storageFilename = fmt.Sprintf("%s_%s%s", codeWord, fileCode, fileExt)
	} else {
		storageFilename = fileCode + fileExt
	}

	// Store the file
	if err := s.store.Store(storageFilename, file); err != nil {
		log.Printf("Error storing file: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response with file information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "File uploaded successfully",
		"filename": header.Filename,
		"size": header.Size,
		"type": header.Header.Get("Content-Type"),
		"file_hash": fileCode,
		"code_word": codeWord,
	})
}

func (s *Server) handleReceive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// First, list all files in the bucket for debugging
	log.Printf("Listing all files in bucket...")
	files, err := s.store.List()
	if err != nil {
		log.Printf("Error listing files: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Files found in bucket:")
	for _, file := range files {
		log.Printf("- %s", file)
	}

	// Get file hash and code word from form
	fileHash := r.FormValue("file_hash")
	codeWord := r.FormValue("code_word")

	// If neither hash nor code word is provided
	if fileHash == "" && codeWord == "" {
		http.Error(w, "Either file hash or code word is required", http.StatusBadRequest)
		return
	}

	// Try to find the file using the provided information
	var storageFilename string
	if codeWord != "" && fileHash != "" {
		// Try exact match with code and hash
		storageFilename = fmt.Sprintf("%s_%s", codeWord, fileHash)
	} else if codeWord != "" {
		// Try to find file by code word prefix
		storageFilename = codeWord + "_"
	} else if fileHash != "" {
		// Try to find file by hash
		storageFilename = fileHash
	}

	// Log the search attempt
	log.Printf("Searching for file with pattern: %s", storageFilename)

	// Check if file exists
	exists, err := s.store.Exists(storageFilename)
	if err != nil {
		log.Printf("Error checking file existence: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If file not found with exact pattern, try to find it with any extension
	if !exists {
		// List all files in storage
		files, err := s.store.List()
		if err != nil {
			log.Printf("Error listing files: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Try to find a matching file
		found := false
		log.Printf("Searching through %d files for a match", len(files))
		
		for _, file := range files {
			// Remove extension for comparison
			fileWithoutExt := strings.TrimSuffix(file, filepath.Ext(file))
			log.Printf("Checking file: %s (without ext: %s)", file, fileWithoutExt)
			
			if codeWord != "" && fileHash != "" {
				// Try to match both code and hash
				expectedPrefix := codeWord + "_" + fileHash
				if strings.HasPrefix(fileWithoutExt, expectedPrefix) {
					storageFilename = file
					found = true
					log.Printf("Found match with code word and hash: %s", file)
					break
				}
			} else if codeWord != "" {
				// Try to match code word prefix
				if strings.HasPrefix(fileWithoutExt, codeWord+"_") {
					storageFilename = file
					found = true
					log.Printf("Found match with code word: %s", file)
					break
				}
			} else if fileHash != "" {
				// Try to match hash - check if the file contains the hash
				// This handles cases where the hash might be part of a more complex filename
				// Also check for common prefixes like "res_", "file_", etc.
				if strings.Contains(fileWithoutExt, fileHash) || 
				   strings.Contains(fileWithoutExt, "res_"+fileHash) || 
				   strings.Contains(fileWithoutExt, "file_"+fileHash) {
					storageFilename = file
					found = true
					log.Printf("Found match with hash: %s", file)
					break
				}
			}
		}

		if !found {
			log.Printf("File not found. Search criteria - Hash: %s, Code Word: %s", fileHash, codeWord)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
	}

	log.Printf("Found file: %s", storageFilename)

	// Get file from storage
	reader, err := s.store.Get(storageFilename)
	if err != nil {
		log.Printf("Error getting file: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// Extract the original filename from the storage filename
	originalFilename := filepath.Base(storageFilename)
	
	// If the original filename has a prefix like "res_" or "file_", try to extract a more meaningful name
	if strings.Contains(originalFilename, "_") {
		// This might be a prefix_hash.ext format
		parts := strings.SplitN(originalFilename, "_", 2)
		if len(parts) == 2 {
			// Try to extract a more meaningful name from the second part
			hashPart := parts[1]
			// Remove the extension
			hashPart = strings.TrimSuffix(hashPart, filepath.Ext(hashPart))
			// If it's a long hash, use a shorter version for the filename
			if len(hashPart) > 16 {
				hashPart = hashPart[:16] + "..."
			}
			originalFilename = "file_" + hashPart + filepath.Ext(originalFilename)
		}
	}
	
	// Determine content type based on file extension
	contentType := "application/octet-stream"
	ext := strings.ToLower(filepath.Ext(originalFilename))
	switch ext {
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

	// Set response headers for direct download with original filename and content type
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", originalFilename))
	w.Header().Set("Content-Type", contentType)

	// Copy file to response
	if _, err := io.Copy(w, reader); err != nil {
		log.Printf("Error copying file to response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get file hash from URL
	fileHash := filepath.Base(r.URL.Path)
	if fileHash == "" {
		http.Error(w, "File hash required", http.StatusBadRequest)
		return
	}

	// Check if file exists
	exists, err := s.store.Exists(fileHash)
	if err != nil {
		log.Printf("Error checking file existence: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Get file from storage
	reader, err := s.store.Get(fileHash)
	if err != nil {
		log.Printf("Error getting file: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// Determine content type based on file extension
	contentType := "application/octet-stream"
	ext := strings.ToLower(filepath.Ext(fileHash))
	switch ext {
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

	// Set response headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileHash))
	w.Header().Set("Content-Type", contentType)

	// Copy file to response
	if _, err := io.Copy(w, reader); err != nil {
		log.Printf("Error copying file to response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) saveCodeWord(fileHash, codeWord string) error {
	// Create code words directory if it doesn't exist
	if err := os.MkdirAll("code_words", 0755); err != nil {
		return err
	}

	// Save code word to file
	file, err := os.Create(filepath.Join("code_words", codeWord))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fileHash)
	return err
}

func (s *Server) getHashFromCodeWord(codeWord string) (string, error) {
	// Read hash from file
	file, err := os.Open(filepath.Join("code_words", codeWord))
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(hash), nil
} 