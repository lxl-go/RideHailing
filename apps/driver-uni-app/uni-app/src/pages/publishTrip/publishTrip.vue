<template>
  <view class="page">
    <view class="card map-card">
      <map
        class="route-map"
        :latitude="mapCenter.latitude"
        :longitude="mapCenter.longitude"
        :scale="mapScale"
        :markers="mapMarkers"
        :include-points="mapIncludePoints"
      />
    </view>

    <view class="card">
      <u-form :model="form" label-position="left" label-width="150rpx">
        <u-form-item label="出发地">
          <view class="loc-field">
            <u-input
              v-model="form.origin"
              border
              placeholder="搜索出发地点"
              @focus="onOriginFocus"
              @blur="hideOriginSuggests"
              @input="onOriginInput"
            />
            <view v-if="originSuggests.length" class="suggest-list" @mousedown.stop @touchstart.stop>
              <view v-for="(item, idx) in originSuggests" :key="idx" class="suggest-item" @click="pickLocation('origin', item)">
                <text class="suggest-name">{{ item.name }}</text>
                <text v-if="item.formatted_address || item.formattedAddress" class="suggest-addr">{{ item.formatted_address || item.formattedAddress }}</text>
              </view>
            </view>
          </view>
        </u-form-item>
        <u-form-item label="目的地">
          <view class="loc-field">
            <u-input
              v-model="form.destination"
              border
              placeholder="搜索目的地点"
              @focus="onDestinationFocus"
              @blur="hideDestinationSuggests"
              @input="onDestinationInput"
            />
            <view v-if="destinationSuggests.length" class="suggest-list" @mousedown.stop @touchstart.stop>
              <view v-for="(item, idx) in destinationSuggests" :key="idx" class="suggest-item" @click="pickLocation('destination', item)">
                <text class="suggest-name">{{ item.name }}</text>
                <text v-if="item.formatted_address || item.formattedAddress" class="suggest-addr">{{ item.formatted_address || item.formattedAddress }}</text>
              </view>
            </view>
          </view>
        </u-form-item>
        <u-form-item label="出发日期">
          <picker mode="date" :value="form.depart_date" @change="onDepartDate">
            <view class="picker-field" :class="{ 'is-placeholder': !form.depart_date }">
              <text class="picker-text">{{ form.depart_date || '请选择出发日期' }}</text>
            </view>
          </picker>
        </u-form-item>
        <u-form-item label="出发时刻">
          <picker mode="time" :value="form.depart_time" @change="onDepartTime">
            <view class="picker-field" :class="{ 'is-placeholder': !form.depart_time }">
              <text class="picker-text">{{ form.depart_time || '请选择出发时刻' }}</text>
            </view>
          </picker>
        </u-form-item>
        <u-form-item label="座位数"><u-input v-model.number="form.seats" type="number" border placeholder="4" /></u-form-item>
        <u-form-item label="备注"><u-input v-model="form.remark" border placeholder="选填" /></u-form-item>
      </u-form>
    </view>

    <u-button type="primary" :loading="submitting" @click="submit">发布行程</u-button>
  </view>
</template>

<script setup>
import { computed, onUnmounted, reactive, ref, watch } from 'vue'
import { publishTrip, suggestLocations } from '@/api/trip'

const form = reactive({
  origin: '', destination: '', originLocation: null, destinationLocation: null,
  depart_date: '',
  depart_time: '',
  seats: 4, remark: '',
})
const submitting = ref(false)
const originSuggests = ref([])
const destinationSuggests = ref([])
const suggestTimers = { origin: null, destination: null }
const lastSuggestSeq = { origin: 0, destination: 0 }
let lastSuggestErrorAt = 0

const parseCoordinate = (item = {}) => {
  const location = item.location ?? item.lnglat ?? item.lngLat
  if (typeof location === 'string') {
    const [longitude, latitude] = location.split(',').map((value) => Number(value))
    return { longitude: longitude || 0, latitude: latitude || 0 }
  }
  if (location && typeof location === 'object') {
    return {
      longitude: Number(location.longitude ?? location.lng ?? location[0] ?? 0),
      latitude: Number(location.latitude ?? location.lat ?? location[1] ?? 0),
    }
  }
  return {
    longitude: Number(item.longitude ?? item.lng ?? 0),
    latitude: Number(item.latitude ?? item.lat ?? 0),
  }
}

