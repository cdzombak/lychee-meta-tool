package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cdzombak/lychee-meta-tool/backend/db"
	"github.com/cdzombak/lychee-meta-tool/backend/models"
)

type AlbumHandler struct {
	db *db.DB
}

func NewAlbumHandler(database *db.DB) *AlbumHandler {
	return &AlbumHandler{db: database}
}

func (h *AlbumHandler) GetAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	albums, err := h.db.GetAlbums()
	if err != nil {
		log.Printf("Failed to get albums: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AlbumHandler) GetAlbumsWithPhotoCounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	albums, err := h.db.GetAlbumsWithPhotoCounts()
	if err != nil {
		log.Printf("Failed to get albums with photo counts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}