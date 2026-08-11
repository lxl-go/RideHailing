export function createAutoRefresh(task, options = {}) {
  const intervalMs = options.intervalMs ?? 5000
  const setIntervalFn = options.setIntervalFn ?? setInterval
  const clearIntervalFn = options.clearIntervalFn ?? clearInterval
  const runImmediately = options.runImmediately ?? true
  let timer = null
  let running = false

  const tick = () => {
    if (running) return Promise.resolve()
    running = true
    let result
    try {
      result = task()
    } catch (error) {
      running = false
      throw error
    }
    return Promise.resolve(result).finally(() => {
      running = false
    })
  }

  const stop = () => {
    if (timer === null) return
    clearIntervalFn(timer)
    timer = null
  }

  const start = () => {
    stop()
    if (runImmediately) void tick()
    timer = setIntervalFn(() => {
      void tick()
    }, intervalMs)
  }

  return { start, stop }
}
