<template>
  <div :class="containerClass">
    <div :class="titleClass">
      <Icon name="chat" size="sm" class="text-primary-500" />
      <span>{{ t('common.contactAdmin') }}</span>
    </div>

    <p class="text-gray-600 dark:text-dark-300">
      如需帮助，请加入 QQ 群：
      <span class="font-semibold text-gray-900 dark:text-white">1027890648</span>
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(
  defineProps<{
    variant?: 'card' | 'plain'
  }>(),
  {
    variant: 'card'
  }
)

const { t } = useI18n()

const containerClass = computed(() =>
  props.variant === 'card'
    ? 'rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-dark-700 dark:bg-dark-800/60'
    : 'text-sm'
)

const titleClass = computed(() =>
  props.variant === 'card'
    ? 'mb-3 flex items-center gap-2 font-medium text-gray-900 dark:text-white'
    : 'mb-2 flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-dark-400'
)
</script>
