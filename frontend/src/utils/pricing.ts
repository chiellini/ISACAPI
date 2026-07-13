export const INTERNAL_TOKEN_PRICE_DIVISOR = 42
export const PUBLIC_RECHARGE_USD_PER_CNY = 6

export interface PublicModelPrice {
  id: string
  name: string
  family: 'gpt' | 'claude'
  benchmarkInputUsdPerMillion: number
  benchmarkOutputUsdPerMillion: number
}

/**
 * Public comparison prices in benchmark USD per million tokens.
 *
 * Keep these values aligned with the backend model-pricing catalogue. Claude
 * rows intentionally use models that currently have an explicit backend price
 * instead of relying on the unknown-model fallback.
 */
export const PUBLIC_MODEL_PRICES: readonly PublicModelPrice[] = [
  {
    id: 'gpt-5.6-sol',
    name: 'GPT-5.6 Sol',
    family: 'gpt',
    benchmarkInputUsdPerMillion: 5,
    benchmarkOutputUsdPerMillion: 30,
  },
  {
    id: 'gpt-5.6-terra',
    name: 'GPT-5.6 Terra',
    family: 'gpt',
    benchmarkInputUsdPerMillion: 2.5,
    benchmarkOutputUsdPerMillion: 15,
  },
  {
    id: 'gpt-5.6-luna',
    name: 'GPT-5.6 Luna',
    family: 'gpt',
    benchmarkInputUsdPerMillion: 1,
    benchmarkOutputUsdPerMillion: 6,
  },
  {
    id: 'gpt-5.5',
    name: 'GPT-5.5',
    family: 'gpt',
    benchmarkInputUsdPerMillion: 5,
    benchmarkOutputUsdPerMillion: 30,
  },
  {
    id: 'claude-opus-4-8',
    name: 'Claude Opus 4.8',
    family: 'claude',
    benchmarkInputUsdPerMillion: 5,
    benchmarkOutputUsdPerMillion: 25,
  },
  {
    id: 'claude-sonnet-4-6',
    name: 'Claude Sonnet 4.6',
    family: 'claude',
    benchmarkInputUsdPerMillion: 3,
    benchmarkOutputUsdPerMillion: 15,
  },
  {
    id: 'claude-haiku-4-5',
    name: 'Claude Haiku 4.5',
    family: 'claude',
    benchmarkInputUsdPerMillion: 1,
    benchmarkOutputUsdPerMillion: 5,
  },
] as const

/**
 * formatScaled formats a per-token (or per-request) USD price scaled by `scale`.
 *
 *   formatScaled(0.000003, 1_000_000) → "$3"        // per 1M tokens
 *   formatScaled(0.5,        1)        → "$0.5"      // per request
 *   formatScaled(null,       1_000_000) → "-"
 *
 * Uses toPrecision(10) then strips trailing zeros to avoid IEEE 754 display noise.
 */
export function formatScaled(value: number | null, scale: number): string {
  if (value == null) return '-'
  return `$${(value * scale).toPrecision(10).replace(/\.?0+$/, '')}`
}

export function applyInternalTokenPriceRate(value: number | null): number | null {
  if (value == null) return null
  return value / INTERNAL_TOKEN_PRICE_DIVISOR
}

export function formatInternalTokenPrice(value: number | null, scale: number): string {
  return formatScaled(applyInternalTokenPriceRate(value), scale)
}

/**
 * Converts a benchmark USD price into the effective CNY paid on this platform:
 * benchmark USD / internal token divisor / credited USD per CNY.
 */
export function benchmarkUsdToEffectiveCny(
  benchmarkUsd: number,
  usdPerCny = PUBLIC_RECHARGE_USD_PER_CNY,
): number {
  const normalizedRate = Number.isFinite(usdPerCny) && usdPerCny > 0
    ? usdPerCny
    : PUBLIC_RECHARGE_USD_PER_CNY
  return benchmarkUsd / INTERNAL_TOKEN_PRICE_DIVISOR / normalizedRate
}

/** Converts an original per-token USD price into the effective CNY for `scale`. */
export function tokenPriceToEffectiveCny(
  value: number | null,
  scale: number,
  usdPerCny = PUBLIC_RECHARGE_USD_PER_CNY,
): number | null {
  if (value == null) return null
  return benchmarkUsdToEffectiveCny(value * scale, usdPerCny)
}

export function formatCompactNumber(value: number, fractionDigits = 2): string {
  return value
    .toFixed(fractionDigits)
    .replace(/\.?0+$/, '')
}
