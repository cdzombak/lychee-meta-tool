package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/cdzombak/lychee-meta-tool/backend/ai"
	"github.com/cdzombak/lychee-meta-tool/backend/constants"
	"github.com/cdzombak/lychee-meta-tool/backend/db"
	"github.com/cdzombak/lychee-meta-tool/backend/models"
)

// PhotoHandler handles HTTP requests related to photos
type PhotoHandler struct {
	db            *db.DB
	lycheeBaseURL string
	aiClient      ai.Client
}

// NewPhotoHandler creates a new PhotoHandler with the provided dependencies
func NewPhotoHandler(database *db.DB, lycheeBaseURL string, aiClient ai.Client) *PhotoHandler {
	return &PhotoHandler{
		db:            database,
		lycheeBaseURL: lycheeBaseURL,
		aiClient:      aiClient,
	}
}

// PhotosNeedingMetadataResponse represents the response for photos needing metadata
type PhotosNeedingMetadataResponse struct {
	Photos []models.PhotoResponse `json:"photos"`
	Total  int                    `json:"total"`
}

// GetPhotosNeedingMetadata handles GET requests to retrieve photos that need metadata
func (h *PhotoHandler) GetPhotosNeedingMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		MethodNotAllowed(w)
		return
	}

	// Parse and validate query parameters
	query := r.URL.Query()
	var albumID *string
	if aid := sanitizeQueryParam(query.Get("album_id")); aid != "" {
		if !validateAlbumID(aid) {
			BadRequest(w, "Invalid album_id format. Must be alphanumeric with underscores and hyphens only.", nil)
			return
		}
		albumID = &aid
	}

	limit := DefaultLimit
	if l := sanitizeQueryParam(query.Get("limit")); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = validateLimit(parsed)
		} else {
			BadRequest(w, fmt.Sprintf("Invalid limit parameter. Must be a number between 1 and %d.", MaxLimit), nil)
			return
		}
	}

	offset := 0
	if o := sanitizeQueryParam(query.Get("offset")); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = validateOffset(parsed)
		} else {
			BadRequest(w, "Invalid offset parameter. Must be a non-negative number.", nil)
			return
		}
	}

	photos, err := h.db.GetPhotosNeedingMetadata(albumID, limit, offset)
	if err != nil {
		log.Printf("Failed to get photos needing metadata (album_id=%v, limit=%d, offset=%d): %v", albumID, limit, offset, err)
		InternalServerError(w, "Failed to retrieve photos. Please try again.")
		return
	}

	// Convert to response format
	photoResponses := make([]models.PhotoResponse, len(photos))
	for i, photo := range photos {
		photoResponses[i] = photo.ToPhotoResponse(h.lycheeBaseURL)
	}

	response := PhotosNeedingMetadataResponse{
		Photos: photoResponses,
		Total:  len(photoResponses),
	}

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(response)
}

// GetPhotoByID handles GET requests to retrieve a specific photo by ID
func (h *PhotoHandler) GetPhotoByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		MethodNotAllowed(w)
		return
	}

	// Extract and validate photo ID from URL path
	photoID, valid := extractPhotoIDFromPath(r.URL.Path)
	if !valid {
		InvalidID(w, "photo ID")
		return
	}

	photo, err := h.db.GetPhotoByID(photoID)
	if err != nil {
		DatabaseError(w, fmt.Sprintf("get photo by ID %s", photoID), err)
		return
	}

	if photo == nil {
		NotFound(w, fmt.Sprintf("Photo with ID '%s' not found", photoID))
		return
	}

	response := photo.ToPhotoResponse(h.lycheeBaseURL)

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *PhotoHandler) UpdatePhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract and validate photo ID from URL path
	photoID, valid := extractPhotoIDFromPath(r.URL.Path)
	if !valid {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Invalid photo ID format. Must be 1-64 characters, alphanumeric with underscores and hyphens only.",
		})
		return
	}

	// Parse and validate JSON input
	var update models.PhotoUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: fmt.Sprintf("Invalid JSON format: %v", err),
		})
		return
	}

	// Validate and sanitize the update data
	if validationErrors := ValidatePhotoUpdate(&update); len(validationErrors) > 0 {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		errorMessages := make([]string, len(validationErrors))
		for i, err := range validationErrors {
			errorMessages[i] = err.Error()
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Validation failed",
			"details": errorMessages,
		})
		return
	}

	// Update the photo
	if err := h.db.UpdatePhoto(photoID, update); err != nil {
		log.Printf("Failed to update photo %s: %v", photoID, err)
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Failed to update photo. Please try again.",
		})
		return
	}

	// Get updated photo
	photo, err := h.db.GetPhotoByID(photoID)
	if err != nil {
		log.Printf("Failed to get updated photo %s: %v", photoID, err)
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Photo updated successfully but failed to retrieve updated data.",
		})
		return
	}

	response := struct {
		Success bool                 `json:"success"`
		Photo   models.PhotoResponse `json:"photo"`
	}{
		Success: true,
		Photo:   photo.ToPhotoResponse(h.lycheeBaseURL),
	}

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *PhotoHandler) GenerateAITitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.aiClient == nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "AI title generation is not configured. Please check your AI backend configuration.",
		})
		return
	}

	// Extract and validate photo ID from URL path
	photoID, valid := extractPhotoIDFromPath(r.URL.Path)
	if !valid {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Invalid photo ID format. Must be 1-64 characters, alphanumeric with underscores and hyphens only.",
		})
		return
	}

	// Get photo details
	photo, err := h.db.GetPhotoByID(photoID)
	if err != nil {
		log.Printf("Failed to get photo by ID %s for AI title generation: %v", photoID, err)
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Failed to retrieve photo details. Please try again.",
		})
		return
	}

	if photo == nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: fmt.Sprintf("Photo with ID '%s' not found.", photoID),
		})
		return
	}

	// Construct photo URL
	photoResponse := photo.ToPhotoResponse(h.lycheeBaseURL)
	imageURL := photoResponse.FullURL

	// Validate image URL
	if imageURL == "" {
		log.Printf("Empty image URL for photo %s", photoID)
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Photo image URL is not available.",
		})
		return
	}

	// Generate title with timeout
	ctx, cancel := context.WithTimeout(context.Background(), constants.AIGenerationTimeout)
	defer cancel()

	log.Printf("Generating AI title for photo %s using image URL: %s", photoID, imageURL)
	title, err := h.aiClient.GenerateTitle(ctx, imageURL)
	if err != nil {
		log.Printf("Failed to generate AI title for photo %s: %v", photoID, err)
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Failed to generate AI title. Please check your network connection and try again.",
		})
		return
	}

	// Sanitize and validate the generated title
	title = sanitizeText(strings.Trim(strings.TrimSpace(title), `"'`))
	if title == "" {
		log.Printf("AI generated empty title for photo %s", photoID)
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: "AI generated an empty title. Please try again.",
		})
		return
	}

	// Validate the generated title length
	if len(title) > MaxTitleLength {
		log.Printf("AI generated title too long for photo %s: %d characters", photoID, len(title))
		// Truncate to max length
		title = title[:MaxTitleLength]
	}

	log.Printf("Successfully generated AI title for photo %s: %s", photoID, title)

	response := struct {
		Success bool   `json:"success"`
		Title   string `json:"title"`
	}{
		Success: true,
		Title:   title,
	}

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(response)
}
