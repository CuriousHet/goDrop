package main

import (
	"flag"
	"fmt"
	"log"

	"dcas/config"
	"dcas/storage"
	"dcas/web"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	mode := flag.String("mode", "web", "Mode to run in (web or node)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override config with command line flags
	cfg.Port = *port

	// Initialize storage
	if cfg.Storage.Type != "s3" {
		log.Fatalf("Only s3 storage type is supported")
	}

	store, err := storage.NewS3Storage(
		cfg.Storage.Endpoint,
		cfg.Storage.AccessKey,
		cfg.Storage.SecretKey,
		cfg.Storage.Bucket,
		cfg.Storage.UseSSL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create web server
	server := web.NewServer(*cfg, store)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s in %s mode", addr, *mode)
	if err := server.Start(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
