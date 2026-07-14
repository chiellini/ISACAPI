<template>
  <AppLayout>
    <div class="space-y-5">
      <div>
        <RouterLink to="/admin/users" class="text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">
          ← {{ t('provider.admin.backToUsers') }}
        </RouterLink>
        <h1 class="mt-3 text-2xl font-semibold text-gray-900 dark:text-white">
          {{ t('provider.admin.providerUsageTitle') }}
          <span v-if="providerLabel" class="text-base font-normal text-gray-500">· {{ providerLabel }}</span>
        </h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('provider.admin.providerUsageDescription') }}</p>
      </div>
      <ProviderUsagePanel v-if="providerId > 0" :provider-id="providerId" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProviderUsagePanel from '@/components/provider/ProviderUsagePanel.vue'
import { adminAPI } from '@/api/admin'

const route = useRoute()
const { t } = useI18n()
const providerId = computed(() => Number(route.params.id) || 0)
const providerLabel = ref('')

onMounted(async () => {
  if (providerId.value <= 0) return
  try {
    const provider = await adminAPI.users.getById(providerId.value)
    providerLabel.value = provider.username || provider.email || `#${provider.id}`
  } catch {
    providerLabel.value = `#${providerId.value}`
  }
})
</script>

