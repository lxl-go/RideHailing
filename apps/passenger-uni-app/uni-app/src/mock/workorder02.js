export const mockShuttleRoutes = Array.from({ length: 8 }, (_, i) => ({
  id: i + 1, name: `班车线路${i + 1}`,
  start: ['静安寺', '人民广场', '徐家汇', '陆家嘴'][i % 4],
  end: ['虹桥枢纽', '浦东机场', '上海南站', '上海站'][i % 4],
  departTime: `${6 + i}:00`, price: 15 + i * 5, seats: 40 - i * 5
}))
