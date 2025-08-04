import { defineStore } from 'pinia'
import { photosAPI, albumsAPI } from '../api/client'

export const usePhotosStore = defineStore('photos', {
  state: () => ({
    photos: [],
    albums: [],
    currentPhotoIndex: 0,
    loading: false,
    error: null,
    filter: {
      albumId: null
    }
  }),

  getters: {
    currentPhoto: (state) => {
      return state.photos[state.currentPhotoIndex] || null
    },

    hasPhotos: (state) => {
      return state.photos.length > 0
    },

    hasPrevious: (state) => {
      return state.currentPhotoIndex > 0
    },

    hasNext: (state) => {
      return state.currentPhotoIndex < state.photos.length - 1
    },

    albumById: (state) => {
      return (id) => state.albums.find(album => album.id === id)
    }
  },

  actions: {
    async loadPhotos() {
      this.loading = true
      this.error = null
      
      try {
        const params = {}
        if (this.filter.albumId) {
          params.album_id = this.filter.albumId
        }

        const response = await photosAPI.getPhotosNeedingMetadata(params)
        this.photos = response.data.photos || []
        
        // Reset current photo index if no photos or out of bounds
        if (this.photos.length === 0) {
          this.currentPhotoIndex = 0
        } else if (this.currentPhotoIndex >= this.photos.length) {
          this.currentPhotoIndex = 0
        }
      } catch (error) {
        this.error = error.response?.data?.error || 'Failed to load photos'
        console.error('Failed to load photos:', error)
      } finally {
        this.loading = false
      }
    },

    async loadAlbums() {
      try {
        const response = await albumsAPI.getAlbumsWithPhotoCounts()
        this.albums = response.data.albums || []
      } catch (error) {
        console.error('Failed to load albums:', error)
      }
    },

    async updatePhoto(id, data) {
      try {
        const response = await photosAPI.updatePhoto(id, data)
        
        // Remove the updated photo from the list since it no longer needs metadata
        const photoIndex = this.photos.findIndex(photo => photo.id === id)
        if (photoIndex !== -1) {
          this.photos.splice(photoIndex, 1)
          
          // Adjust current photo index
          if (this.currentPhotoIndex >= this.photos.length) {
            this.currentPhotoIndex = Math.max(0, this.photos.length - 1)
          }
        }
        
        return response.data
      } catch (error) {
        const errorMessage = error.response?.data?.error || 'Failed to update photo'
        throw new Error(errorMessage)
      }
    },

    selectPhoto(index) {
      if (index >= 0 && index < this.photos.length) {
        this.currentPhotoIndex = index
      }
    },

    selectPhotoById(id) {
      const index = this.photos.findIndex(photo => photo.id === id)
      if (index !== -1) {
        this.currentPhotoIndex = index
      }
    },

    nextPhoto() {
      if (this.hasNext) {
        this.currentPhotoIndex++
      }
    },

    previousPhoto() {
      if (this.hasPrevious) {
        this.currentPhotoIndex--
      }
    },

    setAlbumFilter(albumId) {
      this.filter.albumId = albumId
      this.loadPhotos()
    },

    clearFilter() {
      this.filter.albumId = null
      this.loadPhotos()
    }
  }
})