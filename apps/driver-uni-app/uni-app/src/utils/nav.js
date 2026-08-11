const TABBAR_PAGES = [
  'pages/home/home',
  'pages/pendingOrders/pendingOrders',
  'pages/incomeLedger/incomeLedger',
  'pages/locationReport/locationReport',
  'pages/profile/profile'
]

export function navigate(url) {
  const path = String(url).split('?')[0].replace(/^\/+/, '')
  if (TABBAR_PAGES.includes(path)) {
    uni.switchTab({ url })
    return
  }
  uni.navigateTo({ url })
}

export function goDetail(url) {
  uni.navigateTo({ url })
}
