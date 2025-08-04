<template>
  <div class="filmstrip">
    <div v-if="photosStore.loading" class="loading">
      <div class="spinner"></div>
      Loading photos...
    </div>
    
    <div v-else-if="!photosStore.hasPhotos" class="loading">
      No photos need metadata updates
    </div>
    
    <template v-else>
      <img
        v-for="(photo, index) in photosStore.photos"
        :key="photo.id"
        :src="photo.thumbnail_url"
        :alt="photo.title"
        class="photo-thumb"
        :class="{ selected: index === photosStore.currentPhotoIndex }"
        @click="photosStore.selectPhoto(index)"
        @error="handleImageError"
      />
    </template>
  </div>
</template>

<script>
import { usePhotosStore } from '../stores/photos'

export default {
  name: 'FilmStrip',
  setup() {
    const photosStore = usePhotosStore()

    const handleImageError = (event) => {
      // Replace broken image with placeholder
      event.target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iODAiIGhlaWdodD0iODAiIHZpZXdCb3g9IjAgMCA4MCA4MCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHJlY3Qgd2lkdGg9IjgwIiBoZWlnaHQ9IjgwIiBmaWxsPSIjRjBGMEYwIi8+CjxwYXRoIGQ9Ik0yNSAzNUMzMC41MjI4IDM1IDM1IDMwLjUyMjggMzUgMjVDMzUgMTkuNDc3MiAzMC41MjI4IDE1IDI1IDE1QzE5LjQ3NzIgMTUgMTUgMTkuNDc3MiAxNSAyNUMxNSAzMC41MjI4IDE5LjQ3NzIgMzUgMjUgMzVaIiBmaWxsPSIjQzRDNEM0Ii8+CjxwYXRoIGQ9Ik0xMCA1NUw2NSA1NVY2NUgxMFY1NVoiIGZpbGw9IiNDNEM0QzQiLz4KPC9zdmc+'
    }


    return {
      photosStore,
      handleImageError
    }
  }
}
</script>