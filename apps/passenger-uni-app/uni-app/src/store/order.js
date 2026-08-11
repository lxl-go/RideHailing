import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listOrders } from '@/api/order'

export const useOrderStore = defineStore('order', () => {
  const list = ref([])
  const total = ref(0)
  const loading = ref(false)
  const currentOrder = ref(null)

  async function loadList(params = { page: 1, page_size: 20 }) {
    loading.value = true
    const res = await listOrders(params)
    loading.value = false
    if (res.code === 0) {
      list.value = res.data?.items || res.data?.list || res.data || []
      total.value = res.data?.total || list.value.length
    } else {
      list.value = []
    }
    return res
  }

  function setCurrent(order) {
    currentOrder.value = order
  }

  function reset() {
    list.value = []
    total.value = 0
    currentOrder.value = null
  }

  return { list, total, loading, currentOrder, loadList, setCurrent, reset }
})
