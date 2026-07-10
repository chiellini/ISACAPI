export const INTERNAL_TOKEN_PRICE_DIVISOR = 42
export const PUBLIC_RECHARGE_USD_PER_CNY = 6

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

export function formatCompactNumber(value: number, fractionDigits = 2): string {
  return value
    .toFixed(fractionDigits)
    .replace(/\.?0+$/, '')
}
