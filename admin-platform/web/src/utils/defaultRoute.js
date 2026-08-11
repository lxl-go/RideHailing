export const RIDE_HAILING_DASHBOARD_ROUTE = 'RideHailingDashboard'

const DEFAULT_ROUTE_ALIASES = {
  'dashboard': RIDE_HAILING_DASHBOARD_ROUTE,
  'Dashboard': RIDE_HAILING_DASHBOARD_ROUTE
}

export const resolveDefaultRouteName = (routeName, hasRoute) => {
  const resolvedName = DEFAULT_ROUTE_ALIASES[routeName] || routeName || RIDE_HAILING_DASHBOARD_ROUTE
  if (typeof hasRoute !== 'function') return resolvedName
  return hasRoute(resolvedName) ? resolvedName : RIDE_HAILING_DASHBOARD_ROUTE
}

export const normalizeUserInfoDefaultRoute = (userInfo = {}) => {
  if (!userInfo?.authority) return userInfo
  return {
    ...userInfo,
    authority: {
      ...userInfo.authority,
      defaultRouter: resolveDefaultRouteName(userInfo.authority.defaultRouter)
    }
  }
}
