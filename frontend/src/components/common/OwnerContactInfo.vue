<template>
  <div :class="containerClass">
    <div :class="titleClass">
      <Icon name="chat" size="sm" class="text-primary-500" />
      <span>{{ displayTitle }}</span>
    </div>

    <div :class="listClass">
      <a
        href="https://wa.me/85294893166"
        target="_blank"
        rel="noopener noreferrer"
        class="transition-colors hover:text-primary-600 dark:hover:text-primary-400"
      >
        Whatsapp: +852 94893166
      </a>
      <span>Wechat/Weixin: 18319277690</span>
      <div class="min-w-0 break-all">
        <span>Email: </span>
        <a
          href="mailto:chiellini.lee@gmail.com"
          class="transition-colors hover:text-primary-600 dark:hover:text-primary-400"
        >
          chiellini.lee@gmail.com
        </a>
        <span class="mx-1 text-gray-400 dark:text-dark-500">/</span>
        <a
          href="mailto:18319277690@163.com"
          class="transition-colors hover:text-primary-600 dark:hover:text-primary-400"
        >
          18319277690@163.com
        </a>
      </div>
      <span>Telegram: +852 94893166</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(
  defineProps<{
    variant?: 'card' | 'plain'
    columns?: boolean
    title?: string
  }>(),
  {
    variant: 'card',
    columns: true,
    title: ''
  }
)

const { t } = useI18n()

const displayTitle = computed(() => props.title || t('common.contactAdmin'))

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

const listClass = computed(() => {
  const colorClass = 'text-gray-600 dark:text-dark-300'
  return props.columns
    ? `grid gap-2 ${colorClass} sm:grid-cols-2`
    : `space-y-2 ${colorClass}`
})
</script>
