package models

import (
	"regexp"
	"strings"
)

var (
	// Common camera naming patterns
	cameraPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^IMG_\d+(\.\w+)?$`),           // IMG_1234 or IMG_1234.jpg
		regexp.MustCompile(`^DSC_\d+(\.\w+)?$`),           // DSC_1234 or DSC_1234.jpg
		regexp.MustCompile(`^DSCN\d+(\.\w+)?$`),           // DSCN1234 or DSCN1234.jpg
		regexp.MustCompile(`^DSCF\d+(\.\w+)?$`),           // DSCF1234 or DSCF1234.jpg
		regexp.MustCompile(`^CDZ_\d+(\.\w+)?$`),           // CDZ_1234 or CDZ_1234.jpg
		regexp.MustCompile(`^P\d{7}(\.\w+)?$`),            // P1234567 or P1234567.jpg
		regexp.MustCompile(`^\d{8}_\d{6}(\.\w+)?$`),       // 20230101_123456 or 20230101_123456.jpg
		regexp.MustCompile(`^IMG-\d{8}-WA\d{4}(\.\w+)?$`), // WhatsApp format
		regexp.MustCompile(`^Screenshot.*(\.\w+)?$`),      // Screenshot files
	}

	// UUID pattern (with or without dashes, with optional file extension)
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{12}(\.\w+)?$`)
)

func IsGenericTitle(title string) bool {
	if title == "" {
		return true
	}

	// Remove leading/trailing whitespace
	title = strings.TrimSpace(title)

	if title == "" {
		return true
	}

	// Check for UUID patterns
	if uuidPattern.MatchString(title) {
		return true
	}

	// Check for camera naming patterns
	for _, pattern := range cameraPatterns {
		if pattern.MatchString(title) {
			return true
		}
	}

	// Check for prefix "IDG_" indicating the image is named for the Adobe Indigo camera app
	if strings.HasPrefix(title, "IDG_") {
		return true
	}

	return false
}
