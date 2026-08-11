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
  const id = toSafeId(item.id ?? item.ID)
  const driverId = toSafeId(item.driver_id ?? item.driverId)
  const auditOperatorId = toSafeId(item.audit_operator_id ?? item.auditOperatorId)
  const status = normalizeTripStatus(item.status)
  return {
    ...item,
    id,
    driverId,
    driver_id: driverId,
    depart_time: item.depart_time ?? item.departTime,
    arrive_time: item.arrive_time ?? item.arriveTime,
    seats_total: item.seats_total ?? item.seatsTotal,
    seats_available: item.seats_available ?? item.seatsAvailable,
    auditOperatorId,
    audit_operator_id: auditOperatorId,
    reject_reason: item.reject_reason ?? item.rejectReason,
    route_distance_meters: item.route_distance_meters ?? item.routeDistanceMeters,
    route_duration_seconds: item.route_duration_seconds ?? item.routeDurationSeconds,
    is_deleted: item.is_deleted ?? item.isDeleted,
    status,
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

function normalizeTripStatus(status) {
  const value = Number(status)
  return Number.isFinite(value) ? value : status
}

export function normalizeDriver(item = {}) {
  return {
    ...item,
    user_id: item.user_id ?? item.userId ?? item.id,
    avatar_url: item.avatar_url ?? item.avatarUrl,
    service_status: item.service_status ?? item.serviceStatus,
    certification_status: item.certification_status ?? item.certificationStatus,
    created_at: item.created_at ?? item.createdAt,
    updated_at: item.updated_at ?? item.updatedAt,
  }
}

export function normalizeCertification(item = {}) {
  const statusValue = item.status
  const statusText = statusValue === 3 ? 'approved' : statusValue === 2 ? 'reviewing' : statusValue === 4 ? 'rejected' : statusValue
  return {
    ...item,
    driver_id: item.driver_id ?? item.driverId,
    id_card_no: item.id_card_no ?? item.idCardNo,
    license_type: item.license_type ?? item.licenseType,
    city: item.city ?? '',
    vehicle_license_no: item.vehicle_license_no ?? item.vehicleLicenseNo,
    vehicle_photo_url: item.vehicle_photo_url ?? item.vehiclePhotoUrl,
    face_photo_url: item.face_photo_url ?? item.facePhotoUrl,
    reject_reason: item.reject_reason ?? item.rejectReason,
    status: statusText,
    raw_status: statusValue,
    created_at: item.created_at ?? item.createdAt,
    updated_at: item.updated_at ?? item.updatedAt,
  }
}

export function normalizeVehicle(item = {}) {
  return {
    ...item,
    driver_id: item.driver_id ?? item.driverId,
    plate_no: item.plate_no ?? item.plateNo,
    plateNo: item.plateNo ?? item.plate_no,
    vehicle_type: item.vehicle_type ?? item.vehicleType,
    vehicleType: item.vehicleType ?? item.vehicle_type,
    audit_id: item.audit_id ?? item.auditId,
    auditId: item.auditId ?? item.audit_id,
    review_status: item.review_status ?? item.reviewStatus,
    reviewStatus: item.reviewStatus ?? item.review_status,
    reject_reason: item.reject_reason ?? item.rejectReason,
    rejectReason: item.rejectReason ?? item.reject_reason,
    can_edit: item.can_edit ?? item.canEdit,
    canEdit: item.canEdit ?? item.can_edit,
    can_delete: item.can_delete ?? item.canDelete,
    canDelete: item.canDelete ?? item.can_delete,
    source: item.source || 'vehicle',
    created_at: item.created_at ?? item.createdAt,
    updated_at: item.updated_at ?? item.updatedAt,
  }
}

export function normalizeMessage(item = {}) {
  return {
    ...item,
    id: item.id ?? item.ID,
    topic: item.topic || '',
    title: item.title || '',
    payload: item.payload || '',
    delivered: Boolean(item.delivered),
    created_at: item.created_at ?? item.createdAt,
  }
}

export function mobilePageParams(params = {}) {
  const next = { ...params }
  if (next.page_size !== undefined && next.pageSize === undefined) {
    next.pageSize = next.page_size
    delete next.page_size
  }
  if (next.trip_id !== undefined && next.tripId === undefined) {
    next.tripId = next.trip_id
    delete next.trip_id
  }
  if (next.order_id !== undefined && next.orderId === undefined) {
    next.orderId = next.order_id
    delete next.order_id
  }
  return next
}
