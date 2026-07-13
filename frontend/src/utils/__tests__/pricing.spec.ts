import { describe, expect, it } from 'vitest'
import {
  INTERNAL_TOKEN_PRICE_DIVISOR,
  PUBLIC_MODEL_PRICES,
  PUBLIC_RECHARGE_USD_PER_CNY,
  applyInternalTokenPriceRate,
  benchmarkUsdToEffectiveCny,
  formatCompactNumber,
  formatInternalTokenPrice,
  tokenPriceToEffectiveCny,
} from '../pricing'

describe('pricing utilities', () => {
  it('converts original token prices with the internal divisor', () => {
    expect(INTERNAL_TOKEN_PRICE_DIVISOR).toBe(42)
    expect(applyInternalTokenPriceRate(0.000042)).toBeCloseTo(0.000001)
    expect(formatInternalTokenPrice(0.000042, 1_000_000)).toBe('$1')
  })

  it('formats compact decimal numbers without losing integer zeroes', () => {
    expect(formatCompactNumber(6)).toBe('6')
    expect(formatCompactNumber(6.5)).toBe('6.5')
    expect(formatCompactNumber(100)).toBe('100')
  })

  it('converts benchmark model prices to effective CNY after both rates', () => {
    expect(PUBLIC_RECHARGE_USD_PER_CNY).toBe(6)
    expect(benchmarkUsdToEffectiveCny(5)).toBeCloseTo(5 / 252)
    expect(benchmarkUsdToEffectiveCny(30)).toBeCloseTo(30 / 252)
    expect(tokenPriceToEffectiveCny(0.000005, 1_000_000)).toBeCloseTo(5 / 252)
    expect(benchmarkUsdToEffectiveCny(5, 0)).toBeCloseTo(5 / 252)
  })

  it('keeps the public comparison list aligned with the requested model families', () => {
    expect(PUBLIC_MODEL_PRICES.map((model) => [
      model.id,
      model.benchmarkInputUsdPerMillion,
      model.benchmarkOutputUsdPerMillion,
    ])).toEqual([
      ['gpt-5.6-sol', 5, 30],
      ['gpt-5.6-terra', 2.5, 15],
      ['gpt-5.6-luna', 1, 6],
      ['gpt-5.5', 5, 30],
      ['claude-opus-4-8', 5, 25],
      ['claude-sonnet-4-6', 3, 15],
      ['claude-haiku-4-5', 1, 5],
    ])
  })
})
