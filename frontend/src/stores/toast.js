import { defineStore } from 'pinia'

export const useToastStore = defineStore('toast', {
  state: () => ({
    currentToast: null,
    timeoutId: null
  }),

  actions: {
    showToast(message, type = 'info', duration = 3000) {
      // Clear existing timeout
      if (this.timeoutId) {
        clearTimeout(this.timeoutId)
      }

      this.currentToast = {
        message,
        type
      }

      // Auto-hide after duration
      this.timeoutId = setTimeout(() => {
        this.currentToast = null
        this.timeoutId = null
      }, duration)
    },

    showSuccess(message, duration = 3000) {
      this.showToast(message, 'success', duration)
    },

    showError(message, duration = 5000) {
      this.showToast(message, 'error', duration)
    },

    showInfo(message, duration = 3000) {
      this.showToast(message, 'info', duration)
    },

    hide() {
      if (this.timeoutId) {
        clearTimeout(this.timeoutId)
        this.timeoutId = null
      }
      this.currentToast = null
    }
  }
})