<template>
  <div class="flex justify-between gap-2">
    <span class="text-gray-500 dark:text-gray-400">{{ label }}</span>
    <span class="text-right font-mono">
      <span class="block">{{ display }}</span>
      <span v-if="cnyDisplay" class="mt-0.5 block text-[10px] font-semibold text-emerald-600 dark:text-emerald-300">
        ≈ {{ cnyDisplay }}
      </span>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  PUBLIC_RECHARGE_USD_PER_CNY,
  formatCompactNumber,
  formatInternalTokenPrice,
  formatScaled,
  tokenPriceToEffectiveCny,
} from '@/utils/pricing'

const props = withDefaults(
  defineProps<{
    label: string
    value: number | null
    unit: string
    scale: number
    internalTokenRate?: boolean
    showCnyEquivalent?: boolean
    usdPerCny?: number
  }>(),
  {
    value: null,
    internalTokenRate: false,
    showCnyEquivalent: false,
    usdPerCny: PUBLIC_RECHARGE_USD_PER_CNY,
  }
)

const display = computed(() => {
  if (props.value == null) return '-'
  const price = props.internalTokenRate
    ? formatInternalTokenPrice(props.value, props.scale)
    : formatScaled(props.value, props.scale)
  return `${price} ${props.unit}`
})

const cnyDisplay = computed(() => {
  if (!props.showCnyEquivalent || !props.internalTokenRate || props.value == null) return ''
  const cny = tokenPriceToEffectiveCny(props.value, props.scale, props.usdPerCny)
  return cny == null ? '' : `¥${formatCompactNumber(cny, 6)} ${props.unit}`
})
</script>
