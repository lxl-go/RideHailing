import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useLocationStore = defineStore('location', () => {
  const longitude = ref(null)
  const latitude = ref(null)
  const address = ref('')
  const ready = ref(false)

  function setLocation({ longitude: lng, latitude: lat, address: addr } = {}) {
    if (typeof lng === 'number') longitude.value = lng
    if (typeof lat === 'number') latitude.value = lat
    if (addr) address.value = addr
    ready.value = true
  }

  function get() {
    return { longitude: longitude.value, latitude: latitude.value, address: address.value }
  }

  return { longitude, latitude, address, ready, setLocation, get }
})
