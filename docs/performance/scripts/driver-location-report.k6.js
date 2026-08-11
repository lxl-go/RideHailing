import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  scenarios: {
    driver_location_report: {
      executor: 'constant-arrival-rate',
      rate: Number(__ENV.RATE || 200),
      timeUnit: '1s',
      duration: __ENV.DURATION || '3m',
      preAllocatedVUs: Number(__ENV.VUS || 80),
      maxVUs: Number(__ENV.MAX_VUS || 300),
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.0001'],
    http_req_duration: ['p(99)<100'],
  },
}

const baseUrl = __ENV.BASE_URL || 'http://localhost:8888'
const token = __ENV.TOKEN || ''

function headers() {
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  }
}

export default function () {
  const payload = JSON.stringify({
    driverId: __ENV.DRIVER_ID || 'WO08-DEMO-DRIVER',
    orderId: __ENV.ORDER_ID || 'WO08-DEMO-ORDER',
    longitude: 116.397 + Math.random() / 1000,
    latitude: 39.908 + Math.random() / 1000,
    speed: 32 + Math.random() * 8,
    reportedAt: new Date().toISOString(),
  })

  const res = http.post(`${baseUrl}/driver/location/report`, payload, { headers: headers() })
  check(res, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
  })
  sleep(0.5)
}
