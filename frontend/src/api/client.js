import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor for logging
api.interceptors.request.use(
  (config) => {
    console.log(`API Request: ${config.method?.toUpperCase()} ${config.url}`)
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    console.error('API Error:', error.response?.data || error.message)
    return Promise.reject(error)
  }
)

export const photosAPI = {
  // Get photos that need metadata
  getPhotosNeedingMetadata(params = {}) {
    return api.get('/photos/needsmetadata', { params })
  },

  // Get a specific photo by ID
  getPhotoById(id) {
    return api.get(`/photos/${id}`)
  },

  // Update photo metadata
  updatePhoto(id, data) {
    return api.put(`/photos/${id}`, data)
  }
}

export const albumsAPI = {
  // Get all albums
  getAlbums() {
    return api.get('/albums')
  },

  // Get albums that have photos needing metadata
  getAlbumsWithPhotoCounts() {
    return api.get('/albums/withphotocounts')
  }
}

export const healthAPI = {
  // Health check
  check() {
    return api.get('/health')
  }
}

export default api