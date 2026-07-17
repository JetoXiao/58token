import { describe, expect, it } from 'vitest'

import {
  resolveReadonlyAdminFallbackPath,
  resolveReadonlyAdminRouteRedirect,
} from '../adminMenuPermissions'

describe('read-only admin menu permissions', () => {
  it('does not restrict user-facing routes', () => {
    expect(resolveReadonlyAdminRouteRedirect({
      isReadonlyAdmin: true,
      requiresAdmin: false,
      permissions: [],
    })).toBeUndefined()
  })

  it('allows an authorized admin route', () => {
    expect(resolveReadonlyAdminRouteRedirect({
      isReadonlyAdmin: true,
      requiresAdmin: true,
      adminMenuKey: 'admin_users',
      permissions: ['admin_users'],
    })).toBeUndefined()
  })

  it('redirects an unauthorized admin route to the first authorized admin menu', () => {
    expect(resolveReadonlyAdminRouteRedirect({
      isReadonlyAdmin: true,
      requiresAdmin: true,
      adminMenuKey: 'admin_users',
      permissions: ['redeem', 'admin_usage'],
    })).toBe('/admin/usage')
  })

  it('ignores legacy user-menu permissions when choosing an admin fallback', () => {
    expect(resolveReadonlyAdminFallbackPath(['redeem', 'api_keys'])).toBe('/dashboard')
  })
})
