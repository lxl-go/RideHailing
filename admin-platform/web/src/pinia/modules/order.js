import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useOrderAdminStore = defineStore('orderAdmin', () => {
  const currentOrder = ref(null)
  const orderList = ref([])
  const totalOrders = ref(0)

  function setCurrentOrder(order) {
    currentOrder.value = order
  }

  function setOrderList(list, total) {
    orderList.value = list
    if (total !== undefined) totalOrders.value = total
  }

  function clearOrder() {
    currentOrder.value = null
  }

  return { currentOrder, orderList, totalOrders, setCurrentOrder, setOrderList, clearOrder }
})
