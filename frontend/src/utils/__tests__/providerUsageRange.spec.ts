import { describe, expect, it } from 'vitest'
import { providerUsageRange } from '@/utils/providerUsageRange'

describe('provider usage ranges', () => {
  it('uses exactly one shared end instant for the default 30-day range', () => {
    const now = new Date('2026-07-13T12:34:56.789Z')
    const result = providerUsageRange('30d', now)

    expect(Date.parse(result.end_time)).toBe(now.getTime())
    expect(Date.parse(result.end_time) - Date.parse(result.start_time)).toBe(30 * 24 * 60 * 60 * 1000)
  })

  it('clamps this-month on the 31st to the backend 30-day hard limit', () => {
    const result = providerUsageRange('month', new Date('2026-07-31T23:00:00.000Z'))
    const duration = Date.parse(result.end_time) - Date.parse(result.start_time)

    expect(duration).toBeGreaterThan(0)
    expect(duration).toBeLessThanOrEqual(30 * 24 * 60 * 60 * 1000)
  })
})

