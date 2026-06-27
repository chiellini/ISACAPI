<template>
  <div class="min-h-screen bg-gray-50 text-gray-900 dark:bg-dark-950 dark:text-white">
    <!-- Header -->
    <header class="border-b border-gray-200 bg-white/95 dark:border-dark-800 dark:bg-dark-900/95">
      <div class="mx-auto flex max-w-5xl items-center justify-between gap-4 px-4 py-4 sm:px-6">
        <RouterLink to="/home" class="flex min-w-0 items-center gap-3">
          <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-gray-200 dark:bg-dark-800 dark:ring-dark-700">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </span>
          <span class="truncate text-base font-semibold text-gray-950 dark:text-white">{{ siteName }}</span>
        </RouterLink>
        <RouterLink
          to="/login"
          class="inline-flex flex-shrink-0 items-center justify-center rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm shadow-primary-600/20 transition hover:bg-primary-700"
        >
          {{ t('home.login') }}
        </RouterLink>
      </div>
    </header>

    <main class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:py-10">
      <!-- Overall banner -->
      <section
        class="flex flex-col gap-3 rounded-2xl border p-5 sm:flex-row sm:items-center sm:justify-between"
        :class="bannerClass"
      >
        <div class="flex items-center gap-3">
          <span class="relative flex h-3.5 w-3.5">
            <span
              v-if="overallStatus === 'operational'"
              class="absolute inline-flex h-full w-full animate-ping rounded-full opacity-60"
              :class="dotClass(overallStatus)"
            ></span>
            <span class="relative inline-flex h-3.5 w-3.5 rounded-full" :class="dotClass(overallStatus)"></span>
          </span>
          <div>
            <h1 class="text-lg font-semibold">{{ t('publicStatus.title') }}</h1>
            <p class="text-sm opacity-90">{{ overallText }}</p>
          </div>
        </div>
        <div class="flex items-center gap-3 text-xs">
          <span v-if="lastUpdated" class="opacity-80">{{ t('publicStatus.updatedAt', { time: lastUpdated }) }}</span>
          <button
            type="button"
            @click="reload(false)"
            :disabled="loading"
            class="inline-flex items-center gap-1.5 rounded-lg border border-current/20 bg-white/60 px-3 py-1.5 font-medium transition hover:bg-white dark:bg-white/10 dark:hover:bg-white/20"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            {{ t('common.refresh', 'Refresh') }}
          </button>
        </div>
      </section>

      <!-- Loading -->
      <div v-if="loading && providers.length === 0" class="flex min-h-[240px] items-center justify-center">
        <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
      </div>

      <!-- Error -->
      <section
        v-else-if="loadError"
        class="mt-6 rounded-xl border border-red-200 bg-red-50 p-6 text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200"
      >
        <p class="text-sm">{{ loadError }}</p>
      </section>

      <!-- Disabled / empty -->
      <section
        v-else-if="!enabled || providers.length === 0"
        class="mt-6 rounded-xl border border-gray-200 bg-white p-8 text-center dark:border-dark-700 dark:bg-dark-900"
      >
        <p class="text-sm text-gray-500 dark:text-dark-300">{{ t('publicStatus.empty') }}</p>
      </section>

      <!-- Providers -->
      <section v-else class="mt-6 space-y-4">
        <article
          v-for="provider in providers"
          :key="provider.provider"
          class="overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900"
        >
          <header class="flex items-center justify-between gap-3 border-b border-gray-100 px-5 py-3.5 dark:border-dark-800">
            <div class="flex items-center gap-3">
              <span class="inline-flex h-2.5 w-2.5 rounded-full" :class="dotClass(provider.status)"></span>
              <h2 class="font-semibold">{{ providerLabel(provider.provider) }}</h2>
              <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="badgeClass(provider.status)">
                {{ statusLabel(provider.status) }}
              </span>
            </div>
            <span v-if="provider.availability_7d !== null" class="text-sm font-medium tabular-nums text-gray-500 dark:text-dark-300">
              {{ formatPct(provider.availability_7d) }}
            </span>
          </header>

          <ul class="divide-y divide-gray-100 dark:divide-dark-800">
            <li
              v-for="model in provider.models"
              :key="model.model"
              class="flex items-center justify-between gap-3 px-5 py-2.5"
            >
              <div class="flex min-w-0 items-center gap-2.5">
                <span class="inline-flex h-2 w-2 flex-shrink-0 rounded-full" :class="dotClass(model.status)"></span>
                <span class="truncate font-mono text-sm text-gray-700 dark:text-dark-200">{{ model.model }}</span>
              </div>
              <div class="flex flex-shrink-0 items-center gap-3">
                <span
                  v-if="model.availability_7d !== null"
                  class="text-xs tabular-nums text-gray-400 dark:text-dark-400"
                >{{ formatPct(model.availability_7d) }}</span>
                <span class="text-xs font-medium" :class="textClass(model.status)">{{ statusLabel(model.status) }}</span>
              </div>
            </li>
          </ul>
        </article>

        <p class="pt-1 text-center text-xs text-gray-400 dark:text-dark-500">
          {{ t('publicStatus.windowHint') }}
        </p>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import Icon from '@/components/icons/Icon.vue'
