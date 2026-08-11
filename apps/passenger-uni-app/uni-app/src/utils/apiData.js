export function mapOkData(promise, mapper) {
  return promise.then((res) => {
    if (res?.code !== 0) return res
    return { ...res, data: mapper(res.data) }
  })
}

export function unwrap(data, keys = []) {
  if (!data || typeof data !== 'object' || Array.isArray(data)) return data
  for (const key of keys) {
    if (data[key] !== undefined && data[key] !== null) return data[key]
  }
  return data
}

export function listFrom(data) {
  if (Array.isArray(data)) return data
  if (Array.isArray(data?.items)) return data.items
  if (Array.isArray(data?.list)) return data.list
  return []
}

export function normalizeTrip(item = {}) {
  return {
    ...item,
    driver_id: item.driver_id ?? item.driverId,
    depart_time: item.depart_time ?? item.departTime,
    arrive_time: item.arrive_time ?? item.arriveTime,
    seats_total: item.seats_total ?? item.seatsTotal,
    seats_available: item.seats_available ?? item.seatsAvailable,
    created_at: item.created_at ?? item.createdAt,
  }
}

export function normalizeOrder(item = {}) {
  const seats = item.seats ?? item.seats_booked ?? item.seatsBooked
  const totalPrice = item.total_price ?? item.totalPrice ?? item.amount
  const totalPriceText = item.total_price_text ?? item.totalPriceText ?? item.amount_text ?? item.amountText ?? formatMoneyText(totalPrice)
  const totalPriceCents = item.total_price_cents ?? item.totalPriceCents ?? item.amount_cents ?? item.amountCents ?? moneyCents(totalPrice)
  const rawStatus = item.rawStatus ?? item.raw_status ?? item.status
  return {
    ...item,
    id: item.id === undefined || item.id === null ? item.id : String(item.id),
    orderId: item.orderId === undefined || item.orderId === null ? item.orderId : String(item.orderId),
    trip_id: toSafeId(item.trip_id ?? item.tripId),
    passenger_id: toSafeId(item.passenger_id ?? item.passengerId),
    driver_id: toSafeId(item.driver_id ?? item.driverId),
    depart_time: item.depart_time ?? item.departTime,
    seats,
    seats_booked: seats,
    total_price: totalPrice,
    total_price_text: totalPriceText,
    total_price_cents: totalPriceCents,
    amount: totalPrice,
    amount_text: totalPriceText,
    amount_cents: totalPriceCents,
    status: normalizeOrderStatus(rawStatus),
    raw_status: rawStatus,
    created_at: item.created_at ?? item.createdAt,
  }
}

function toSafeId(value) {
  return value === undefined || value === null || value === '' ? value : String(value)
}

function formatMoneyText(value) {
  const number = Number(value)
  if (!Number.isFinite(number)) return '0.00'
  return number.toFixed(2)
}

function moneyCents(value) {
  const number = Number(value)
  if (!Number.isFinite(number)) return 0
  return Math.round(number * 100)
}

export function normalizeOrderStatus(status) {
  const map = {
    0: 'pending',
    1: 'accepted',
    2: 'completed',
    3: 'cancelled',
    4: 'paid',
    5: 'picking_up',
    6: 'delivering',
    pending: 'pending',
    waiting: 'pending',
    waiting_pay: 'waiting_pay',
    paid: 'paid',
    accepted: 'accepted',
    picking_up: 'picking_up',
    delivering: 'delivering',
    ongoing: 'accepted',
    in_progress: 'in_progress',
    completed: 'completed',
    cancelled: 'cancelled',
    canceled: 'cancelled'
  }
  return map[status] || 'unknown'
}

export function normalizePassenger(item = {}) {
  return {
    ...item,
    user_id: item.user_id ?? item.userId ?? item.id,
    name: item.name ?? item.nickname ?? item.nick_name,
    nickname: item.nickname ?? item.nick_name ?? item.name,
    mobile: item.mobile ?? item.phone,
    phone: item.phone ?? item.mobile,
    avatar: item.avatar ?? item.avatar_url ?? item.avatarUrl,
    avatar_url: item.avatar_url ?? item.avatarUrl ?? item.avatar,
    common_address: item.common_address ?? item.commonAddress,
    payment_preference: item.payment_preference ?? item.paymentPreference,
    created_at: item.created_at ?? item.createdAt,
    updated_at: item.updated_at ?? item.updatedAt,
  }
}
