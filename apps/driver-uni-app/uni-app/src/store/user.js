import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getDriverInfo } from '@/api/driver'

export const useUserStore = defineStore('user', () => {
  const userId = ref('')
  const accessToken = ref('')
  const refreshToken = ref('')
  const online = ref(false)
  const driverInfo = ref(null)

  const isLoggedIn = computed(() => Boolean(accessToken.value))

  function setUserId(id) {
    userId.value = id
    uni.setStorageSync('driverUserId', id)
  }

  function setSession(session) {
    const nextUserId = String(session?.user_id || session?.userId || session?.userID || userId.value || '')
    const token = session?.access_token || session?.accessToken || ''
    const nextRefreshToken = session?.refresh_token || session?.refreshToken || ''
    setUserId(nextUserId)
    accessToken.value = token
    refreshToken.value = nextRefreshToken
    if (token) uni.setStorageSync('driverAccessToken', token)
    if (nextRefreshToken) uni.setStorageSync('driverRefreshToken', nextRefreshToken)
  }

  async function loadDriverInfo() {
    const res = await getDriverInfo()
    if (res.code === 0) {
      driverInfo.value = res.data
      if (res.data?.user_id || res.data?.id) setUserId(String(res.data.user_id || res.data.id))
    }
    return res
  }

  function toggleOnline() {
    online.value = !online.value
  }

  function setOnline(val) {
    online.value = val
  }

  function clearSession() {
    accessToken.value = ''
    refreshToken.value = ''
    driverInfo.value = null
    uni.removeStorageSync('driverAccessToken')
    uni.removeStorageSync('driverRefreshToken')
  }

  return {
    userId, accessToken, refreshToken, online, driverInfo, isLoggedIn,
    setUserId, setSession, loadDriverInfo, toggleOnline, setOnline, clearSession
  }
})
