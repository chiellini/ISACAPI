<template>
  <AppLayout>
    <div class="space-y-5">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('provider.usage.title') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('provider.usage.description') }}</p>
      </div>
      <ProviderUsagePanel />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProviderUsagePanel from '@/components/provider/ProviderUsagePanel.vue'

const { t } = useI18n()
</script>

