<template>
  <div class="photo-viewer">
    <div v-if="!photosStore.hasPhotos" class="loading">
      Select a photo to view
    </div>
    
    <template v-else-if="currentPhoto">
      <img
        :src="currentPhoto.full_url"
        :alt="currentPhoto.title"
        @error="handleImageError"
      />
      
      <!-- Navigation buttons -->
      <button
        v-if="photosStore.hasPrevious"
        class="nav-buttons nav-prev"
        @click="photosStore.previousPhoto()"
        title="Previous photo (Cmd+J)"
      >
        ←
      </button>
      
      <button
        v-if="photosStore.hasNext"
        class="nav-buttons nav-next"
        @click="photosStore.nextPhoto()"
        title="Next photo (Cmd+K)"
      >
        →
      </button>
    </template>
  </div>
</template>

<script>
import { computed } from 'vue'
import { usePhotosStore } from '../stores/photos'

export default {
  name: 'PhotoViewer',
  setup() {
    const photosStore = usePhotosStore()
    
    const currentPhoto = computed(() => photosStore.currentPhoto)

    const handleImageError = (event) => {
      // Replace broken image with placeholder
      event.target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAwIiBoZWlnaHQ9IjMwMCIgdmlld0JveD0iMCAwIDQwMCAzMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxyZWN0IHdpZHRoPSI0MDAiIGhlaWdodD0iMzAwIiBmaWxsPSIjRjBGMEYwIi8+CjxwYXRoIGQ9Ik0xNTAgMTIwQzE3Mi4wOTEgMTIwIDE5MCA5Ny45MDg2IDE5MCA3NUMxOTAgNTIuMDkxNCAxNzIuMDkxIDMwIDE1MCAzMEMxMjcuOTA5IDMwIDExMCA1Mi4wOTE0IDExMCA3NUMxMTAgOTcuOTA4NiAxMjcuOTA5IDEyMCAxNTAgMTIwWiIgZmlsbD0iI0M0QzRDNCIvPgo8cGF0aCBkPSJNNzAgMjEwTDMzMCAyMTBWMjcwSDcwVjIxMFoiIGZpbGw9IiNDNEM0QzQiLz4KPHRLEHN0eWxlPSJmb250LWZhbWlseTogQXJpYWwsIHNhbnMtc2VyaWY7IGZvbnQtc2l6ZTogMTRweDsgZmlsbDogIzk5OTk5OTsiIHg9IjIwMCIgeT0iMTYwIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIj5JbWFnZSBub3QgZm91bmQ8L3RleHQ+Cjwvc3ZnPg=='
    }

    return {
      photosStore,
      currentPhoto,
      handleImageError
    }
  }
}
</script>