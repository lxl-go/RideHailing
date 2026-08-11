const toIDString = (value) => {
  if (value === undefined || value === null || value === '') return ''
  return String(value)
}

export const normalizeVehicleForView = (item = {}) => ({
  ...item,
  id: toIDString(item.id ?? item.ID),
  auditId: toIDString(item.auditId ?? item.audit_id),
  source: item.source || 'vehicle',
  plateNo: item.plateNo || item.plate_no || item.vehicle_no || item.plate || '',
  brand: item.brand || '',
  model: item.model || '',
  vehicleType: item.vehicleType || item.vehicle_type || item.model || '',
  color: item.color || '',
  seats: Number(item.seats) || 4,
  status: Number(item.status ?? item.reviewStatus ?? item.review_status ?? 1),
  reviewStatus: Number(item.reviewStatus ?? item.review_status ?? item.status ?? 1),
  canEdit: item.canEdit ?? item.can_edit,
  canDelete: item.canDelete ?? item.can_delete,
})
