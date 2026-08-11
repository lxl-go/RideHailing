<template>
  <view class="page">
    <!-- #ifdef H5 -->
    <view id="demand-amap" class="route-map amap-h5"></view>
    <!-- #endif -->
    <!-- #ifndef H5 -->
    <map
      class="route-map"
      :latitude="mapCenter.latitude"
      :longitude="mapCenter.longitude"
      :scale="mapScale"
      :markers="mapMarkers"
      :polyline="mapPolyline"
      show-location
    />
    <!-- #endif -->

    <view class="form-card">
      <view class="form-line" @click="openSearch('origin')">
        <text class="form-label">出发地</text>
        <text :class="['form-value', form.origin ? '' : 'placeholder']">
          {{ form.origin?.name || '搜索出发地点' }}
        </text>
      </view>

      <view class="form-line" @click="openSearch('destination')">
        <text class="form-label">目的地</text>
        <text :class="['form-value', form.destination ? '' : 'placeholder']">
          {{ form.destination?.name || '搜索目的地点' }}
        </text>
      </view>

      <picker mode="selector" :range="dateLabels" @change="selectDate">
        <view class="form-line">
          <text class="form-label">出发日期</text>
          <text :class="['select-value', form.departDate ? '' : 'placeholder']">
            {{ selectedDateLabel || '请选择出发日期' }}
          </text>
        </view>
      </picker>

      <picker mode="selector" :range="timeLabels" @change="selectTime">
        <view class="form-line">
          <text class="form-label">出发时刻</text>
          <text :class="['select-value', form.departTime ? '' : 'placeholder']">
            {{ selectedTimeLabel || '请选择出发时刻' }}
          </text>
        </view>
      </picker>

      <view class="form-line">
        <text class="form-label">座位数</text>
        <view class="stepper">
          <button class="stepper-btn" :disabled="form.seats <= 1" @click="changeSeats(-1)">-</button>
          <text class="seat-count">{{ form.seats }}</text>
          <button class="stepper-btn" :disabled="form.seats >= 6" @click="changeSeats(1)">+</button>
        </view>
      </view>

      <view class="form-line">
        <text class="form-label">预估价格</text>
        <text :class="['form-value price-value', form.estimatedBudget ? '' : 'placeholder']">
          {{ form.estimatedBudget ? `¥${form.estimatedBudget}` : '选择地点后自动预估' }}
        </text>
      </view>

      <view class="form-line remark-line">
        <text class="form-label">备注</text>
        <input v-model="form.remark" class="remark-input" placeholder="选填" placeholder-class="input-placeholder" />
      </view>
    </view>

    <button class="submit-btn" :loading="submitting" :disabled="submitting" @click="submit">发布需求</button>

    <view v-if="activeSearch" class="search-mask" @click="closeSearch">
      <view class="search-panel" @click.stop>
        <view class="search-header">
          <text class="search-title">{{ activeSearch === 'origin' ? '选择出发地' : '选择目的地' }}</text>
          <text class="search-close" @click="closeSearch">取消</text>
        </view>
        <input
          v-model="searchKeyword"
          class="search-input"
          :placeholder="activeSearch === 'origin' ? '搜索出发地点' : '搜索目的地点'"
          placeholder-class="input-placeholder"
          confirm-type="search"
          focus
          @input="handleSearchInput"
        />
        <view class="suggestion-list">
          <view v-if="searchLoading" class="empty-text">搜索中...</view>
          <view v-else-if="!suggestions.length" class="empty-text">请输入地点名称进行搜索</view>
          <view v-for="item in suggestions" :key="item.id" class="suggestion-item" @click="selectSuggestion(item)">
            <text class="suggestion-name">{{ item.name }}</text>
            <text class="suggestion-address">{{ item.district }} {{ item.address }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'

import { publishDemand } from '@/api/trip'
import { getAmapDrivingRoute, getAmapInputTips } from '@/api/map'
import {
  buildDateOptions,
  buildDemandPayload,
  buildTimeOptions,
  estimateDemandBudget,
  formatDateLabel,
  formatTimeLabel,
} from '@/utils/demandForm.mjs'

const defaultCenter = { latitude: 39.9042, longitude: 116.4074 }
const amapWebKey = '22ba26c4d757d904aef8138acda60ab7'
const dateOptions = buildDateOptions(new Date(), 14)
const timeOptions = buildTimeOptions(30)

const form = reactive({
  origin: null,
  destination: null,
  departDate: '',
  departTime: '',
  seats: 1,
  estimatedBudget: '',
  remark: '',
})

const routePreview = ref({ distanceMeters: 0, durationSeconds: 0, points: [] })
const submitting = ref(false)
const activeSearch = ref('')
const searchKeyword = ref('')
const suggestions = ref([])
const searchLoading = ref(false)
let searchTimer = null

const dateLabels = computed(() => dateOptions.map(formatDateLabel))
const timeLabels = computed(() => timeOptions.map(formatTimeLabel))
const selectedDateLabel = computed(() => {
  const item = dateOptions.find((option) => option.value === form.departDate)
  return item ? formatDateLabel(item) : ''
})
const selectedTimeLabel = computed(() => {
  const item = timeOptions.find((option) => option.value === form.departTime)
  return item ? formatTimeLabel(item) : ''
})

const mapCenter = computed(() => form.origin || form.destination || defaultCenter)
const mapScale = computed(() => (form.origin && form.destination ? 11 : 13))
const mapMarkers = computed(() => {
  const markers = []
  if (form.origin) {
    markers.push({ id: 1, latitude: form.origin.latitude, longitude: form.origin.longitude, title: '出发地' })
  }
  if (form.destination) {
    markers.push({ id: 2, latitude: form.destination.latitude, longitude: form.destination.longitude, title: '目的地' })
  }
  return markers
})
const mapPolyline = computed(() => {
  const points = routePreview.value.points || []
  if (!points.length) return []
  return [{ points, color: '#1677ff', width: 5, dottedLine: false, arrowLine: true }]
})

// #ifdef H5
let amapLoader = null
let h5Map = null
let h5Markers = []
let h5Polyline = null

function loadAmap() {
  if (typeof window === 'undefined') return Promise.resolve(null)
  if (window.AMap) return Promise.resolve(window.AMap)
  if (amapLoader) return amapLoader
  amapLoader = new Promise((resolve) => {
    const existing = document.getElementById('amap-web-sdk')
    if (existing) {
      existing.addEventListener('load', () => resolve(window.AMap || null), { once: true })
      existing.addEventListener('error', () => resolve(null), { once: true })
      return
    }
    const script = document.createElement('script')
    script.id = 'amap-web-sdk'
    script.src = `https://webapi.amap.com/maps?v=2.0&key=${amapWebKey}`
    script.onload = () => resolve(window.AMap || null)
    script.onerror = () => resolve(null)
    document.head.appendChild(script)
  })
  return amapLoader
}

function toLngLat(point) {
  return [point.longitude, point.latitude]
}

async function renderH5Map() {
  await nextTick()
  const AMap = await loadAmap()
  const container = typeof document !== 'undefined' ? document.getElementById('demand-amap') : null
  if (!AMap || !container) return
  const center = toLngLat(mapCenter.value)
  if (!h5Map) {
    h5Map = new AMap.Map('demand-amap', {
      center,
      zoom: mapScale.value,
      resizeEnable: true,
      viewMode: '2D',
    })
  } else {
    h5Map.setZoomAndCenter(mapScale.value, center)
  }
  if (h5Markers.length) {
    h5Map.remove(h5Markers)
    h5Markers = []
  }
  h5Markers = mapMarkers.value.map((marker) =>
    new AMap.Marker({
      position: [marker.longitude, marker.latitude],
      title: marker.title,
      label: { content: marker.title, direction: 'top' },
    })
  )
  if (h5Markers.length) h5Map.add(h5Markers)
  if (h5Polyline) {
    h5Map.remove(h5Polyline)
    h5Polyline = null
  }
  const routePoints = routePreview.value.points || []
  if (routePoints.length) {
    h5Polyline = new AMap.Polyline({
      path: routePoints.map(toLngLat),
      strokeColor: '#1677ff',
      strokeWeight: 6,
      lineJoin: 'round',
    })
    h5Map.add(h5Polyline)
    h5Map.setFitView([...h5Markers, h5Polyline], false, [28, 28, 28, 28])
  }
}

onMounted(renderH5Map)
watch([mapCenter, mapMarkers, mapPolyline], renderH5Map, { deep: true })
// #endif

function selectDate(event) {
  const item = dateOptions[Number(event.detail.value)]
  form.departDate = item?.value || ''
}

function selectTime(event) {
  const item = timeOptions[Number(event.detail.value)]
  form.departTime = item?.value || ''
}

function changeSeats(delta) {
  form.seats = Math.min(6, Math.max(1, form.seats + delta))
  refreshBudget()
}

function openSearch(type) {
  activeSearch.value = type
  const selected = type === 'origin' ? form.origin : form.destination
  searchKeyword.value = selected?.name || ''
  suggestions.value = []
}

function closeSearch() {
  activeSearch.value = ''
  searchKeyword.value = ''
  suggestions.value = []
  searchLoading.value = false
  if (searchTimer) clearTimeout(searchTimer)
}

function handleSearchInput(event) {
  const keyword = event.detail?.value ?? searchKeyword.value
  searchKeyword.value = keyword
  if (searchTimer) clearTimeout(searchTimer)
  if (!String(keyword || '').trim()) {
    suggestions.value = []
    return
  }
  searchTimer = setTimeout(async () => {
    searchLoading.value = true
    suggestions.value = await getAmapInputTips(searchKeyword.value)
    searchLoading.value = false
  }, 260)
}

async function selectSuggestion(item) {
  if (activeSearch.value === 'origin') form.origin = item
  if (activeSearch.value === 'destination') form.destination = item
  closeSearch()
  await refreshRoute()
}

async function refreshRoute() {
  if (!form.origin || !form.destination) {
    routePreview.value = { distanceMeters: 0, durationSeconds: 0, points: [] }
    refreshBudget()
    return
  }
  routePreview.value = await getAmapDrivingRoute(form.origin, form.destination)
  refreshBudget()
}

function refreshBudget() {
  form.estimatedBudget = estimateDemandBudget({
    distanceMeters: routePreview.value.distanceMeters,
    durationSeconds: routePreview.value.durationSeconds,
    seats: form.seats,
  })
}

async function submit() {
  if (!form.origin || !form.destination) return uni.showToast({ title: '请选择出发地与目的地', icon: 'none' })
  if (!form.departDate) return uni.showToast({ title: '请选择出发日期', icon: 'none' })
  if (!form.departTime) return uni.showToast({ title: '请选择出发时刻', icon: 'none' })
  if (!form.estimatedBudget) return uni.showToast({ title: '路线预估失败，请重新选择地点', icon: 'none' })
  submitting.value = true
  try {
    const res = await publishDemand(buildDemandPayload(form))
    if (res.code === 0) {
      uni.showToast({ title: '发布成功', icon: 'success' })
      setTimeout(() => uni.navigateBack(), 600)
    } else {
      uni.showToast({ title: res.msg || '发布失败', icon: 'none' })
    }
  } catch (e) {
    uni.showToast({ title: e?.msg || '网络异常，发布失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 24rpx;
  background: #f3f6fb;
  box-sizing: border-box;
}

.route-map {
  width: 100%;
  height: 356rpx;
  border-radius: 18rpx;
  overflow: hidden;
}

.form-card {
  margin-top: 22rpx;
  padding: 28rpx 24rpx;
  background: #fff;
  border-radius: 24rpx;
  box-shadow: 0 12rpx 32rpx rgba(15, 23, 42, 0.06);
}

.form-line {
  min-height: 88rpx;
  display: flex;
  align-items: center;
  gap: 26rpx;
}

.form-line + .form-line {
  border-top: 1rpx solid #eef2f7;
}

.form-label {
  width: 150rpx;
  flex: 0 0 150rpx;
  color: #172033;
  font-size: 30rpx;
  font-weight: 500;
}

.form-value,
.select-value {
  flex: 1;
  min-width: 0;
  color: #1f2937;
  font-size: 30rpx;
  line-height: 1.4;
}

.placeholder,
.input-placeholder {
  color: #a3adbd;
}

.price-value {
  color: #1677ff;
  font-weight: 600;
}

.stepper {
  display: flex;
  align-items: center;
  gap: 22rpx;
}

.stepper-btn {
  width: 56rpx;
  height: 56rpx;
  line-height: 52rpx;
  padding: 0;
  border-radius: 50%;
  border: 1rpx solid #d7dfeb;
  background: #f8fafc;
  color: #1f2937;
  font-size: 34rpx;
}

.stepper-btn::after,
.submit-btn::after {
  border: 0;
}

.seat-count {
  width: 42rpx;
  text-align: center;
  color: #1f2937;
  font-size: 30rpx;
}

.remark-line {
  align-items: center;
}

.remark-input {
  flex: 1;
  min-width: 0;
  height: 72rpx;
  color: #1f2937;
  font-size: 30rpx;
}

.submit-btn {
  margin-top: 22rpx;
  height: 86rpx;
  line-height: 86rpx;
  border-radius: 10rpx;
  background: #2f8cff;
  color: #fff;
  font-size: 30rpx;
  font-weight: 600;
}

.search-mask {
  position: fixed;
  inset: 0;
  z-index: 20;
  display: flex;
  align-items: flex-end;
  background: rgba(15, 23, 42, 0.36);
}

.search-panel {
  width: 100%;
  max-height: 74vh;
  padding: 28rpx 24rpx calc(24rpx + env(safe-area-inset-bottom));
  border-radius: 28rpx 28rpx 0 0;
  background: #fff;
  box-sizing: border-box;
}

.search-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 22rpx;
}

.search-title {
  color: #172033;
  font-size: 32rpx;
  font-weight: 600;
}

.search-close {
  color: #1677ff;
  font-size: 28rpx;
}

.search-input {
  height: 78rpx;
  padding: 0 24rpx;
  border-radius: 14rpx;
  background: #f3f6fb;
  color: #1f2937;
  font-size: 30rpx;
  box-sizing: border-box;
}

.suggestion-list {
  max-height: 52vh;
  margin-top: 18rpx;
  overflow: auto;
}

.suggestion-item {
  padding: 22rpx 0;
  border-bottom: 1rpx solid #edf1f6;
}

.suggestion-name {
  display: block;
  color: #172033;
  font-size: 30rpx;
  font-weight: 500;
}

.suggestion-address {
  display: block;
  margin-top: 8rpx;
  color: #7b8798;
  font-size: 24rpx;
  line-height: 1.4;
}

.empty-text {
  padding: 48rpx 0;
  color: #98a2b3;
  font-size: 26rpx;
  text-align: center;
}
</style>
