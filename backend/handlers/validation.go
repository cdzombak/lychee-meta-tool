package handlers

import (
	"regexp"
	"strings"
)

const (
	// API path constants
	PhotosAPIPrefix = "/api/photos/"
	PhotosAPIPrefixLen = 12

	// Query parameter limits
	DefaultLimit = 50
	MaxLimit = 100
	MinOffset = 0
	
	// Photo ID validation
	MinPhotoIDLength = 1
	MaxPhotoIDLength = 64
)

var (
	// Photo ID validation patterns
	// Lychee uses various ID formats including UUIDs, timestamps, and custom formats
	photoIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// validatePhotoID validates that a photo ID is in an acceptable format
func validatePhotoID(id string) bool {
	if len(id) < MinPhotoIDLength || len(id) > MaxPhotoIDLength {
		return false
	}
	
	// Remove any potential file extensions for validation
	cleanID := strings.TrimSuffix(id, ".jpg")
	cleanID = strings.TrimSuffix(cleanID, ".jpeg")
	cleanID = strings.TrimSuffix(cleanID, ".png")
	cleanID = strings.TrimSuffix(cleanID, ".gif")
	cleanID = strings.TrimSuffix(cleanID, ".webp")
	
	return photoIDPattern.MatchString(cleanID)
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