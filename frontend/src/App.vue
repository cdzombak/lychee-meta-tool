<template>
  <div id="app">
    <!-- Filmstrip -->
    <FilmStrip />
    
    <!-- Main content area -->
    <div class="main-content">
      <!-- Photo viewer -->
      <PhotoViewer />
      
      <!-- Right column with filter and editor -->
      <div class="right-column">
        <!-- Album filter -->
        <div class="filter-section">
          <label for="album-filter">Filter by Album:</label>
          <AlbumSelector
            v-model="selectedAlbumId"
            :current-album-title="currentAlbumTitle"
            @update:model-value="handleAlbumChange"
          />
          <button 
            v-if="selectedAlbumId" 
            class="clear-filter-btn"
            @click="clearAlbumFilter"
          >
            Clear Filter
          </button>
        </div>
        
        <!-- Photo editor -->
        <PhotoEditor />
      </div>
    </div>
    
    <!-- Toast notifications -->
    <Toast />
  </div>
</template>

<script>
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { usePhotosStore } from './stores/photos'
import FilmStrip from './components/FilmStrip.vue'
import PhotoViewer from './components/PhotoViewer.vue'
import PhotoEditor from './components/PhotoEditor.vue'
import AlbumSelector from './components/AlbumSelector.vue'
import Toast from './components/Toast.vue'

export default {
  name: 'App',
  components: {
    FilmStrip,
    PhotoViewer,
    PhotoEditor,
    AlbumSelector,
    Toast
  },
  setup() {
    const photosStore = usePhotosStore()
    
    // Album filtering
    const selectedAlbumId = ref(null)
    
    const currentAlbumTitle = computed(() => {
      if (!selectedAlbumId.value) return null
      const album = photosStore.albumById(selectedAlbumId.value)
      return album ? album.title : null
    })
    
    const handleAlbumChange = (albumId) => {
      selectedAlbumId.value = albumId
      photosStore.setAlbumFilter(albumId)
    }
    
    const clearAlbumFilter = () => {
      selectedAlbumId.value = null
      photosStore.clearFilter()
    }

    // Keyboard shortcuts
    const handleKeydown = (event) => {
      // Check if user is typing in an input field
      if (event.target.tagName === 'INPUT' || event.target.tagName === 'TEXTAREA') {
        return
      }

      if ((event.metaKey || event.ctrlKey) && event.key === 'j') {
        event.preventDefault()
        photosStore.previousPhoto()
      } else if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
        event.preventDefault()
        photosStore.nextPhoto()
      }
    }

    onMounted(async () => {
      try {
        // Load initial data
        await Promise.all([
          photosStore.loadPhotos(),
          photosStore.loadAlbums()
        ])
      } catch (error) {
        console.error('Failed to load initial data:', error)
      }

      // Add keyboard event listener
      document.addEventListener('keydown', handleKeydown)
    })

    onUnmounted(() => {
      document.removeEventListener('keydown', handleKeydown)
    })

    return {
      selectedAlbumId,
      currentAlbumTitle,
      handleAlbumChange,
      clearAlbumFilter
    }
  }
}
</script>

<style scoped>
.right-column {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.filter-section {
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.filter-section label {
  font-weight: 500;
  color: #495057;
  font-size: 14px;
  margin-bottom: 4px;
}

.clear-filter-btn {
  background: #dc3545;
  color: white;
  border: none;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.2s;
  align-self: flex-start;
  margin-top: 8px;
}

.clear-filter-btn:hover {
  background: #c82333;
}
</style>