import {
  getPublicStatus,
  type PublicStatusProvider,
  type PublicStatusValue,
} from '@/api/publicStatus'

const { t } = useI18n()
const appStore = useAppStore()
const { siteName, siteLogo } = storeToRefs(appStore)

const providers = ref<PublicStatusProvider[]>([])
const overallStatus = ref<PublicStatusValue>('operational')
const enabled = ref(true)
const loading = ref(false)
const loadError = ref('')
const lastUpdated = ref('')

let abortController: AbortController | null = null

const autoRefresh = useAutoRefresh({
  storageKey: 'public-status-auto-refresh',
  intervals: [60, 120, 300] as const,
  defaultInterval: 60,
  onRefresh: () => reload(true),
  shouldPause: () => document.hidden || loading.value,
})

const PROVIDER_LABELS: Record<string, string> = {
  anthropic: 'Claude',
  openai: 'OpenAI',
  gemini: 'Gemini',
  antigravity: 'Antigravity',
}

function providerLabel(provider: string): string {
  return PROVIDER_LABELS[provider] || provider.charAt(0).toUpperCase() + provider.slice(1)
}

function statusLabel(status: PublicStatusValue): string {
  return t(`publicStatus.status.${status}`)
}

function dotClass(status: PublicStatusValue): string {
  switch (status) {
    case 'operational':
      return 'bg-emerald-500'
    case 'degraded':
      return 'bg-amber-500'
    default:
      return 'bg-red-500'
  }
}

function textClass(status: PublicStatusValue): string {
  switch (status) {
    case 'operational':
      return 'text-emerald-600 dark:text-emerald-400'
    case 'degraded':
      return 'text-amber-600 dark:text-amber-400'
    default:
      return 'text-red-600 dark:text-red-400'
  }
}

function badgeClass(status: PublicStatusValue): string {
  switch (status) {
    case 'operational':
      return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
    case 'degraded':
      return 'bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300'
    default:
      return 'bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-300'
  }
}

const bannerClass = computed(() => {
  switch (overallStatus.value) {
    case 'operational':
      return 'border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-500/30 dark:bg-emerald-500/10 dark:text-emerald-200'
    case 'degraded':
      return 'border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-200'
    default:
      return 'border-red-200 bg-red-50 text-red-800 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200'
  }
})

const overallText = computed(() => {
  if (!enabled.value || providers.value.length === 0) return t('publicStatus.empty')
  return t(`publicStatus.overall.${overallStatus.value}`)
})

function formatPct(value: number): string {
  // Backend availability is a 0..1 ratio.
  return `${(value * 100).toFixed(2)}%`
}

async function reload(silent = false) {
  if (abortController) abortController.abort()
  const ctrl = new AbortController()
  abortController = ctrl
  if (!silent) loading.value = true
  loadError.value = ''
  try {
    const res = await getPublicStatus({ signal: ctrl.signal })
    if (ctrl.signal.aborted || abortController !== ctrl) return
    enabled.value = res.enabled
    providers.value = res.providers || []
    overallStatus.value = res.overall_status || 'operational'
    lastUpdated.value = new Date().toLocaleTimeString()
  } catch (err: unknown) {
    const e = err as { name?: string; code?: string }
    if (e?.name === 'AbortError' || e?.code === 'ERR_CANCELED') return
    loadError.value = extractApiErrorMessage(err, t('publicStatus.loadError'))
  } finally {
    if (abortController === ctrl) {
      if (!silent) loading.value = false
      abortController = null
    }
  }
}

onMounted(() => {
  void appStore.fetchPublicSettings()
  void reload(false)
  autoRefresh.setEnabled(true)
})

onBeforeUnmount(() => {
  if (abortController) abortController.abort()
  autoRefresh.stop()
})
</script>
