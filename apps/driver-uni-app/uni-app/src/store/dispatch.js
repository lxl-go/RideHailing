import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listAvailableOrders } from '@/api/order'

export const useDispatchStore = defineStore('dispatch', () => {
  const pendingOrders = ref([])
  const loading = ref(false)
  const lastLoadAt = ref(0)

  async function loadPending(params = {}) {
    loading.value = true
    const res = await listAvailableOrders(params)
    loading.value = false
    lastLoadAt.value = Date.now()
    if (res?.code === 0) {
      pendingOrders.value = res.data?.list || res.data || []
    } else {
      pendingOrders.value = []
    }
    return res
  }

  function removeOrder(id) {
    pendingOrders.value = pendingOrders.value.filter((o) => o.id !== id)
  }

  function reset() {
    pendingOrders.value = []
  }

  return { pendingOrders, loading, lastLoadAt, loadPending, removeOrder, reset }
})
