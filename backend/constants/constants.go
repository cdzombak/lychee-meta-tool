// Package constants defines shared constants used throughout the Lychee Meta Tool.
// This centralized approach ensures consistency and makes it easy to modify
// configuration values, timeouts, limits, and other magic numbers.
//
// The constants are organized into logical groups:
//   - HTTP-related constants (content types, methods)
//   - API path constants and patterns
//   - Database query limits and constraints  
//   - Timeout and duration settings
//   - File format and validation patterns
//   - Application metadata and defaults
package constants

import "time"

// HTTP Constants
const (
	// Content types
	ContentTypeJSON = "application/json"
	ContentTypeHTML = "text/html"
	ContentTypeText = "text/plain"

	// HTTP methods (for documentation/consistency)
	MethodGET    = "GET"
	MethodPOST   = "POST"
	MethodPUT    = "PUT"
	MethodDELETE = "DELETE"
)

// API Constants
const (
	// API path prefixes
	APIPrefix     = "/api"
	PhotosPrefix  = "/api/photos"
	AlbumsPrefix  = "/api/albums"
	HealthPrefix  = "/health"

	// API path suffixes
	GenerateTitleSuffix = "/generate-title"
	WithPhotoCountsSuffix = "/withphotocounts"
	NeedsMetadataSuffix = "/needsmetadata"
)

// Database Constants
const (
	// Query limits
	DefaultPhotoLimit = 1000
	MaxPhotoLimit     = 1000
	MinPhotoOffset    = 0

	// ID constraints
	MinIDLength = 1
	MaxIDLength = 64

	// Text field limits
	MaxPhotoTitleLength       = 255
	MaxPhotoDescriptionLength = 2000
)

// Timeout Constants
const (
	// HTTP timeouts
	DefaultHTTPTimeout = 30 * time.Second
	ImageDownloadTimeout = 30 * time.Second

	// AI generation timeouts
	AIGenerationTimeout = 2 * time.Minute
	OllamaClientTimeout = 5 * time.Minute

	// Database timeouts
	DatabaseConnectionTimeout = 10 * time.Second
	DatabaseQueryTimeout     = 30 * time.Second
)

// File and Image Constants
const (
	// Image file extensions
	ExtJPG  = ".jpg"
	ExtJPEG = ".jpeg"
	ExtPNG  = ".png"
	ExtGIF  = ".gif"
	ExtWEBP = ".webp"
	ExtTIFF = ".tiff"
	ExtBMP  = ".bmp"

	// MIME types
	MimeJPEG = "image/jpeg"
	MimePNG  = "image/png"
	MimeGIF  = "image/gif"
	MimeWEBP = "image/webp"

	// File size limits
	MaxImageSize = 5 * 1024 * 1024 // 5MB
)

// Application Constants
const (
	// Application metadata
	AppName    = "lychee-meta-tool"
	AppVersion = "1.0.0"

	// Environment variables
	EnvConfigPath = "CONFIG_PATH"
	EnvLogLevel   = "LOG_LEVEL"
	EnvPort       = "PORT"
)

// Validation Constants
const (
	// Pattern names for validation
	PhotoIDPattern = `^[a-zA-Z0-9_-]+$`
	AlbumIDPattern = `^[a-zA-Z0-9_-]+$`

	// Validation error templates
	ErrInvalidIDFormat     = "invalid %s format (must be %d-%d characters, alphanumeric with underscores and hyphens)"
	ErrTextTooLong         = "%s too long (max %d characters, got %d)"
	ErrInvalidUTF8         = "%s contains invalid UTF-8 characters"
	ErrDangerousContent    = "%s contains potentially dangerous content"
	ErrRequiredField       = "%s is required"
	ErrInvalidRange        = "%s must be between %d and %d, got %d"
)

// Log Message Templates
const (
	LogPhotoUpdate          = "Updated photo %s with fields: %+v"
	LogAITitleGeneration    = "Generated AI title for photo %s: %s"
	LogImageDownload        = "Downloaded image: Content-Type=%s, Status=%d, URL=%s"
	LogDatabaseOperation    = "Database operation %s completed in %v"
	LogValidationFailed     = "Validation failed for %s: %v"
	LogOllamaClientCreated  = "Ollama client configured with URL: %s, Model: %s"
	LogServerStarted        = "Server started on port %d"
	LogConfigLoaded         = "Configuration loaded from %s"
)

// Configuration Defaults
const (
	DefaultServerPort    = 8080
	DefaultDatabasePort  = 3306
	DefaultPostgresPort  = 5432
	DefaultOllamaPort    = 11434
	DefaultLogLevel      = "info"
	DefaultConfigPath    = "config.yaml"
)
