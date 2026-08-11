<template>
  <div class="ride-page-header">
    <div class="header-left">
      <h2 class="header-title">{{ title }}</h2>
      <p v-if="subtitle" class="header-subtitle">{{ subtitle }}</p>
    </div>
    <div class="header-right">
      <slot />
    </div>
  </div>
</template>

<script setup>
defineOptions({ name: 'RidePageHeader' })
defineProps({
  title: { type: String, required: true },
  subtitle: { type: String, default: '' }
})
</script>

<style scoped>
.ride-page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.header-title {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: #1d2129;
}
.header-subtitle {
  margin: 6px 0 0;
  font-size: 13px;
  color: #86909c;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
