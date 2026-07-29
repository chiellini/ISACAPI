<template>
  <div
    v-if="brands.length"
    data-test="payment-brand-support"
    :class="[
      'flex flex-wrap items-center gap-1.5',
      align === 'center' ? 'justify-center' : 'justify-start',
    ]"
  >
    <span
      v-for="brand in brands"
      :key="brand.label"
      class="inline-flex h-6 items-center gap-1 rounded-md border border-blue-200 bg-blue-50 px-1.5 text-[11px] font-medium text-gray-700 dark:border-blue-800 dark:bg-blue-950/40 dark:text-gray-200"
    >
      <img :src="brand.icon" :alt="brand.label" class="h-3.5 w-3.5 object-contain" />
      {{ brand.label }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { isBuiltInAlipayMethod } from './providerConfig'
import alipayIcon from '@/assets/icons/alipay.svg'

const ALIPAY_HK_LOGO = 'https://www.alipayhk.com/wp-content/uploads/2025/11/AlipayHK_logo.png'

const props = withDefaults(defineProps<{
  method: string
  includePrimary?: boolean
  align?: 'start' | 'center'
}>(), {
  includePrimary: false,
  align: 'start',
})

const brands = computed(() => {
  if (!isBuiltInAlipayMethod(props.method)) return []

  const supported = [{ label: 'AlipayHK', icon: ALIPAY_HK_LOGO }]
  return props.includePrimary
    ? [{ label: 'Alipay', icon: alipayIcon }, ...supported]
    : supported
})
</script>
