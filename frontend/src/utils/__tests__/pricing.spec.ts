import { describe, expect, it } from 'vitest'
import {
  PUBLIC_GROUP_RATE_MULTIPLIER,
  PUBLIC_MODEL_PRICES,
  PUBLIC_RECHARGE_USD_PER_CNY,
  PUBLIC_RECHARGE_VALUE_MULTIPLIER,
  applyInternalTokenPriceRate,
  benchmarkUsdToEffectiveCny,
  formatCompactNumber,
  formatInternalTokenPrice,
  normalizeRechargeUsdPerCny,
  tokenPriceToEffectiveCny,
} from '../pricing'

describe('pricing utilities', () => {
  it('keeps balance token prices at the configured x1 group rate', () => {
    expect(PUBLIC_GROUP_RATE_MULTIPLIER).toBe(1)
    expect(applyInternalTokenPriceRate(0.000005)).toBeCloseTo(0.000005)
    expect(formatInternalTokenPrice(0.000005, 1_000_000)).toBe('$5')
  })

  it('formats compact decimal numbers without losing integer zeroes', () => {
    expect(formatCompactNumber(6)).toBe('6')
    expect(formatCompactNumber(6.5)).toBe('6.5')
    expect(formatCompactNumber(100)).toBe('100')
  })

  it('converts benchmark balance prices to effective CNY through the recharge rate', () => {
    expect(PUBLIC_RECHARGE_USD_PER_CNY).toBe(6)
    expect(PUBLIC_RECHARGE_VALUE_MULTIPLIER).toBe(42)
    expect(benchmarkUsdToEffectiveCny(5)).toBeCloseTo(5 / 6)
    expect(benchmarkUsdToEffectiveCny(30)).toBeCloseTo(30 / 6)
    expect(tokenPriceToEffectiveCny(0.000005, 1_000_000)).toBeCloseTo(5 / 6)
    expect(benchmarkUsdToEffectiveCny(5, 0)).toBeCloseTo(5 / 6)
  })

  it('uses a configured recharge rate and normalizes invalid public values', () => {
    expect(normalizeRechargeUsdPerCny(5)).toBe(5)
    expect(benchmarkUsdToEffectiveCny(5, 5)).toBe(1)
    expect(tokenPriceToEffectiveCny(0.000005, 1_000_000, 5)).toBe(1)
    expect(normalizeRechargeUsdPerCny(undefined)).toBe(PUBLIC_RECHARGE_USD_PER_CNY)
    expect(normalizeRechargeUsdPerCny(0)).toBe(PUBLIC_RECHARGE_USD_PER_CNY)
    expect(normalizeRechargeUsdPerCny(Number.NaN)).toBe(PUBLIC_RECHARGE_USD_PER_CNY)
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
