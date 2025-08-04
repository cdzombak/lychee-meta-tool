<template>
  <div class="photo-editor">
    <div v-if="!photosStore.hasPhotos" class="loading">
      No photo selected
    </div>
    
    <template v-else-if="currentPhoto">
      <h3>Edit Photo</h3>
      
      <div class="form-group">
        <label for="title">Title</label>
        <input
          id="title"
          ref="titleInput"
          v-model="formData.title"
          type="text"
          placeholder="Enter photo title..."
          @keydown.enter="saveTitle"
          @keydown.tab="focusDescription"
        />
      </div>
      
      <div class="form-group">
        <label for="description">Description</label>
        <textarea
          id="description"
          ref="descriptionInput"
          v-model="formData.description"
          placeholder="Enter photo description..."
          @keydown.enter="saveDescription"
        ></textarea>
      </div>
      
      <div class="form-group">
        <label for="album">Album</label>
        <AlbumSelector
          v-model="formData.albumId"
          :current-album-title="currentPhoto.album_title"
        />
      </div>
      
      <div class="form-group">
        <button
          @click="saveChanges"
          :disabled="saving"
          class="save-button"
        >
          {{ saving ? 'Saving...' : 'Save Changes' }}
        </button>
      </div>
    </template>
  </div>
</template>

<script>
import { ref, computed, watch, nextTick } from 'vue'
import { usePhotosStore } from '../stores/photos'
import { useToastStore } from '../stores/toast'
import AlbumSelector from './AlbumSelector.vue'

export default {
  name: 'PhotoEditor',
  components: {
    AlbumSelector
  },
  setup() {
    const photosStore = usePhotosStore()
    const toastStore = useToastStore()
    
    const titleInput = ref(null)
    const descriptionInput = ref(null)
    const saving = ref(false)
    
    const formData = ref({
      title: '',
      description: '',
      albumId: null
    })
    
    const currentPhoto = computed(() => photosStore.currentPhoto)
    
    // Watch for photo changes and update form data
    watch(currentPhoto, (newPhoto) => {
      if (newPhoto) {
        formData.value = {
          title: newPhoto.title || '',
          description: newPhoto.description || '',
          albumId: newPhoto.album_id
        }
        
        // Focus and select title input
        nextTick(() => {
          if (titleInput.value) {
            titleInput.value.focus()
            titleInput.value.select()
          }
        })
      }
    }, { immediate: true })
    
    const saveTitle = () => {
      saveChanges()
    }
    
    const saveDescription = () => {
      saveChanges()
    }
    
    const focusDescription = (event) => {
      event.preventDefault()
      if (descriptionInput.value) {
        descriptionInput.value.focus()
      }
    }
    
    const saveChanges = async () => {
      if (!currentPhoto.value || saving.value) return
      
      saving.value = true
      
      try {
        const updateData = {}
        
        // Only include changed fields
        if (formData.value.title !== currentPhoto.value.title) {
          updateData.title = formData.value.title
        }
        
        if (formData.value.description !== (currentPhoto.value.description || '')) {
          updateData.description = formData.value.description
        }
        
        if (formData.value.albumId !== currentPhoto.value.album_id) {
          updateData.album_id = formData.value.albumId
        }
        
        // Only save if there are changes
        if (Object.keys(updateData).length > 0) {
          await photosStore.updatePhoto(currentPhoto.value.id, updateData)
          toastStore.showSuccess('Photo updated successfully!')
        } else {
          toastStore.showInfo('No changes to save')
        }
      } catch (error) {
        toastStore.showError(error.message || 'Failed to update photo')
      } finally {
        saving.value = false
      }
    }
    
    return {
      photosStore,
      titleInput,
      descriptionInput,
      saving,
      formData,
      currentPhoto,
      saveTitle,
      saveDescription,
      focusDescription,
      saveChanges
    }
  }
}
</script>

<style scoped>
.save-button {
  background-color: #007bff;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.save-button:hover:not(:disabled) {
  background-color: #0056b3;
}

.save-button:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
}

h3 {
  margin-bottom: 15px;
  color: #333;
}
</style>