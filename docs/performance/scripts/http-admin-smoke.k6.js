import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  scenarios: {
    admin_http_smoke: {
      executor: 'constant-vus',
      vus: Number(__ENV.VUS || 20),
      duration: __ENV.DURATION || '2m',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.001'],
    http_req_duration: ['p(99)<200'],
  },
}

const baseUrl = __ENV.BASE_URL || 'http://localhost:8888'
const token = __ENV.TOKEN || ''

function headers() {
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export default function () {
  const responses = http.batch([
    ['GET', `${baseUrl}/carpool/performance/summary`, null, { headers: headers() }],
    ['GET', `${baseUrl}/carpool/performance/scenario/list?enabled=true`, null, { headers: headers() }],
    ['GET', `${baseUrl}/carpool/performance/runtime`, null, { headers: headers() }],
  ])

  responses.forEach((res) => {
    check(res, {
      'status is 2xx': (r) => r.status >= 200 && r.status < 300,
      'body has data': (r) => r.body && r.body.length > 0,
    })
  })

  sleep(1)
}
