package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cdzombak/lychee-meta-tool/backend/constants"
	"github.com/cdzombak/lychee-meta-tool/backend/db"
	"github.com/cdzombak/lychee-meta-tool/backend/models"
)

// AlbumHandler handles HTTP requests related to photo albums
type AlbumHandler struct {
	db *db.DB
}

// NewAlbumHandler creates a new AlbumHandler with the provided database connection
func NewAlbumHandler(database *db.DB) *AlbumHandler {
	return &AlbumHandler{db: database}
}

// GetAlbums handles GET requests to retrieve all albums
func (h *AlbumHandler) GetAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		MethodNotAllowed(w)
		return
	}

	albums, err := h.db.GetAlbums()
	if err != nil {
		DatabaseError(w, "get albums", err)
		return
	}

	// Convert to response format
	albumResponses := make([]models.AlbumResponse, len(albums))
	for i, album := range albums {
		albumResponses[i] = models.AlbumResponse{
			ID:    album.ID,
			Title: album.Title,
		}
	}

	response := models.AlbumsResponse{
		Albums: albumResponses,
	}

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode albums response: %v", err)
	}
}

// GetAlbumsWithPhotoCounts handles GET requests to retrieve albums containing photos that need metadata
func (h *AlbumHandler) GetAlbumsWithPhotoCounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		MethodNotAllowed(w)
		return
	}

	albums, err := h.db.GetAlbumsWithPhotoCounts()
	if err != nil {
		DatabaseError(w, "get albums with photo counts", err)
		return
	}

	// Convert to response format - only include albums with photos needing metadata
	albumResponses := make([]models.AlbumResponse, len(albums))
	for i, album := range albums {
		albumResponses[i] = models.AlbumResponse{
			ID:    album.ID,
			Title: album.Title,
		}
	}

	response := models.AlbumsResponse{
		Albums: albumResponses,
	}

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode albums with photo counts response: %v", err)
	}
}