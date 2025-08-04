<template>
  <div class="album-selector">
    <input
      v-model="searchTerm"
      type="text"
      class="album-search"
      :placeholder="currentAlbumTitle || 'Search albums...'"
      @focus="showDropdown = true"
      @blur="hideDropdown"
      @keydown.enter="selectFirstMatch"
      @keydown.escape="hideDropdown"
      @keydown.arrow-down="navigateDown"
      @keydown.arrow-up="navigateUp"
    />
    
    <div v-if="showDropdown && filteredAlbums.length > 0" class="album-dropdown">
      <div
        v-for="(album, index) in filteredAlbums"
        :key="album.id"
        class="album-option"
        :class="{
          selected: index === selectedIndex,
          current: album.id === modelValue
        }"
        @mousedown="selectAlbum(album)"
        @mouseenter="selectedIndex = index"
      >
        {{ album.title }}
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, watch } from 'vue'
import { usePhotosStore } from '../stores/photos'

export default {
  name: 'AlbumSelector',
  props: {
    modelValue: {
      type: String,
      default: null
    },
    currentAlbumTitle: {
      type: String,
      default: null
    }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const photosStore = usePhotosStore()
    
    const searchTerm = ref('')
    const showDropdown = ref(false)
    const selectedIndex = ref(0)
    
    const filteredAlbums = computed(() => {
      if (!searchTerm.value) {
        return photosStore.albums
      }
      
      return photosStore.albums.filter(album =>
        album.title.toLowerCase().includes(searchTerm.value.toLowerCase())
      )
    })
    
    // Reset selected index when filtered albums change
    watch(filteredAlbums, () => {
      selectedIndex.value = 0
    })
    
    const selectAlbum = (album) => {
      emit('update:modelValue', album.id)
      searchTerm.value = ''
      showDropdown.value = false
    }
    
    const selectFirstMatch = () => {
      if (filteredAlbums.value.length > 0) {
        selectAlbum(filteredAlbums.value[selectedIndex.value])
      }
    }
    
    const hideDropdown = () => {
      // Delay hiding to allow click events to register
      setTimeout(() => {
        showDropdown.value = false
        searchTerm.value = ''
      }, 150)
    }
    
    const navigateDown = (event) => {
      event.preventDefault()
      if (selectedIndex.value < filteredAlbums.value.length - 1) {
        selectedIndex.value++
      }
    }
    
    const navigateUp = (event) => {
      event.preventDefault()
      if (selectedIndex.value > 0) {
        selectedIndex.value--
      }
    }
    
    return {
      searchTerm,
      showDropdown,
      selectedIndex,
      filteredAlbums,
      selectAlbum,
      selectFirstMatch,
      hideDropdown,
      navigateDown,
      navigateUp
    }
  }
}
</script>

<style scoped>
.album-selector {
  position: relative;
  min-width: 200px;
}

.album-search {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ced4da;
  border-radius: 4px;
  font-size: 14px;
  outline: none;
}

.album-search:focus {
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
}

.album-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: white;
  border: 1px solid #ced4da;
  border-top: none;
  border-radius: 0 0 4px 4px;
  max-height: 200px;
  overflow-y: auto;
  z-index: 1000;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.album-option {
  padding: 8px 12px;
  cursor: pointer;
  border-bottom: 1px solid #f8f9fa;
}

.album-option:hover,
.album-option.selected {
  background-color: #f8f9fa;
}

.album-option.current {
  background-color: #e3f2fd;
  font-weight: 500;
}

.album-option:last-child {
  border-bottom: none;
}
</style>