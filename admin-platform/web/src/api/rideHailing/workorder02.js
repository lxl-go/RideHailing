import service from '@/utils/request'

export const listShuttleLines = (params = {}) => {
  return service({
    url: '/carpool/shuttle/line/list',
    method: 'get',
    params
  })
}

export const getShuttleLine = (id) => {
  return service({
    url: `/carpool/shuttle/line/${id}`,
    method: 'get'
  })
}

export const createShuttleLine = (data) => {
  return service({
    url: '/carpool/shuttle/line',
    method: 'post',
    data
  })
}

export const updateShuttleLine = (id, data) => {
  return service({
    url: `/carpool/shuttle/line/${id}`,
    method: 'put',
    data
  })
}

export const publishShuttleLines = (ids) => {
  return service({
    url: '/carpool/shuttle/line/publish',
    method: 'post',
    data: { ids }
  })
}

export const exportShuttleLines = () => {
  return service({
    url: '/carpool/shuttle/line/export',
    method: 'post'
  })
}
