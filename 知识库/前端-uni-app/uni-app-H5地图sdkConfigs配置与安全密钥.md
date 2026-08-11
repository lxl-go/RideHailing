# uni-app H5 端地图：manifest 配置 sdkConfigs 与安全密钥

- 创建日期：2026-08-10
- 最近更新：2026-08-10
- 标签：uni-app, H5, 高德地图, map组件, manifest.json

## 适用场景

uni-app 项目在 H5 端使用原生 `<map>` 组件，但页面只有空白区域、地图不渲染；或需要在 App / H5 / 小程序多端共享地图能力，希望 `<map>` 常显并动态更新 marker。

## 问题背景

`manifest.json` 顶层写了 `aMapKey`，但 H5 端地图仍空白。原因：**H5 端 `<map>` 需要把高德 key 配置到 `h5.sdkConfigs.maps.amap`，而不是只用顶层 `aMapKey`**（顶层 aMapKey 主要供原生 App 端 / 各平台构建期读取）。

```json
{
  "aMapKey": "22ba26c4d757d904aef8138acda60ab7",
  "h5": {
    "router": { "mode": "hash" },
    "sdkConfigs": {
      "maps": {
        "amap": {
          "key": "22ba26c4d757d904aef8138acda60ab7",
          "securityJsCode": ""
        }
      }
    }
  }
}
```

## 核心结论

1. H5 `<map>` 渲染依赖 `h5.sdkConfigs.maps.amap.key`。
2. **2021-12-02 之后新申请的 key 必须配合 `securityJsCode`（高德安全密钥）**，且要求设置在高德 JS API 脚本加载之前。没有它会出现地图空白/权限报错（INVALID_USER_SCODE 等）。
3. `securityJsCode` 需要在高德开放平台控制台的应用详情里生成，与 key 一一对应；项目侧把生成值填入 `securityJsCode` 字段即可。

## 关键原理：模板化 map 组件的使用

uni-app 的 `<map>` 是跨端组件，属性包括：

- `latitude` / `longitude`：地图中心（必填，否则 H5 端报错）
- `scale`：缩放级别
- `markers`：标记点数组
- `include-points`：让地图自动缩放以包含这些点（多标记自动适应视野）

多端（App/H5）都支持 `include-points`，可避免手工算中心点与缩放级别。

## 示例：常显地图 + 动态 marker（发布页）

```vue
<template>
  <view class="map-card">
    <map
      class="route-map"
      :latitude="mapCenter.latitude"
      :longitude="mapCenter.longitude"
      :scale="mapScale"
      :markers="mapMarkers"
      :include-points="mapIncludePoints"
    />
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'

const form = ref({
  originLocation: null,
  destinationLocation: null,
})

const mapMarkers = computed(() => {
  const list = []
  const o = form.value.originLocation
  const d = form.value.destinationLocation
  if (o?.longitude) list.push({ id: 1, latitude: o.latitude, longitude: o.longitude, title: '出发' })
  if (d?.longitude) list.push({ id: 2, latitude: d.latitude, longitude: d.longitude, title: '到达' })
  return list
})

const mapIncludePoints = computed(() => {
  const points = []
  for (const loc of [form.value.originLocation, form.value.destinationLocation]) {
    if (loc?.longitude) points.push({ latitude: loc.latitude, longitude: loc.longitude })
  }
  return points
})

const mapScale = computed(() => (mapIncludePoints.value.length >= 2 ? 12 : 15))
</script>
```

## 常见误区或注意事项

- 只配顶层 `aMapKey` 不够，H5 必须配 `h5.sdkConfigs.maps.amap`。
- `securityJsCode` 与 key 不匹配、或漏配，会在控制台出现权限类错误/空白。
- 地图容器必须有显式高度（`<map>` 默认高度可能为 0），否则空白。
- `include-points` 用经纬度数组即可自动适配视野，不必手写中心计算。
- 若页面无中心点时给默认城市中心（如北京天安门），避免 map 缺坐标报错。

## 延伸方向

- 生产环境建议「代理转发」方式配置安全密钥（`_AMapSecurityConfig.serviceHost`），避免明文暴露。
- 需要在地图上画路线时，可先由服务端（高德 Web 服务）取 polyline 坐标再交给 `<map>` 的 `polyline` 属性，保持多端一致。