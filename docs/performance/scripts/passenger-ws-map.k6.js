import ws from 'k6/ws'
import { check, sleep } from 'k6'

export const options = {
  scenarios: {
    passenger_ws_map: {
      executor: 'constant-vus',
      vus: Number(__ENV.VUS || 50),
      duration: __ENV.DURATION || '3m',
    },
  },
  thresholds: {
    checks: ['rate>0.98'],
  },
}

const wsUrl = __ENV.WS_URL || 'ws://localhost:8888/ws/passenger/tracking'
const token = __ENV.TOKEN || ''

export default function () {
  const url = token ? `${wsUrl}?token=${encodeURIComponent(token)}` : wsUrl
  const response = ws.connect(url, {}, (socket) => {
    socket.on('open', () => {
      socket.send(JSON.stringify({ type: 'subscribe_trip', tripId: __ENV.TRIP_ID || 'WO08-DEMO-TRIP' }))
    })

    socket.on('message', (message) => {
      check(message, {
        'message is not empty': (value) => value && value.length > 0,
      })
    })

    socket.setTimeout(() => socket.close(), 10000)
  })

  check(response, {
    'websocket connected': (res) => res && res.status === 101,
  })
  sleep(1)
}
