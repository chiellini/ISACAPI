import { describe, expect, it } from 'vitest'
import {
  INTERNAL_TOKEN_PRICE_DIVISOR,
  applyInternalTokenPriceRate,
  formatCompactNumber,
  formatInternalTokenPrice,
} from '../pricing'

describe('pricing utilities', () => {
  it('converts official token prices with the internal divisor', () => {
    expect(INTERNAL_TOKEN_PRICE_DIVISOR).toBe(42)
    expect(applyInternalTokenPriceRate(0.000042)).toBeCloseTo(0.000001)
    expect(formatInternalTokenPrice(0.000042, 1_000_000)).toBe('$1')
  })

  it('formats compact decimal numbers without losing integer zeroes', () => {
    expect(formatCompactNumber(6)).toBe('6')
    expect(formatCompactNumber(6.5)).toBe('6.5')
    expect(formatCompactNumber(100)).toBe('100')
  })
})
