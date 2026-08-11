import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getProfile } from '@/api/profile'

export const useUserStore = defineStore('user', () => {
  const userId = ref('')
  const accessToken = ref('')
  const refreshToken = ref('')
  const profile = ref(null)

  const isLoggedIn = computed(() => Boolean(accessToken.value))

  function setUserId(id) {
    userId.value = id
    uni.setStorageSync('passengerUserId', id)
  }

  function setSession(session) {
    const nextUserId = String(session?.user_id || session?.userId || session?.userID || userId.value || '')
    const token = session?.access_token || session?.accessToken || ''
    const nextRefreshToken = session?.refresh_token || session?.refreshToken || ''
    setUserId(nextUserId)
    accessToken.value = token
    refreshToken.value = nextRefreshToken
    if (token) uni.setStorageSync('passengerAccessToken', token)
    if (nextRefreshToken) uni.setStorageSync('passengerRefreshToken', nextRefreshToken)
  }

  async function loadProfile() {
    const res = await getProfile()
    if (res.code === 0) {
      profile.value = res.data
      if (res.data?.user_id || res.data?.id) {
        setUserId(String(res.data.user_id || res.data.id))
      }
    }
    return res
  }

  function clearSession() {
    accessToken.value = ''
    refreshToken.value = ''
    profile.value = null
    uni.removeStorageSync('passengerAccessToken')
    uni.removeStorageSync('passengerRefreshToken')
  }

  return { userId, accessToken, refreshToken, profile, isLoggedIn, setUserId, setSession, loadProfile, clearSession }
})
