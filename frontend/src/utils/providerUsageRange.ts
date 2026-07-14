export type ProviderUsagePreset = 'today' | '7d' | '30d' | 'month'

const MAX_RANGE_MS = 30 * 24 * 60 * 60 * 1000

export function providerUsageRange(preset: ProviderUsagePreset, now = new Date()) {
  const end = new Date(now)
  const start = new Date(end)

  if (preset === 'today') {
    start.setHours(0, 0, 0, 0)
  } else if (preset === '7d') {
    start.setTime(end.getTime() - 7 * 24 * 60 * 60 * 1000)
  } else if (preset === 'month') {
    start.setDate(1)
    start.setHours(0, 0, 0, 0)
  } else {
    start.setTime(end.getTime() - MAX_RANGE_MS)
  }

  const earliestAllowed = end.getTime() - MAX_RANGE_MS
  if (start.getTime() < earliestAllowed) start.setTime(earliestAllowed)
  if (start.getTime() >= end.getTime()) start.setTime(end.getTime() - 1)

  return { start_time: start.toISOString(), end_time: end.toISOString() }
}

