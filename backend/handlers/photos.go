package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/cdzombak/lychee-meta-tool/backend/db"
	"github.com/cdzombak/lychee-meta-tool/backend/models"
)

type PhotoHandler struct {
	db            *db.DB
	lycheeBaseURL string
}

func NewPhotoHandler(database *db.DB, lycheeBaseURL string) *PhotoHandler {
	return &PhotoHandler{
		db:            database,
		lycheeBaseURL: lycheeBaseURL,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type PhotosNeedingMetadataResponse struct {
	Photos []models.PhotoResponse `json:"photos"`
	Total  int                    `json:"total"`
}

func (h *PhotoHandler) GetPhotosNeedingMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	var albumID *string
	if aid := query.Get("album_id"); aid != "" {
		albumID = &aid
	}

	limit := DefaultLimit
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = validateLimit(parsed)
		}
	}

	offset := 0
	if o := query.Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = validateOffset(parsed)
		}
	}

	photos, err := h.db.GetPhotosNeedingMetadata(albumID, limit, offset)
	if err != nil {
		log.Printf("Failed to get photos needing metadata: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PhotoHandler) GetPhotoByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract and validate photo ID from URL path
	photoID, valid := extractPhotoIDFromPath(r.URL.Path)
	if !valid {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	photo, err := h.db.GetPhotoByID(photoID)
	if err != nil {
		log.Printf("Failed to get photo by ID %s: %v", photoID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if photo == nil {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	response := photo.ToPhotoResponse(h.lycheeBaseURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PhotoHandler) UpdatePhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract and validate photo ID from URL path
	photoID, valid := extractPhotoIDFromPath(r.URL.Path)
	if !valid {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	var update models.PhotoUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.db.UpdatePhoto(photoID, update); err != nil {
		log.Printf("Failed to update photo %s: %v", photoID, err)
		http.Error(w, "Failed to update photo", http.StatusInternalServerError)
		return
	}

	// Get updated photo
	photo, err := h.db.GetPhotoByID(photoID)
	if err != nil {
		log.Printf("Failed to get updated photo %s: %v", photoID, err)
		http.Error(w, "Failed to retrieve updated photo", http.StatusInternalServerError)
		return
	}

	response := struct {
		Success bool                   `json:"success"`
		Photo   models.PhotoResponse   `json:"photo"`
	}{
		Success: true,
		Photo: photo.ToPhotoResponse(h.lycheeBaseURL),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}