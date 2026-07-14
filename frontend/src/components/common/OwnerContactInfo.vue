<template>
  <QqGroupContact :variant="variant" />
</template>

<script setup lang="ts">
import QqGroupContact from '@/components/common/QqGroupContact.vue'

withDefaults(
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
</script>
