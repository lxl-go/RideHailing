export function buildAuthenticatedLoginRedirect({ readStorage, tokenKey, homeUrl }) {
  const token = String(readStorage(tokenKey) || '').trim()
  if (!token) {
    return { shouldRedirect: false, url: '' }
  }
  return { shouldRedirect: true, url: homeUrl }
}
