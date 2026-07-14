import { describe, expect, it } from 'vitest'
import { sanitizeAuthRedirect } from '../authRedirect'

describe('sanitizeAuthRedirect', () => {
  it('keeps a same-app research group target and its query string', () => {
    expect(sanitizeAuthRedirect('/research-group')).toBe('/research-group')
    expect(sanitizeAuthRedirect('/research-group?invitation=pending')).toBe(
      '/research-group?invitation=pending'
    )
  })

  it.each([
    undefined,
    null,
    '',
    'research-group',
    '//example.com/path',
    'https://example.com/path',
    '/safe\r\nLocation: https://example.com'
  ])('falls back for an unsafe target: %s', (target) => {
    expect(sanitizeAuthRedirect(target)).toBe('/dashboard')
  })

  it('supports an explicit role-aware fallback', () => {
    expect(sanitizeAuthRedirect('//example.com', '/admin/dashboard')).toBe('/admin/dashboard')
  })
})
