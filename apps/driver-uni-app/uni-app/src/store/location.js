import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useLocationStore = defineStore('location', () => {
  const longitude = ref(null)
  const latitude = ref(null)
  const address = ref('')
  const reporting = ref(false)

  function setLocation(payload = {}) {
    if (typeof payload.longitude === 'number') longitude.value = payload.longitude
    if (typeof payload.latitude === 'number') latitude.value = payload.latitude
    if (payload.address) address.value = payload.address
  }

  return { longitude, latitude, address, reporting, setLocation }
})
