<template>
  <div class="flex justify-between gap-2">
    <span class="text-gray-500 dark:text-gray-400">{{ label }}</span>
    <span class="font-mono">{{ display }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatInternalTokenPrice, formatScaled } from '@/utils/pricing'

const props = withDefaults(
  defineProps<{
    label: string
    value: number | null
    unit: string
    scale: number
    internalTokenRate?: boolean
  }>(),
  { value: null, internalTokenRate: false }
)

const display = computed(() => {
  if (props.value == null) return '-'
  const price = props.internalTokenRate
    ? formatInternalTokenPrice(props.value, props.scale)
    : formatScaled(props.value, props.scale)
  return `${price} ${props.unit}`
})
</script>
