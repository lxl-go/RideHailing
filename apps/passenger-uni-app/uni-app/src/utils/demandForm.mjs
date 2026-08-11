const DAY_MS = 24 * 60 * 60 * 1000

function pad2(value) {
  return String(value).padStart(2, '0')
}

function toDateValue(date) {
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())}`
}

export function buildDateOptions(baseDate = new Date(), days = 7) {
  const start = new Date(baseDate)
  start.setHours(0, 0, 0, 0)
  return Array.from({ length: days }, (_, index) => {
    const date = new Date(start.getTime() + index * DAY_MS)
    return { value: toDateValue(date), date, offset: index }
  })
}

export function formatDateLabel(option) {
  if (!option) return ''
  const prefix = option.offset === 0 ? '今天 ' : option.offset === 1 ? '明天 ' : ''
  const date = option.date instanceof Date ? option.date : new Date(`${option.value}T00:00:00`)
  return `${prefix}${pad2(date.getMonth() + 1)}月${pad2(date.getDate())}日`
}

export function buildTimeOptions(stepMinutes = 30) {
  const options = []
  for (let minutes = 0; minutes < 24 * 60; minutes += stepMinutes) {
    const hour = Math.floor(minutes / 60)
    const minute = minutes % 60
    const value = `${pad2(hour)}:${pad2(minute)}`
    options.push({ value, hour, minute })
  }
  return options
}

export function formatTimeLabel(option) {
  return option?.value || ''
}

export function estimateDemandBudget({ distanceMeters = 0, durationSeconds = 0, seats = 1 } = {}) {
  const distanceKm = Number(distanceMeters) / 1000
  if (!Number.isFinite(distanceKm) || distanceKm <= 0) return ''
  const durationMinutes = Math.max(0, Number(durationSeconds) / 60)
  const seatCount = Math.max(1, Number(seats) || 1)
  const base = 9
  const distanceFee = distanceKm * 3
  const durationFee = durationMinutes * 0.4
  const seatFactor = 1 + (seatCount - 1) * 0.5
  return (Math.ceil((base + distanceFee + durationFee) * seatFactor)).toFixed(2)
}

export function buildDemandPayload(form) {
  return {
    origin: form.origin?.name || '',
    destination: form.destination?.name || '',
    depart_time: `${form.departDate}T${form.departTime}:00+08:00`,
    seats: Number(form.seats || 1),
    budget: Number(form.estimatedBudget || 0),
    remark: String(form.remark || '').trim(),
  }
}
