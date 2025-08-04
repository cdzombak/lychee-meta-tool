// Package main provides the Lychee Meta Tool server application.
//
// The Lychee Meta Tool is a web-based tool for managing Lychee photo libraries,
// specifically designed for setting titles on photos that don't have them yet.
// It provides both a web interface and REST API for photo metadata management
// with optional AI-powered title generation using Ollama.
//
// Key features:
//   - Photo metadata editing (titles, descriptions)
//   - Album-based filtering
//   - AI-powered title suggestions via Ollama integration
//   - Support for MySQL, PostgreSQL, and SQLite databases
//   - Embedded web frontend for easy deployment
//
// Usage:
//   lychee-meta-tool -config config.yaml
//
// Configuration is provided via a YAML file specifying database connection,
// server settings, Lychee base URL, and optional Ollama configuration.
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cdzombak/lychee-meta-tool/backend/config"
	"github.com/cdzombak/lychee-meta-tool/backend/db"
	"github.com/cdzombak/lychee-meta-tool/backend/handlers"
	"github.com/cdzombak/lychee-meta-tool/backend/ollama"
)

// frontendFS embeds the built frontend assets into the binary.
// This allows the application to serve the web interface without
// requiring external files, enabling single-binary deployment.
//go:embed frontend/dist
var frontendFS embed.FS

// main is the entry point for the Lychee Meta Tool server.
// It handles configuration loading, database connection, optional
// Ollama client initialization, HTTP server setup, and graceful shutdown.
func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Printf("Connected to %s database", database.Driver())

	// Initialize Ollama client if configured
	var ollamaClient *ollama.Client
	if cfg.Ollama.URL != "" && cfg.Ollama.Model != "" {
		var err error
		ollamaClient, err = ollama.NewClient(cfg.Ollama.URL, cfg.Ollama.Model)
		if err != nil {
			log.Printf("Warning: Failed to initialize Ollama client: %v", err)
			log.Printf("AI title generation will be disabled")
		} else {
			log.Printf("Ollama client initialized with model %s at %s", cfg.Ollama.Model, cfg.Ollama.URL)
		}
	}

	photoHandler := handlers.NewPhotoHandler(database, cfg.LycheeBaseURL, ollamaClient)
	albumHandler := handlers.NewAlbumHandler(database)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/photos/needsmetadata", photoHandler.GetPhotosNeedingMetadata)
	mux.HandleFunc("/api/photos/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/generate-title") && r.Method == http.MethodPost {
			photoHandler.GenerateAITitle(w, r)
		} else if r.Method == http.MethodPut {
			photoHandler.UpdatePhoto(w, r)
		} else {
			photoHandler.GetPhotoByID(w, r)
		}
	})
	mux.HandleFunc("/api/albums", albumHandler.GetAlbums)
	mux.HandleFunc("/api/albums/withphotocounts", albumHandler.GetAlbumsWithPhotoCounts)

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Health(); err != nil {
			http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Serve frontend static files from embedded filesystem
	// Since we embedded frontend/dist, we need to create a sub-filesystem from that path
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to create dist sub filesystem: %v", err)
	}

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For SPA, serve index.html for any non-API route that doesn't exist
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the requested file
		if _, err := distFS.Open(strings.TrimPrefix(r.URL.Path, "/")); err == nil {
			http.FileServer(http.FS(distFS)).ServeHTTP(w, r)
			return
		}

		// If file doesn't exist, serve index.html (for SPA routing)
		r.URL.Path = "/"
		http.FileServer(http.FS(distFS)).ServeHTTP(w, r)
	}))

	// Add CORS middleware
	handler := corsMiddleware(mux, cfg.Server.CORS.AllowedOrigins)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func corsMiddleware(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed (only set CORS headers for allowed origins)
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			// Only allow methods actually used by the application
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours
		}

		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		if r.Method == "OPTIONS" {
			if allowed {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}