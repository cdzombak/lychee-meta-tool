// Package models defines the data structures used throughout the Lychee Meta Tool.
// These models represent photos, albums, and related metadata as stored in the
// Lychee database and as exchanged via the REST API.
package models

import (
	"strings"
	"time"
)

// Photo represents a photo record from the Lychee database.
// It contains all metadata fields including EXIF data, location information,
// and user-provided metadata like title and description.
type Photo struct {
	ID           string     `json:"id" db:"id"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	OwnerID      int        `json:"owner_id" db:"owner_id"`
	AlbumID      *string    `json:"album_id" db:"old_album_id"`
	Title        string     `json:"title" db:"title"`
	Description  *string    `json:"description" db:"description"`
	License      string     `json:"license" db:"license"`
	IsStarred    bool       `json:"is_starred" db:"is_starred"`
	ISO          *string    `json:"iso" db:"iso"`
	Make         *string    `json:"make" db:"make"`
	Model        *string    `json:"model" db:"model"`
	Lens         *string    `json:"lens" db:"lens"`
	Aperture     *string    `json:"aperture" db:"aperture"`
	Shutter      *string    `json:"shutter" db:"shutter"`
	Focal        *string    `json:"focal" db:"focal"`
	Latitude     *float64   `json:"latitude" db:"latitude"`
	Longitude    *float64   `json:"longitude" db:"longitude"`
	Altitude     *float64   `json:"altitude" db:"altitude"`
	ImgDirection *float64   `json:"img_direction" db:"img_direction"`
	Location     *string    `json:"location" db:"location"`
	TakenAt      *time.Time `json:"taken_at" db:"taken_at"`
	Type         string     `json:"type" db:"type"`
	Filesize     int64      `json:"filesize" db:"filesize"`
	Checksum     string     `json:"checksum" db:"checksum"`
}

// PhotoWithAlbum extends Photo with album information.
// This is used when retrieving photos along with their album details.
type PhotoWithAlbum struct {
	Photo
	AlbumTitle *string `json:"album_title" db:"album_title"`
}

// PhotoUpdate represents the fields that can be updated for a photo.
// All fields are optional (pointers) to support partial updates.
type PhotoUpdate struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	AlbumID     *string `json:"album_id"`
}

// PhotoResponse represents the JSON response format for photo data.
// It includes computed URLs for thumbnail and full-size images.
type PhotoResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	AlbumID      *string `json:"album_id"`
	AlbumTitle   *string `json:"album_title"`
	ThumbnailURL string  `json:"thumbnail_url"`
	FullURL      string  `json:"full_url"`
	Type         string  `json:"type"`
}

// NeedsMetadata determines if a photo requires metadata updates.
// Returns true if the photo has a generic/empty title or empty description.
func (p *Photo) NeedsMetadata() bool {
	return p.hasGenericTitle() || p.hasEmptyDescription()
}

// hasGenericTitle checks if the photo has a generic camera-generated title.
func (p *Photo) hasGenericTitle() bool {
	if p.Title == "" {
		return true
	}

	return IsGenericTitle(p.Title)
}

// hasEmptyDescription checks if the photo has an empty or nil description.
func (p *Photo) hasEmptyDescription() bool {
	return p.Description == nil || *p.Description == ""
}

// ToPhotoResponse converts a PhotoWithSizeVariants to a PhotoResponse with proper URL generation
func (p *PhotoWithSizeVariants) ToPhotoResponse(lycheeBaseURL string) PhotoResponse {
	thumbnailURL := ""
	fullURL := ""

	// Construct thumbnail URL
	if p.ThumbnailPath != nil && *p.ThumbnailPath != "" {
		thumbnailURL = constructImageURL(lycheeBaseURL, *p.ThumbnailPath)
	}

	// Construct full/original image URL
	if p.OriginalPath != nil && *p.OriginalPath != "" {
		fullURL = constructImageURL(lycheeBaseURL, *p.OriginalPath)
	}

	return PhotoResponse{
		ID:           p.ID,
		Title:        p.Title,
		Description:  p.Description,
		AlbumID:      p.AlbumID,
		AlbumTitle:   p.AlbumTitle,
		ThumbnailURL: thumbnailURL,
		FullURL:      fullURL,
		Type:         p.Type,
	}
}

// constructImageURL builds a proper URL from the Lychee base URL and image path
func constructImageURL(baseURL, imagePath string) string {
	if baseURL == "" || imagePath == "" {
		return ""
	}

	// Ensure base URL doesn't end with slash and image path doesn't start with slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	imagePath = strings.TrimPrefix(imagePath, "/")

	return baseURL + "/uploads/" + imagePath
}
