import type { Router } from 'vue-router'

const ATTRIBUTION_KEY = 'useaiforme_visitor_attribution'
const VISITOR_ID_KEY = 'useaiforme_visitor_id'
const SESSION_ID_KEY = 'useaiforme_session_id'
const ATTRIBUTION_TTL = 30 * 24 * 60 * 60 * 1000

interface Attribution {
  code: string
  expiresAt: number
}
function randomID(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
}

function getStorageID(storage: Storage, key: string): string {
  const existing = storage.getItem(key)
  if (existing) return existing
  const value = randomID()
  storage.setItem(key, value)
  return value
}

export function normalizeAttributionCode(value: string | null): string {
  const normalized = (value || '').trim().toLowerCase()
  return /^[a-z0-9][a-z0-9_-]{0,63}$/.test(normalized) ? normalized : ''
}

export function resolveAttribution(search: string, storage: Storage = localStorage): string {
  const params = new URLSearchParams(search)
  const incoming = normalizeAttributionCode(
    params.get('ref') || params.get('channel') || params.get('utm_source'),
  )
  if (incoming) {
    storage.setItem(ATTRIBUTION_KEY, JSON.stringify({ code: incoming, expiresAt: Date.now() + ATTRIBUTION_TTL }))
    return incoming
  }
  try {
    const saved = JSON.parse(storage.getItem(ATTRIBUTION_KEY) || '') as Attribution
    if (saved.expiresAt > Date.now()) return normalizeAttributionCode(saved.code)
  } catch {
    // Ignore malformed or expired first-party attribution state.
  }
  storage.removeItem(ATTRIBUTION_KEY)
  return 'direct'
}

export function shouldTrackPath(path: string): boolean {
  return !path.startsWith('/admin') && !path.startsWith('/setup')
}

export function sanitizeTrackedURL(rawURL: string): string {
  try {
    const url = new URL(rawURL, window.location.origin)
    const safe = new URL(url.pathname, url.origin)
    for (const key of ['ref', 'channel', 'utm_source', 'utm_medium', 'utm_campaign']) {
      const value = url.searchParams.get(key)
      if (value) safe.searchParams.set(key, value.slice(0, 128))
    }
    return safe.toString()
  } catch {
    return ''
  }
}

export function installVisitorTracking(router: Router): () => void {
  let lastSignature = ''
  let lastSentAt = 0

  const send = (path: string, fullPath: string) => {
    if (!shouldTrackPath(path)) return
    const now = Date.now()
    const signature = fullPath
    if (signature === lastSignature && now - lastSentAt < 1000) return
    lastSignature = signature
    lastSentAt = now

    const payload = {
      channel_code: resolveAttribution(window.location.search),
      visitor_id: getStorageID(localStorage, VISITOR_ID_KEY),
      session_id: getStorageID(sessionStorage, SESSION_ID_KEY),
      path,
      referrer: sanitizeTrackedURL(document.referrer),
      landing_url: sanitizeTrackedURL(window.location.href),
      language: navigator.language,
      screen: `${window.screen.width}x${window.screen.height}`,
    }

    const headers: Record<string, string> = { 'Content-Type': 'application/json' }
    const authToken = localStorage.getItem('auth_token')
    if (authToken) headers.Authorization = `Bearer ${authToken}`

    void fetch('/api/v1/analytics/visit', {
      method: 'POST',
      headers,
      body: JSON.stringify(payload),
      credentials: 'same-origin',
      keepalive: true,
    }).catch(() => undefined)
  }

  const current = router.currentRoute.value
  send(current.path, current.fullPath)
  return router.afterEach((to) => send(to.path, to.fullPath))
}
