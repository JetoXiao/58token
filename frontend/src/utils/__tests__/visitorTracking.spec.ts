import { describe, expect, it, vi, beforeEach } from 'vitest'
import { normalizeAttributionCode, resolveAttribution, sanitizeTrackedURL, shouldTrackPath } from '../visitorTracking'

describe('visitorTracking', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.useRealTimers()
  })

  it('normalizes safe channel codes and rejects arbitrary values', () => {
    expect(normalizeAttributionCode(' Discord_Main ')).toBe('discord_main')
    expect(normalizeAttributionCode('https://example.com')).toBe('')
  })

  it('persists first-party attribution across navigation', () => {
    expect(resolveAttribution('?ref=reddit')).toBe('reddit')
    expect(resolveAttribution('')).toBe('reddit')
  })

  it('does not collect administrator or setup page views', () => {
    expect(shouldTrackPath('/home')).toBe(true)
    expect(shouldTrackPath('/login')).toBe(true)
    expect(shouldTrackPath('/admin/dashboard')).toBe(false)
    expect(shouldTrackPath('/setup')).toBe(false)
  })

  it('drops sensitive query parameters while preserving attribution fields', () => {
    const sanitized = new URL(sanitizeTrackedURL('https://example.com/payment/result?resume_token=secret&ref=reddit&utm_campaign=launch'))
    expect(sanitized.searchParams.get('resume_token')).toBeNull()
    expect(sanitized.searchParams.get('ref')).toBe('reddit')
    expect(sanitized.searchParams.get('utm_campaign')).toBe('launch')
  })
})
