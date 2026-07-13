<template>
  <div
    :dir="isRtl ? 'rtl' : 'ltr'"
    class="min-h-screen bg-slate-50 text-slate-950 dark:bg-dark-950 dark:text-white"
  >
    <header
      class="sticky top-0 z-40 border-b border-slate-200/80 bg-white/90 px-4 py-3 backdrop-blur-xl dark:border-dark-800 dark:bg-dark-950/90"
    >
      <nav class="mx-auto flex max-w-7xl items-center justify-between gap-4">
        <router-link to="/home" class="flex min-w-0 items-center gap-3">
          <span
            class="h-10 w-10 shrink-0 overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm dark:border-dark-700"
          >
            <img src="/logo.png" alt="ISACAI" class="h-full w-full object-contain" />
          </span>
          <span class="min-w-0">
            <span class="block truncate text-sm font-bold text-slate-950 dark:text-white">
              {{ siteName }}
            </span>
            <span class="hidden text-xs text-slate-500 dark:text-dark-400 sm:block">
              {{ t('home.nav.tagline') }}
            </span>
          </span>
        </router-link>

        <div class="hidden items-center gap-1 lg:flex">
          <router-link
            to="/home"
            class="rounded-lg px-3 py-2 text-sm font-medium text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-950 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
          >
            {{ t('pricingPage.navHome') }}
          </router-link>
          <router-link
            to="/pricing"
            class="rounded-lg bg-sky-50 px-3 py-2 text-sm font-semibold text-sky-700 dark:bg-sky-950/40 dark:text-sky-300"
          >
            {{ t('pricingPage.navPricing') }}
          </router-link>
        </div>

        <div class="flex items-center gap-1.5">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <button
            type="button"
            class="rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" size="md" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex items-center gap-1.5 rounded-lg bg-slate-950 px-3 py-2 text-xs font-semibold text-white transition-colors hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-200"
          >
            <Icon :name="isAuthenticated ? 'grid' : 'login'" size="sm" />
            <span class="hidden sm:inline">
              {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
            </span>
          </router-link>
        </div>
      </nav>
    </header>

    <main>
      <section
        class="relative overflow-hidden border-b border-slate-200 bg-white dark:border-dark-800 dark:bg-dark-950"
      >
        <div
          class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_15%_5%,rgba(14,165,233,0.13),transparent_34%),radial-gradient(circle_at_85%_5%,rgba(16,185,129,0.11),transparent_32%)]"
        ></div>
        <div class="relative mx-auto max-w-7xl px-4 py-12 text-center md:px-6 md:py-16">
          <div
            class="inline-flex items-center gap-2 rounded-full border border-sky-200 bg-sky-50 px-3.5 py-2 text-xs font-bold text-sky-700 dark:border-sky-900/70 dark:bg-sky-950/40 dark:text-sky-300"
          >
            <Icon name="sparkles" size="sm" />
            {{ t('pricingPage.eyebrow') }}
          </div>
          <h1
            class="mx-auto mt-5 max-w-4xl text-4xl font-black tracking-tight text-slate-950 dark:text-white md:text-5xl"
          >
            {{ t('pricingPage.title') }}
          </h1>
          <p
            class="mx-auto mt-4 max-w-3xl text-sm leading-7 text-slate-600 dark:text-dark-300 md:text-base"
          >
            {{ t('pricingPage.description', { rate: rechargeRateLabel }) }}
          </p>

          <article
            class="mx-auto mt-8 flex max-w-2xl flex-col items-center justify-between gap-5 rounded-2xl border border-emerald-200 bg-emerald-50/80 p-5 text-left shadow-sm dark:border-emerald-900/60 dark:bg-emerald-950/20 sm:flex-row"
          >
            <div>
              <p
                class="flex items-center justify-center gap-2 text-xs font-bold text-emerald-700 dark:text-emerald-300 sm:justify-start"
              >
                <Icon name="creditCard" size="sm" />
                {{ t('pricingPage.rechargeLabel') }}
              </p>
              <p
                class="mt-2 text-center text-3xl font-black text-slate-950 dark:text-white sm:text-left"
              >
                ¥1 = <span data-testid="recharge-rate">{{ rechargeRateLabel }}</span> USD
              </p>
            </div>
            <div class="max-w-sm text-center sm:text-right">
              <span
                class="inline-flex rounded-full border border-violet-200 bg-white/80 px-3 py-1 text-xs font-bold text-violet-700 dark:border-violet-900/60 dark:bg-dark-950/50 dark:text-violet-300"
              >
                {{ t('pricingPage.groupRate') }}
              </span>
              <p
                class="mt-2 text-xs leading-5 text-emerald-800/80 dark:text-emerald-200/80"
              >
                {{ t('pricingPage.formulaCash', { rate: rechargeRateLabel }) }}
              </p>
            </div>
          </article>
        </div>
      </section>

      <section class="mx-auto max-w-7xl px-4 py-8 md:px-6 lg:py-10">
        <div class="mb-4 flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-xl font-black text-slate-950 dark:text-white">
              {{ t('pricingPage.navPricing') }}
            </h2>
            <p class="mt-1 text-xs text-slate-500 dark:text-dark-400">
              {{ t('pricingPage.formulaCash', { rate: rechargeRateLabel }) }}
            </p>
          </div>
          <p
            data-testid="model-count"
            class="text-sm font-bold text-slate-700 dark:text-dark-200"
          >
            {{ t('pricingPage.visibleModels', { count: PUBLIC_MODEL_PRICES.length }) }}
          </p>
        </div>

        <div
          class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-dark-800 dark:bg-dark-900"
        >
          <div
            class="hidden grid-cols-[minmax(220px,1.45fr)_repeat(3,minmax(145px,1fr))] gap-4 border-b border-slate-200 bg-slate-50 px-5 py-3 text-xs font-bold uppercase tracking-wide text-slate-500 dark:border-dark-800 dark:bg-dark-950/50 dark:text-dark-400 md:grid"
          >
            <span>{{ t('pricingPage.allModels') }}</span>
            <span>{{ t('pricingPage.input') }} · {{ t('pricingPage.perMillionShort') }}</span>
            <span>{{ t('pricingPage.output') }} · {{ t('pricingPage.perMillionShort') }}</span>
            <span>{{ t('pricingPage.cacheRead') }} · {{ t('pricingPage.perMillionShort') }}</span>
          </div>

          <article
            v-for="model in PUBLIC_MODEL_PRICES"
            :key="model.id"
            :data-testid="`model-card-${model.id}`"
            class="grid gap-4 border-b border-slate-100 p-5 last:border-b-0 dark:border-dark-800 md:grid-cols-[minmax(220px,1.45fr)_repeat(3,minmax(145px,1fr))] md:items-center"
          >
            <div class="flex min-w-0 items-center gap-3">
              <span
                class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-slate-200 bg-white shadow-sm dark:border-dark-700"
              >
                <ModelIcon :model="model.id" size="27px" />
              </span>
              <div class="min-w-0">
                <h3 class="truncate font-mono text-sm font-black text-slate-950 dark:text-white">
                  {{ model.id }}
                </h3>
                <p class="mt-0.5 text-xs text-slate-500 dark:text-dark-400">
                  {{ providerLabel(model.family) }}
                </p>
              </div>
            </div>

            <div
              :data-testid="`input-price-${model.id}`"
              class="flex items-center justify-between rounded-xl bg-slate-50 px-3 py-2.5 font-mono text-sm font-bold text-slate-900 dark:bg-dark-950/50 dark:text-white md:block md:bg-transparent md:px-0 md:py-0 dark:md:bg-transparent"
            >
              <span class="font-sans text-xs font-medium text-slate-400 md:sr-only">
                {{ t('pricingPage.input') }} · {{ t('pricingPage.perMillionShort') }}
              </span>
              <span>{{ formatPrice(model.benchmarkInputUsdPerMillion) }}</span>
            </div>
            <div
              :data-testid="`output-price-${model.id}`"
              class="flex items-center justify-between rounded-xl bg-slate-50 px-3 py-2.5 font-mono text-sm font-bold text-slate-900 dark:bg-dark-950/50 dark:text-white md:block md:bg-transparent md:px-0 md:py-0 dark:md:bg-transparent"
            >
              <span class="font-sans text-xs font-medium text-slate-400 md:sr-only">
                {{ t('pricingPage.output') }} · {{ t('pricingPage.perMillionShort') }}
              </span>
              <span>{{ formatPrice(model.benchmarkOutputUsdPerMillion) }}</span>
            </div>
            <div
              :data-testid="`cache-price-${model.id}`"
              class="flex items-center justify-between rounded-xl bg-slate-50 px-3 py-2.5 font-mono text-sm font-bold text-slate-900 dark:bg-dark-950/50 dark:text-white md:block md:bg-transparent md:px-0 md:py-0 dark:md:bg-transparent"
            >
              <span class="font-sans text-xs font-medium text-slate-400 md:sr-only">
                {{ t('pricingPage.cacheRead') }} · {{ t('pricingPage.perMillionShort') }}
              </span>
              <span>{{ formatPrice(model.benchmarkCacheReadUsdPerMillion) }}</span>
            </div>
          </article>
        </div>

        <p
          class="mt-5 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-xs leading-5 text-amber-800 dark:border-amber-900/60 dark:bg-amber-950/20 dark:text-amber-200"
        >
          {{ t('pricingPage.disclaimer') }}
        </p>
      </section>

      <section
        class="border-t border-slate-200 bg-slate-950 px-4 py-12 text-white dark:border-dark-800 md:px-6"
      >
        <div
          class="mx-auto flex max-w-5xl flex-col items-center justify-between gap-6 text-center md:flex-row md:text-left"
        >
          <div>
            <h2 class="text-2xl font-black">{{ t('pricingPage.ctaTitle') }}</h2>
            <p class="mt-2 text-sm leading-6 text-slate-300">
              {{ t('pricingPage.ctaDescription') }}
            </p>
          </div>
          <div class="flex w-full flex-col gap-3 sm:w-auto sm:flex-row">
            <router-link
              :to="ctaPath"
              class="inline-flex items-center justify-center gap-2 rounded-xl bg-sky-400 px-5 py-3 text-sm font-bold text-slate-950 transition-colors hover:bg-sky-300"
            >
              {{ ctaLabel }}
              <Icon name="arrowRight" size="sm" />
            </router-link>
            <router-link
              to="/home"
              class="inline-flex items-center justify-center rounded-xl border border-white/15 bg-white/5 px-5 py-3 text-sm font-semibold text-white transition-colors hover:bg-white/10"
            >
              {{ t('pricingPage.ctaSecondary') }}
            </router-link>
          </div>
        </div>
      </section>
    </main>

    <footer
      class="border-t border-slate-200 bg-white px-4 py-5 text-center text-xs text-slate-500 dark:border-dark-800 dark:bg-dark-950 dark:text-dark-400"
    >
      © 2026 ISACAI · {{ siteName }}
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import {
  PUBLIC_MODEL_PRICES,
  PUBLIC_RECHARGE_USD_PER_CNY,
  benchmarkUsdToEffectiveCny,
  formatCompactNumber,
  type PublicModelPrice,
} from '@/utils/pricing'
import { sanitizeUrl } from '@/utils/url'

