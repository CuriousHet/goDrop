package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"dcas/config"
	"dcas/storage"
	"dcas/web"
)

func main() {
	// Load .env file only in local dev
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("⚠️ Warning: Could not load .env file: %v", err)
		} else {
			log.Println("✅ .env file loaded (local dev)")
		}
	} else {
		log.Println("ℹ️ .env file not found — using injected environment variables")
	}

	// Parse command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	mode := flag.String("mode", "web", "Mode to run in (web or node)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Override config port with command line flag
	cfg.Port = *port

	// Validate supported storage type
	if cfg.Storage.Type != "s3" {
		log.Fatalf("❌ Unsupported storage type '%s'. Only 's3' is supported.", cfg.Storage.Type)
	}

	// Initialize S3 storage
	store, err := storage.NewS3Storage(
		cfg.Storage.Endpoint,
		cfg.Storage.AccessKey,
		cfg.Storage.SecretKey,
		cfg.Storage.Bucket,
		cfg.Storage.UseSSL,
	)
	if err != nil {
		log.Fatalf("❌ Failed to initialize storage: %v", err)
	}

	// Create and start the web server
	server := web.NewServer(*cfg, store)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("🚀 Starting server on %s in %s mode", addr, *mode)

	if err := server.Start(addr); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