const normalizeLocation = (item = {}) => ({
  poiId: item.poi_id ?? item.poiId ?? '',
  name: item.name ?? item.title ?? item.address ?? '',
  formattedAddress: item.formatted_address ?? item.formattedAddress ?? item.address ?? item.district ?? item.adname ?? '',
  ...parseCoordinate(item),
})

const extractLocationList = (res = {}) => {
  const data = res.data
  if (Array.isArray(data)) return data
  const candidates = [
    data?.locations,
    data?.tips,
    data?.pois,
    data?.list,
    data?.items,
    data?.records,
    data?.data,
    res.locations,
    res.tips,
    res.pois,
    res.list,
    res.items,
  ]
  return candidates.find((item) => Array.isArray(item)) || []
}

const setSuggests = (field, list) => {
  if (field === 'origin') originSuggests.value = list
  else destinationSuggests.value = list
}

const fetchSuggests = async (keyword, field) => {
  const trimmed = (keyword || '').trim()
  if (!trimmed) {
    setSuggests(field, [])
    return
  }
  const seq = ++lastSuggestSeq[field]
  try {
    const res = await suggestLocations({ keyword: trimmed, limit: 8 })
    if (seq !== lastSuggestSeq[field]) return
    if (res?.code !== 0) {
      setSuggests(field, [])
      const now = Date.now()
      if (now - lastSuggestErrorAt > 2500) {
        lastSuggestErrorAt = now
        uni.showToast({ title: res?.msg || '地址联想失败，请稍后重试', icon: 'none' })
      }
      return
    }
    const list = extractLocationList(res)
    const mapped = (list || []).map(normalizeLocation).filter((item) => item.name)
    setSuggests(field, mapped)
  } catch (e) {
    if (seq !== lastSuggestSeq[field]) return
    setSuggests(field, [])
    uni.showToast({ title: '地址联想失败，请稍后重试', icon: 'none' })
  }
}

const scheduleSuggest = (keyword, field) => {
  if (suggestTimers[field]) clearTimeout(suggestTimers[field])
  suggestTimers[field] = setTimeout(() => fetchSuggests(keyword, field), 300)
}

const inputValue = (event) => event?.detail?.value ?? event?.target?.value ?? event ?? ''
const onOriginInput = (event) => { form.origin = inputValue(event) }
const onDestinationInput = (event) => { form.destination = inputValue(event) }
const onOriginFocus = () => { if (form.origin) scheduleSuggest(form.origin, 'origin') }
const onDestinationFocus = () => { if (form.destination) scheduleSuggest(form.destination, 'destination') }

const hideOriginSuggests = () => setTimeout(() => { originSuggests.value = [] }, 200)
const hideDestinationSuggests = () => setTimeout(() => { destinationSuggests.value = [] }, 200)

watch(() => form.origin, (value) => {
  if (form.originLocation && form.originLocation.name === value) {
    originSuggests.value = []
  } else {
    form.originLocation = null
    scheduleSuggest(value, 'origin')
  }
})
watch(() => form.destination, (value) => {
  if (form.destinationLocation && form.destinationLocation.name === value) {
    destinationSuggests.value = []
  } else {
    form.destinationLocation = null
    scheduleSuggest(value, 'destination')
  }
})

const pickLocation = (field, item) => {
  const loc = normalizeLocation(item)
  if (field === 'origin') {
    form.origin = loc.name
    form.originLocation = loc
    originSuggests.value = []
  } else {
    form.destination = loc.name
    form.destinationLocation = loc
    destinationSuggests.value = []
  }
}

const DEFAULT_CENTER = { latitude: 39.9042, longitude: 116.4074 }

const mapMarkers = computed(() => {
  const list = []
  const o = form.originLocation
  const d = form.destinationLocation
  if (o && o.longitude) {
    list.push({
      id: 1,
      latitude: o.latitude,
      longitude: o.longitude,
      title: '出发',
      width: 28,
      height: 28,
      label: { content: '出发', color: '#ffffff', bgColor: '#2979ff', borderRadius: 6, padding: 4, fontSize: 12 },
    })
  }
  if (d && d.longitude) {
    list.push({
      id: 2,
      latitude: d.latitude,
      longitude: d.longitude,
      title: '到达',
      width: 28,
      height: 28,
      label: { content: '到达', color: '#ffffff', bgColor: '#f56c6c', borderRadius: 6, padding: 4, fontSize: 12 },
    })
  }
  return list
})

