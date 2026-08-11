export const mockDriverTrips = Array.from({ length: 10 }, (_, i) => ({
  id: i + 1, origin: '上海静安', destination: '上海虹桥',
  depart_time: `2026-07-${String(28 + i).padStart(2, '0')} 08:00`,
  seats_total: 4, seats_available: 3, price: 32, status: i < 3 ? 1 : 0
}))

export const mockPendingOrders = Array.from({ length: 8 }, (_, i) => ({
  id: i + 1, passenger_id: 1000 + i, seats_booked: (i % 3) + 1,
  total_price: 32 * ((i % 3) + 1), status: 0,
  created_at: `2026-07-28 0${8 + i}:00`
}))
