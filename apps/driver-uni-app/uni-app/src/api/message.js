import { request } from '@/utils/request'
import { mapOkData, normalizeMessage } from '@/utils/apiData'

export const listMessages = () =>
  mapOkData(
    request({ url: '/carpool/drivers/messages', method: 'GET', silent: true }),
    (payload = {}) => ({
      ...payload,
      items: (payload.items || payload.list || []).map(normalizeMessage),
    })
  )

export const ackMessage = (id) =>
  request({ url: `/carpool/drivers/messages/${id}/ack`, method: 'POST', data: { id } })