const mapCenter = computed(() => {
  const o = form.originLocation
  const d = form.destinationLocation
  if (o && o.longitude) return { latitude: o.latitude, longitude: o.longitude }
  if (d && d.longitude) return { latitude: d.latitude, longitude: d.longitude }
  return DEFAULT_CENTER
})

const mapIncludePoints = computed(() => {
  const points = []
  const o = form.originLocation
  const d = form.destinationLocation
  if (o && o.longitude) points.push({ latitude: o.latitude, longitude: o.longitude })
  if (d && d.longitude) points.push({ latitude: d.latitude, longitude: d.longitude })
  return points
})

const mapScale = computed(() => (mapIncludePoints.value.length >= 2 ? 12 : 15))

const onDepartDate = (e) => {
  form.depart_date = e.detail.value
}
const onDepartTime = (e) => {
  form.depart_time = e.detail.value
}

const submit = async () => {
  if (!form.origin || !form.destination) return uni.showToast({ title: '请填写出发地与目的地', icon: 'none' })
  if (!form.originLocation || !form.destinationLocation) {
    return uni.showToast({ title: '请从联想列表中选择有效地点', icon: 'none' })
  }
  if (!form.depart_date || !form.depart_time) {
    return uni.showToast({ title: '请选择出发日期和时刻', icon: 'none' })
  }
  submitting.value = true
  const departAt = new Date(`${form.depart_date}T${form.depart_time}:00+08:00`)
  if (Number.isNaN(departAt.getTime())) {
    submitting.value = false
    return uni.showToast({ title: '请选择有效的出发时间', icon: 'none' })
  }
  if (departAt.getTime() < Date.now() + 15 * 60 * 1000) {
    submitting.value = false
    return uni.showToast({ title: '出发时间至少晚于当前时间15分钟', icon: 'none' })
  }
  const seats = Number(form.seats) || 4
  if (seats < 1 || seats > 6) {
    submitting.value = false
    return uni.showToast({ title: '座位数范围为1-6', icon: 'none' })
  }
  const payload = {
    origin: form.originLocation,
    destination: form.destinationLocation,
    depart_time: `${form.depart_date}T${form.depart_time}:00+08:00`,
    seats_total: seats,
    remark: form.remark,
  }
  try {
    const res = await publishTrip(payload)
    if (res.code === 0) {
      uni.showToast({ title: '发布成功', icon: 'success' })
      setTimeout(() => uni.navigateBack(), 600)
    } else {
      uni.showToast({ title: res.msg || '发布失败', icon: 'none' })
    }
  } catch (e) {
    uni.showToast({ title: (e && e.msg) || '网络异常，发布失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

onUnmounted(() => {
  Object.values(suggestTimers).forEach((timer) => {
    if (timer) clearTimeout(timer)
  })
})
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f4f7fb; }
.card { background: #fff; border-radius: 24rpx; padding: 24rpx; margin-bottom: 24rpx; box-shadow: 0 6rpx 20rpx rgba(16,24,40,0.05); }
.form-row { display: flex; gap: 24rpx; }
.half { flex: 1; }
.form-row .u-form-item { margin-bottom: 0; }
.form-row .u-form-item__content { flex: 1; }
.loc-field { position: relative; flex: 1; }
.picker-field {
  width: 100%;
  min-height: 72rpx;
  padding: 0 24rpx;
  border: 1rpx solid #dcdfe6;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  background: #fff;
}
.picker-text {
  font-size: 28rpx;
  color: #303133;
  line-height: 1.4;
}
.picker-field.is-placeholder .picker-text {
  color: #c0c4cc;
}
.suggest-list {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  z-index: 99;
  background: #fff;
  border-radius: 12rpx;
  margin-top: 8rpx;
  box-shadow: 0 12rpx 32rpx rgba(16,24,40,0.16);
  max-height: 560rpx;
  overflow-y: auto;
}
.suggest-item {
  padding: 20rpx 24rpx;
  border-bottom: 1rpx solid #f0f2f5;
}
.suggest-item:last-child { border-bottom: none; }
.suggest-name { display: block; font-size: 28rpx; color: #1f2937; font-weight: 500; }
.suggest-addr { display: block; margin-top: 4rpx; font-size: 24rpx; color: #98a2b3; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
.map-card {
  overflow: hidden;
  position: relative;
  padding: 0;
}
.route-map {
  width: 100%;
  height: 480rpx;
  display: block;
}
</style>
