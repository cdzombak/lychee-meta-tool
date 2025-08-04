package handlers

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/cdzombak/lychee-meta-tool/backend/constants"
	"github.com/cdzombak/lychee-meta-tool/backend/models"
)

const (
	// API path constants
	PhotosAPIPrefix = "/api/photos/"
	PhotosAPIPrefixLen = 12

	// Query parameter limits (using constants)
	DefaultLimit = constants.DefaultPhotoLimit
	MaxLimit = constants.MaxPhotoLimit
	MinOffset = constants.MinPhotoOffset
	
	// ID validation (using constants)
	MinPhotoIDLength = constants.MinIDLength
	MaxPhotoIDLength = constants.MaxIDLength

	// Text field limits (using constants)
	MaxTitleLength = constants.MaxPhotoTitleLength
	MaxDescriptionLength = constants.MaxPhotoDescriptionLength
	MaxAlbumIDLength = constants.MaxIDLength

	// Content validation
	MinContentLength = 0
)

var (
	// Validation patterns
	photoIDPattern = regexp.MustCompile(constants.PhotoIDPattern)
	albumIDPattern = regexp.MustCompile(constants.AlbumIDPattern)
	
	// Dangerous patterns to detect potential security issues
	scriptTagPattern = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	javascriptPattern = regexp.MustCompile(`(?i)javascript:`)
	dangerousHTMLPattern = regexp.MustCompile(`(?i)<[^>]*on\w+\s*=`)
)

// validatePhotoID validates that a photo ID is in an acceptable format
func validatePhotoID(id string) bool {
	if len(id) < MinPhotoIDLength || len(id) > MaxPhotoIDLength {
		return false
	}
	
	// Check for valid UTF-8
	if !utf8.ValidString(id) {
		return false
	}
	
	// Remove any potential file extensions for validation
	cleanID := removePotentialExtensions(id)
	
	return photoIDPattern.MatchString(cleanID)
}

// validateAlbumID validates that an album ID is in an acceptable format
func validateAlbumID(id string) bool {
	if id == "" {
		return true // Empty album ID is valid (means no album)
	}
	
	if len(id) < MinPhotoIDLength || len(id) > MaxAlbumIDLength {
		return false
	}
	
	if !utf8.ValidString(id) {
		return false
	}
	
	return albumIDPattern.MatchString(id)
}

// removePotentialExtensions removes common image file extensions from an ID
func removePotentialExtensions(id string) string {
	extensions := []string{
		constants.ExtJPG, constants.ExtJPEG, constants.ExtPNG,
		constants.ExtGIF, constants.ExtWEBP, constants.ExtTIFF, constants.ExtBMP,
	}
	cleanID := id
	for _, ext := range extensions {
		cleanID = strings.TrimSuffix(cleanID, ext)
		cleanID = strings.TrimSuffix(cleanID, strings.ToUpper(ext))
	}
	return cleanID
}

// extractPhotoIDFromPath safely extracts photo ID from URL path
func extractPhotoIDFromPath(path string) (string, bool) {
	if len(path) < PhotosAPIPrefixLen {
		return "", false
	}
	
	photoID := path[PhotosAPIPrefixLen:]
	if photoID == "" {
		return "", false
	}
	
	// Remove any trailing slash or additional path components
	if slashIndex := strings.Index(photoID, "/"); slashIndex != -1 {
		photoID = photoID[:slashIndex]
	}
	
	if !validatePhotoID(photoID) {
		return "", false
	}
	
	return photoID, true
}

// validateLimit ensures the limit parameter is within acceptable bounds
func validateLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

// validateOffset ensures the offset parameter is valid
func validateOffset(offset int) int {
	if offset < MinOffset {
		return MinOffset
	}
	return offset
}

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// Error implements the error interface
func (v ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", v.Field, v.Message)
}

// ValidatePhotoUpdate validates a PhotoUpdate struct
func ValidatePhotoUpdate(update *models.PhotoUpdate) []ValidationError {
	var errors []ValidationError

	// Validate title
	if update.Title != nil {
		if err := validateAndSanitizeTitle(*update.Title); err != nil {
			errors = append(errors, ValidationError{Field: "title", Message: err.Error(), Value: *update.Title})
		} else {
			// Update with sanitized value
			sanitized := sanitizeText(*update.Title)
			update.Title = &sanitized
		}
	}

	// Validate description
	if update.Description != nil {
		if err := validateAndSanitizeDescription(*update.Description); err != nil {
			errors = append(errors, ValidationError{Field: "description", Message: err.Error(), Value: *update.Description})
		} else {
			// Update with sanitized value
			sanitized := sanitizeText(*update.Description)
			update.Description = &sanitized
		}
	}

	// Validate album ID
	if update.AlbumID != nil {
		if !validateAlbumID(*update.AlbumID) {
			errors = append(errors, ValidationError{
				Field:   "album_id",
				Message: fmt.Sprintf("invalid album ID format (length: %d-%d, pattern: alphanumeric, underscore, hyphen)", MinPhotoIDLength, MaxAlbumIDLength),
				Value:   *update.AlbumID,
			})
		}
	}

	return errors
}

// validateAndSanitizeTitle validates a photo title
func validateAndSanitizeTitle(title string) error {
	if !utf8.ValidString(title) {
		return fmt.Errorf("title contains invalid UTF-8 characters")
	}

	if len(title) > MaxTitleLength {
		return fmt.Errorf("title too long (max %d characters, got %d)", MaxTitleLength, len(title))
	}

	if containsDangerousContent(title) {
		return fmt.Errorf("title contains potentially dangerous content")
	}

	return nil
}

// validateAndSanitizeDescription validates a photo description
func validateAndSanitizeDescription(description string) error {
	if !utf8.ValidString(description) {
		return fmt.Errorf("description contains invalid UTF-8 characters")
	}

	if len(description) > MaxDescriptionLength {
		return fmt.Errorf("description too long (max %d characters, got %d)", MaxDescriptionLength, len(description))
	}

	if containsDangerousContent(description) {
		return fmt.Errorf("description contains potentially dangerous content")
	}

	return nil
}

// containsDangerousContent checks for potentially dangerous content
func containsDangerousContent(text string) bool {
	return scriptTagPattern.MatchString(text) ||
		javascriptPattern.MatchString(text) ||
		dangerousHTMLPattern.MatchString(text)
}

// sanitizeText sanitizes user input by HTML escaping and trimming
func sanitizeText(text string) string {
	// Trim whitespace
	text = strings.TrimSpace(text)
	
	// HTML escape to prevent XSS
	text = html.EscapeString(text)
	
	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	
	return text
}

// sanitizeQueryParam sanitizes query parameters
func sanitizeQueryParam(param string) string {
	return strings.TrimSpace(html.EscapeString(param))
}