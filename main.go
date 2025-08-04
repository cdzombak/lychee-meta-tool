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
)

//go:embed frontend/dist
var frontendFS embed.FS

var version = "dev"

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("lychee-meta-tool %s\n", version)
		return
	}

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

	photoHandler := handlers.NewPhotoHandler(database, cfg.LycheeBaseURL)
	albumHandler := handlers.NewAlbumHandler(database)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/photos/needsmetadata", photoHandler.GetPhotosNeedingMetadata)
	mux.HandleFunc("/api/photos/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
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
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
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