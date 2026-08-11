const TABBAR_PAGES = [
  'pages/home/home',
  'pages/orders/orders',
  'pages/tracking/tracking',
  'pages/aiAssistant/aiAssistant',
  'pages/profile/profile'
]

export function navigate(url) {
  const path = String(url).split('?')[0]
  if (TABBAR_PAGES.includes(path)) {
    uni.switchTab({ url })
    return
  }
  uni.navigateTo({ url })
}

export function goDetail(url) {
  uni.navigateTo({ url })
}
