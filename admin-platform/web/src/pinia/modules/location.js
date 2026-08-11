import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useLocationAdminStore = defineStore('locationAdmin', () => {
  const location = ref({
    lat: 31.2304,
    lng: 121.4737,
    address: '上海',
    city: '上海市',
    district: '浦东新区',
  })

  function setLocation(pos) {
    location.value = pos
  }

  return { location, setLocation }
})
