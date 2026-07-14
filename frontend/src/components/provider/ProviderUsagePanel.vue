<template>
  <div class="space-y-5">
    <div class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-800/60 dark:bg-amber-950/30 dark:text-amber-200">
      {{ t('provider.usage.platformOnlyNotice') }}
    </div>

    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap gap-2">
        <button
          v-for="preset in presets"
          :key="preset.value"
          type="button"
          :class="['btn btn-sm', selectedPreset === preset.value ? 'btn-primary' : 'btn-secondary']"
          @click="selectPreset(preset.value)"
        >
          {{ preset.label }}
        </button>
      </div>
      <button type="button" class="btn btn-secondary" :disabled="loading" @click="load">
        <Icon name="refresh" size="sm" class="mr-1.5" :class="loading ? 'animate-spin' : ''" />
        {{ t('provider.usage.refresh') }}
      </button>
    </div>

    <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
      <div v-for="item in summary" :key="item.key" class="card p-4">
        <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ item.label }}</div>
        <div class="mt-2 font-mono text-2xl font-semibold text-gray-900 dark:text-white">{{ formatCount(item.value) }}</div>
      </div>
    </div>

    <div class="card overflow-hidden">
      <div class="border-b border-gray-200 px-4 py-3 dark:border-dark-700">
        <h2 class="font-semibold text-gray-900 dark:text-white">{{ t('provider.usage.byAccount') }}</h2>
      </div>
      <DataTable :columns="columns" :data="usage.accounts" :loading="loading" row-key="account_id">
        <template #cell-account_name="{ row }">
          <div>
            <div class="font-medium text-gray-900 dark:text-white">{{ row.account_name || `#${row.account_id}` }}</div>
            <div class="font-mono text-xs text-gray-400">#{{ row.account_id }}</div>
          </div>
        </template>
        <template #cell-platform="{ value }">
          <span class="capitalize">{{ value || '-' }}</span>
        </template>
        <template v-for="field in tokenFields" :key="field" #[`cell-${field}`]="{ value }">
          <span class="font-mono">{{ formatCount(value) }}</span>
        </template>
        <template #empty>
          <div class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('provider.usage.empty') }}</div>
        </template>
      </DataTable>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import DataTable from '@/components/common/DataTable.vue'
import Icon from '@/components/icons/Icon.vue'
import { providerAPI, type ProviderUsageResponse } from '@/api/provider'
import { useAppStore } from '@/stores/app'
import { providerUsageRange, type ProviderUsagePreset } from '@/utils/providerUsageRange'
import type { Column } from '@/components/common/types'

const props = defineProps<{ providerId?: number }>()
const { t, locale } = useI18n()
const appStore = useAppStore()
const selectedPreset = ref<ProviderUsagePreset>('30d')
const loading = ref(false)
const usage = ref<ProviderUsageResponse>({
  provider_id: props.providerId ?? 0,
  start_time: '',
  end_time: '',
  totals: {
    total_requests: 0,
    total_input_tokens: 0,
    total_output_tokens: 0,
    total_cache_creation_tokens: 0,
    total_cache_read_tokens: 0,
    total_tokens: 0,
  },
  accounts: [],
})

const presets = computed(() => [
  { value: 'today' as const, label: t('provider.usage.today') },
  { value: '7d' as const, label: t('provider.usage.last7Days') },
  { value: '30d' as const, label: t('provider.usage.last30Days') },
  { value: 'month' as const, label: t('provider.usage.thisMonth') },
])

const summary = computed(() => [
  { key: 'total', label: t('provider.usage.totalTokens'), value: usage.value.totals.total_tokens },
  { key: 'input', label: t('provider.usage.inputTokens'), value: usage.value.totals.total_input_tokens },
  { key: 'output', label: t('provider.usage.outputTokens'), value: usage.value.totals.total_output_tokens },
  { key: 'cache-create', label: t('provider.usage.cacheCreationTokens'), value: usage.value.totals.total_cache_creation_tokens },
  { key: 'cache-read', label: t('provider.usage.cacheReadTokens'), value: usage.value.totals.total_cache_read_tokens },
  { key: 'requests', label: t('provider.usage.requests'), value: usage.value.totals.total_requests },
])

const columns = computed<Column[]>(() => [
  { key: 'account_name', label: t('provider.usage.account') },
  { key: 'platform', label: t('provider.usage.platform') },
  { key: 'total_requests', label: t('provider.usage.requests') },
  { key: 'total_input_tokens', label: t('provider.usage.inputTokens') },
  { key: 'total_output_tokens', label: t('provider.usage.outputTokens') },
  { key: 'total_cache_creation_tokens', label: t('provider.usage.cacheCreationTokens') },
  { key: 'total_cache_read_tokens', label: t('provider.usage.cacheReadTokens') },
  { key: 'total_tokens', label: t('provider.usage.totalTokens') },
])

const tokenFields = [
  'total_requests',
  'total_input_tokens',
  'total_output_tokens',
  'total_cache_creation_tokens',
  'total_cache_read_tokens',
  'total_tokens',
]

const load = async () => {
  loading.value = true
  try {
    usage.value = props.providerId === undefined
      ? await providerAPI.getUsage(providerUsageRange(selectedPreset.value))
      : await providerAPI.getAdminUsage(props.providerId, providerUsageRange(selectedPreset.value))
  } catch {
    appStore.showError(t('provider.usage.loadFailed'))
  } finally {
    loading.value = false
  }
}

const selectPreset = (preset: ProviderUsagePreset) => {
  selectedPreset.value = preset
  void load()
}

const formatCount = (value: number) => new Intl.NumberFormat(locale.value).format(value ?? 0)

onMounted(() => { void load() })
watch(() => props.providerId, (next, previous) => {
  if (next !== previous) void load()
})
</script>
