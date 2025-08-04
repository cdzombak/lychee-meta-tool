<template>
  <div v-if="toast" :class="['toast', toast.type]">
    {{ toast.message }}
  </div>
</template>

<script>
import { computed } from 'vue'
import { useToastStore } from '../stores/toast'

export default {
  name: 'Toast',
  setup() {
    const toastStore = useToastStore()
    
    const toast = computed(() => toastStore.currentToast)
    
    return {
      toast
    }
  }
}
</script>