const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const isDark = ref(document.documentElement.classList.contains('dark'))

const siteName = computed(
  () => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'ISACAI',
)
const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''),
)
const isRtl = computed(() => locale.value === 'ar')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const ctaPath = computed(() => (isAuthenticated.value ? dashboardPath.value : '/register'))
const ctaLabel = computed(() =>
  isAuthenticated.value ? t('pricingPage.ctaAuthenticated') : t('pricingPage.ctaPrimary'),
)
const rechargeUsdPerCny = computed(() => {
  const configuredRate = Number(appStore.cachedPublicSettings?.balance_recharge_multiplier)
  return Number.isFinite(configuredRate) && configuredRate > 0
    ? configuredRate
    : PUBLIC_RECHARGE_USD_PER_CNY
})
const rechargeRateLabel = computed(() => formatCompactNumber(rechargeUsdPerCny.value, 6))

function providerLabel(family: PublicModelPrice['family']): string {
  return family === 'gpt' ? 'OpenAI' : 'Anthropic'
}

function formatPrice(benchmarkUsdPerMillion: number): string {
  const amount = benchmarkUsdToEffectiveCny(
    benchmarkUsdPerMillion,
    rechargeUsdPerCny.value,
  )
  return `¥${formatCompactNumber(amount, 6)}`
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
})
</script>
