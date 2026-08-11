import service from '@/utils/request'

export const listCertAudits = (params) => {
  return service({
    url: '/carpool/review/list',
    method: 'get',
    params
  })
}

export const getCertAudit = (id) => {
  return service({
    url: `/carpool/review/${id}`,
    method: 'get'
  })
}

export const approveCertAudit = (id) => {
  return service({
    url: `/carpool/review/${id}/approve`,
    method: 'post'
  })
}

export const rejectCertAudit = (id, data) => {
  return service({
    url: `/carpool/review/${id}/reject`,
    method: 'post',
    data
  })
}

export const listVehicleReviews = (params) => {
  return service({
    url: '/carpool/review/vehicle/list',
    method: 'get',
    params
  })
}

export const handleVehicleReview = (id, data) => {
  return service({
    url: `/carpool/review/vehicle/${id}/action`,
    method: 'post',
    data
  })
}

export const getAuditList = listCertAudits

export const getAuditDetail = getCertAudit

export const approveAudit = approveCertAudit

export const rejectAudit = rejectCertAudit

export const getVehicleReviewList = listVehicleReviews

export const getTripList = (params) => {
  return service({
    url: '/carpool/trip/list',
    method: 'get',
    params
  })
}

export const getTripDetail = (id) => {
  return service({
    url: `/carpool/trip/${id}`,
    method: 'get'
  })
}

export const deactivateTrip = (id, data) => {
  return service({
    url: `/carpool/trip/${id}/deactivate`,
    method: 'post',
    data
  })
}
