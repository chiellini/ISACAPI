<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="flex min-h-screen flex-col">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="min-h-screen w-full flex-1 border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else class="flex-1" v-html="homeContent"></div>
    <footer class="bg-white px-4 py-5 dark:bg-dark-950" dir="ltr">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-2 text-center text-sm text-slate-500 dark:text-dark-400">
        <img :src="companyIconUrl" alt="ISACAI" class="h-10 w-10 rounded-lg object-contain" />
        <p>© 2026 ISACAI. All rights reserved. 软件开发交流联系方式：1027890648</p>
        <a
          href="https://beian.miit.gov.cn"
          target="_blank"
          rel="noopener noreferrer"
          class="transition-colors hover:text-slate-800 dark:hover:text-white"
        >
          粤ICP备2026050877号
        </a>
      </div>
    </footer>
  </div>

  <div
    v-else
    :dir="isRtl ? 'rtl' : 'ltr'"
    class="relative flex min-h-screen flex-col overflow-hidden bg-slate-50 text-slate-950 dark:bg-dark-950 dark:text-white"
  >
    <header class="relative z-20 border-b border-slate-200/80 bg-white/90 px-4 py-3 backdrop-blur dark:border-dark-800 dark:bg-dark-950/90">
      <nav class="mx-auto flex max-w-6xl items-center justify-between gap-4">
        <div class="flex min-w-0 items-center gap-3">
          <div class="h-10 w-10 shrink-0 overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
            <img :src="siteLogo || '/logo.png'" alt="ISACAI" class="h-full w-full object-contain" />
          </div>
          <div class="min-w-0">
            <p class="truncate text-sm font-semibold text-slate-950 dark:text-white">
              {{ siteName }}
            </p>
            <p class="hidden text-xs text-slate-500 dark:text-dark-400 sm:block">
              {{ t('home.nav.tagline') }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <a
            :href="CC_SWITCH_DOWNLOAD_LINKS.officialSite"
            target="_blank"
            rel="noopener noreferrer"
            class="hidden items-center gap-1.5 rounded-lg border border-sky-200 bg-sky-50 px-3 py-2 text-xs font-semibold text-sky-700 transition-colors hover:border-sky-300 hover:bg-sky-100 dark:border-sky-900/60 dark:bg-sky-950/40 dark:text-sky-300 dark:hover:bg-sky-950/70 md:inline-flex"
          >
            <Icon name="bolt" size="sm" />
            CC-Switch
          </a>

          <LocaleSwitcher />

          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <button
            @click="toggleTheme"
            class="rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-1.5 rounded-lg bg-slate-950 px-3 py-2 text-xs font-medium text-white transition-colors hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-200"
          >
            <span class="flex h-5 w-5 items-center justify-center rounded-md bg-primary-500 text-[10px] font-semibold text-white">
              {{ userInitial }}
            </span>
            <span class="hidden sm:inline">{{ t('home.dashboard') }}</span>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center gap-1.5 rounded-lg bg-slate-950 px-3 py-2 text-xs font-medium text-white transition-colors hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-200"
          >
            <Icon name="login" size="sm" />
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10 flex-1">
      <section class="relative overflow-hidden border-b border-slate-800 bg-slate-950 text-white">
        <div class="pointer-events-none absolute -left-24 top-12 h-72 w-72 rounded-full bg-sky-500/20 blur-3xl"></div>
        <div class="pointer-events-none absolute -right-20 bottom-0 h-80 w-80 rounded-full bg-violet-500/20 blur-3xl"></div>

        <div class="relative mx-auto grid max-w-6xl gap-12 px-4 py-14 md:px-6 md:py-20 lg:grid-cols-[1.08fr_0.92fr] lg:items-center lg:gap-14">
          <div :class="['text-center lg:text-left', isRtl ? 'lg:text-right' : '']">
            <div class="mb-6 inline-flex items-center gap-2 rounded-full border border-sky-400/30 bg-sky-400/10 px-3.5 py-2 text-xs font-semibold text-sky-200 shadow-lg shadow-sky-950/30">
              <Icon name="bolt" size="sm" />
              {{ t('home.ccSwitch.badge') }}
            </div>

            <p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">
              {{ siteName }} · {{ siteSubtitle }}
            </p>
            <h1 class="mx-auto mt-4 max-w-3xl text-4xl font-bold leading-tight tracking-tight text-white md:text-5xl lg:mx-0 lg:text-[3.25rem]">
              {{ t('home.ccSwitch.title') }}
            </h1>
            <p class="mx-auto mt-6 max-w-2xl text-base leading-8 text-slate-300 md:text-lg lg:mx-0">
              {{ t('home.ccSwitch.description') }}
            </p>

            <div class="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row lg:justify-start">
              <router-link
                to="/keys"
                class="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-sky-400 px-6 py-3.5 text-sm font-bold text-slate-950 shadow-lg shadow-sky-500/20 transition-all hover:-translate-y-0.5 hover:bg-sky-300 sm:w-auto"
              >
                <Icon name="bolt" size="sm" />
                {{ t('home.ccSwitch.primaryAction') }}
              </router-link>
              <a
                :href="CC_SWITCH_DOWNLOAD_LINKS.officialSite"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-white/15 bg-white/5 px-6 py-3.5 text-sm font-semibold text-white transition-colors hover:border-white/30 hover:bg-white/10 sm:w-auto"
              >
                <Icon name="download" size="sm" />
                {{ t('home.ccSwitch.downloadAction') }}
              </a>
            </div>

            <router-link
              to="/cc-switch"
              class="mt-5 inline-flex items-center gap-1.5 text-sm font-medium text-sky-300 transition-colors hover:text-sky-200"
            >
              {{ t('home.ccSwitch.guideAction') }}
              <Icon name="arrowRight" size="sm" />
            </router-link>

            <div class="mt-8 grid grid-cols-2 gap-2.5">
              <div
                v-for="mode in deploymentModes"
                :key="mode.titleKey"
                class="flex items-center gap-2.5 rounded-xl border border-white/10 bg-white/5 px-3 py-2.5 text-left"
              >
                <span :class="['flex h-8 w-8 shrink-0 items-center justify-center rounded-lg', mode.iconClass]">
                  <Icon :name="mode.icon" size="sm" />
                </span>
                <span class="min-w-0">
                  <span class="block truncate text-xs font-bold text-white">{{ t(mode.titleKey) }}</span>
                  <span class="mt-0.5 block truncate text-[10px] text-slate-400 sm:text-[11px]">
                    {{ t(mode.descriptionKey) }}
                  </span>
                </span>
              </div>
            </div>
          </div>

          <div class="mx-auto w-full max-w-xl lg:max-w-none">
            <div class="relative overflow-hidden rounded-2xl border border-white/15 bg-white/[0.07] p-5 shadow-2xl shadow-black/30 backdrop-blur md:p-6">
              <div class="pointer-events-none absolute right-0 top-0 h-32 w-32 rounded-full bg-sky-400/15 blur-3xl"></div>

              <div class="relative flex items-center justify-between gap-4">
                <div class="flex items-center gap-3">
                  <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-sky-400 text-slate-950 shadow-lg shadow-sky-400/20">
                    <Icon name="bolt" size="lg" />
                  </div>
                  <div>
                    <p class="text-base font-bold text-white">CC-Switch</p>
                    <p class="text-xs text-slate-400">ISACAI Setup Hub</p>
                  </div>
                </div>
                <span class="rounded-full border border-emerald-400/25 bg-emerald-400/10 px-3 py-1 text-[11px] font-semibold text-emerald-300">
                  {{ t('home.ccSwitch.oneClick') }}
                </span>
              </div>

              <h2 class="relative mt-6 max-w-md text-xl font-semibold leading-8 text-white md:text-2xl">
                {{ t('home.ccSwitch.panelTitle') }}
              </h2>

              <div class="relative mt-6 grid grid-cols-2 gap-3">
                <div
                  v-for="client in integrationClients"
                  :key="client.name"
                  class="flex min-h-20 items-center gap-3 rounded-xl border border-white/10 bg-slate-950/50 p-3"
                >
                  <div :class="['flex h-10 w-10 shrink-0 items-center justify-center rounded-lg', client.panelClass]">
                    <ModelIcon v-if="'model' in client" :model="client.model" size="25px" />
                    <Icon v-else :name="client.icon" size="md" />
                  </div>
                  <div class="min-w-0">
                    <p class="truncate text-xs font-semibold text-white sm:text-sm">
                      {{ 'panelName' in client ? client.panelName : client.name }}
                    </p>
                    <span class="mt-1 flex items-center gap-1 text-[10px] font-medium text-emerald-300">
                      <span class="h-1.5 w-1.5 rounded-full bg-emerald-400"></span>
                      {{ t('home.ccSwitch.ready') }}
                    </span>
                  </div>
                </div>
              </div>

              <div class="relative mt-4 grid gap-3 sm:grid-cols-2">
                <div class="flex items-start gap-3 rounded-xl border border-sky-400/15 bg-sky-400/10 p-3.5">
                  <Icon name="download" size="md" class="mt-0.5 shrink-0 text-sky-300" />
                  <p class="text-xs font-medium leading-5 text-sky-100">{{ t('home.ccSwitch.localSetup') }}</p>
                </div>
                <div class="flex items-start gap-3 rounded-xl border border-violet-400/15 bg-violet-400/10 p-3.5">
                  <Icon name="terminal" size="md" class="mt-0.5 shrink-0 text-violet-300" />
                  <p class="text-xs font-medium leading-5 text-violet-100">{{ t('home.ccSwitch.remoteSetup') }}</p>
                </div>
              </div>

              <div class="relative mt-4 flex items-start gap-2.5 rounded-xl border border-emerald-400/15 bg-emerald-400/10 px-4 py-3 text-xs leading-5 text-emerald-100">
                <Icon name="checkCircle" size="sm" class="mt-0.5 shrink-0 text-emerald-300" />
                {{ t('home.ccSwitch.noManual') }}
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="bg-white px-4 py-12 dark:bg-dark-950 md:px-6 md:py-16">
        <div class="mx-auto max-w-6xl">
          <div class="mx-auto max-w-3xl text-center">
            <div class="inline-flex items-center gap-2 rounded-full bg-sky-50 px-3 py-1.5 text-xs font-semibold text-sky-700 dark:bg-sky-950/40 dark:text-sky-300">
              <Icon name="terminal" size="sm" />
              {{ t('home.integrations.eyebrow') }}
            </div>
            <h2 class="mt-4 text-2xl font-bold text-slate-950 dark:text-white md:text-3xl">
              {{ t('home.integrations.title') }}
            </h2>
            <p class="mt-3 text-sm leading-7 text-slate-500 dark:text-dark-400">
              {{ t('home.integrations.description') }}
            </p>
          </div>

          <div class="mt-9 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <article
              v-for="client in integrationClients"
              :key="client.name"
              class="group rounded-2xl border border-slate-200 bg-slate-50/70 p-5 transition-all hover:-translate-y-1 hover:border-sky-300 hover:bg-white hover:shadow-lg dark:border-dark-800 dark:bg-dark-900/70 dark:hover:border-sky-700 dark:hover:bg-dark-900"
            >
              <div :class="['flex h-12 w-12 items-center justify-center rounded-xl', client.cardClass]">
                <ModelIcon v-if="'model' in client" :model="client.model" size="29px" />
                <Icon v-else :name="client.icon" size="lg" />
              </div>
              <h3 class="mt-5 text-base font-bold text-slate-950 dark:text-white">{{ client.name }}</h3>
              <p class="mt-2 text-sm leading-6 text-slate-500 dark:text-dark-400">
                {{ t(client.descriptionKey) }}
              </p>
            </article>
          </div>
        </div>
      </section>

      <section class="border-y border-slate-200 bg-white px-4 py-12 dark:border-dark-800 dark:bg-dark-950 md:px-6 md:py-14">
        <div class="mx-auto max-w-6xl">
          <div class="grid gap-6 lg:grid-cols-[0.95fr_1.05fr] lg:items-center">
            <div>
              <p class="text-xs font-semibold uppercase tracking-wide text-emerald-600 dark:text-emerald-300">
                {{ t('home.pricing.eyebrow') }}
              </p>
              <h2 class="mt-2 text-2xl font-semibold text-slate-950 dark:text-white md:text-3xl">
                {{ t('home.pricing.title') }}
              </h2>
              <p class="mt-3 max-w-2xl text-sm leading-6 text-slate-500 dark:text-dark-400">
                {{ t('home.pricing.description') }}
              </p>
            </div>
            <div class="grid gap-3 sm:grid-cols-2">
              <div class="rounded-lg border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-900/60 dark:bg-emerald-950/20">
                <div class="flex items-center gap-2 text-sm font-semibold text-emerald-800 dark:text-emerald-200">
                  <Icon name="creditCard" size="sm" />
                  {{ t('home.pricing.rechargeLabel') }}
                </div>
                <p class="mt-3 text-2xl font-bold text-slate-950 dark:text-white">
                  {{ t('home.pricing.rechargeValue', { usd: formatCompactNumber(PUBLIC_RECHARGE_USD_PER_CNY) }) }}
                </p>
                <p class="mt-2 text-xs leading-5 text-emerald-800/80 dark:text-emerald-200/80">
                  {{ t('home.pricing.rechargeHint') }}
                </p>
              </div>
              <div class="rounded-lg border border-sky-200 bg-sky-50 p-4 dark:border-sky-900/60 dark:bg-sky-950/20">
                <div class="flex items-center gap-2 text-sm font-semibold text-sky-800 dark:text-sky-200">
                  <Icon name="calculator" size="sm" />
                  {{ t('home.pricing.tokenLabel') }}
                </div>
                <p class="mt-3 text-2xl font-bold text-slate-950 dark:text-white">
                  {{ t('home.pricing.tokenValue', { divisor: INTERNAL_TOKEN_PRICE_DIVISOR }) }}
                </p>
                <p class="mt-2 text-xs leading-5 text-sky-800/80 dark:text-sky-200/80">
                  {{ t('home.pricing.tokenHint') }}
                </p>
              </div>
            </div>
          </div>
          <ModelPriceComparison class="mt-8" />
        </div>
      </section>

      <section class="border-y border-slate-200 bg-slate-50 px-4 py-12 dark:border-dark-800 dark:bg-dark-900/40 md:px-6 md:py-16">
        <div class="mx-auto max-w-6xl">
          <div class="mx-auto max-w-2xl text-center">
            <h2 class="text-2xl font-bold text-slate-950 dark:text-white md:text-3xl">
              {{ t('home.providers.title') }}
            </h2>
            <p class="mt-3 text-sm leading-7 text-slate-500 dark:text-dark-400">
              {{ t('home.providers.description') }}
            </p>
          </div>

          <div class="mt-9 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
            <div
              v-for="provider in providerLogos"
              :key="provider.labelKey"
              class="group flex min-h-32 flex-col items-center justify-center gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm transition-all hover:-translate-y-1 hover:border-sky-300 hover:shadow-lg dark:border-dark-800 dark:bg-dark-900 dark:hover:border-sky-700"
              :title="t(provider.labelKey)"
            >
              <div class="flex h-14 w-14 items-center justify-center rounded-xl bg-slate-50 text-slate-900 transition-colors group-hover:bg-sky-50 dark:bg-dark-950 dark:text-white dark:group-hover:bg-sky-950/40">
                <ModelIcon :model="provider.model" size="34px" />
              </div>
              <span class="max-w-full truncate text-sm font-semibold text-slate-700 dark:text-dark-200">
                {{ t(provider.labelKey) }}
              </span>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer class="relative z-10 bg-white px-4 py-5 dark:bg-dark-950" dir="ltr">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-2 text-center text-sm text-slate-500 dark:text-dark-400">
        <img :src="companyIconUrl" alt="ISACAI" class="h-10 w-10 rounded-lg object-contain" />
        <p>© 2026 ISACAI. All rights reserved. 软件开发交流联系方式：1027890648</p>
        <div class="flex flex-wrap items-center justify-center gap-4">
          <a
            href="https://beian.miit.gov.cn"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-slate-800 dark:hover:text-white"
          >
            粤ICP备2026050877号
          </a>
          <router-link
            v-if="publicStatusEnabled"
            to="/status"
            class="inline-flex items-center gap-1.5 transition-colors hover:text-slate-800 dark:hover:text-white"
          >
            <span class="inline-flex h-2 w-2 rounded-full bg-emerald-500"></span>
            {{ t('publicStatus.title') }}
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-slate-800 dark:hover:text-white"
          >
            {{ t('home.docs') }}
          </a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-slate-800 dark:hover:text-white"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>

    <transition name="notice-fade">
      <div
        v-if="showNotice"
        class="fixed inset-0 z-50 flex items-start justify-center bg-slate-950/60 px-4 py-8 backdrop-blur-sm sm:items-center"
      >
        <section class="relative w-full max-w-4xl rounded-lg border border-slate-200 bg-white shadow-2xl dark:border-dark-700 dark:bg-dark-900">
          <div class="flex items-center justify-between border-b border-slate-200 px-5 py-4 dark:border-dark-700">
            <h2 class="text-xl font-bold text-slate-950 dark:text-white">
              {{ t('home.notice.title') }}
            </h2>
            <div class="flex items-center gap-3">
              <span class="hidden items-center gap-1 rounded-lg bg-sky-50 px-3 py-2 text-xs font-semibold text-sky-700 dark:bg-sky-950/40 dark:text-sky-300 sm:inline-flex">
                <Icon name="bell" size="sm" />
                {{ t('home.notice.tabNotice') }}
              </span>
              <span class="hidden items-center gap-1 text-xs font-medium text-slate-500 dark:text-dark-400 sm:inline-flex">
                <Icon name="infoCircle" size="sm" />
                {{ t('home.notice.tabSystem') }}
              </span>
              <button
                @click="closeNoticePermanently"
                class="rounded-lg p-2 text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
                :title="t('common.close')"
              >
                <Icon name="x" size="md" />
              </button>
            </div>
          </div>

          <div class="grid gap-6 px-5 py-6 md:grid-cols-[1fr_190px] md:items-end">
            <div class="space-y-3 text-sm leading-7 text-slate-700 dark:text-dark-200">
              <p>
                <span class="font-semibold text-slate-950 dark:text-white">{{ t('home.notice.important') }}</span>
                {{ t('home.notice.qqGroupLabel') }}
                <span class="font-semibold text-red-500">1027890648</span>
                {{ t('home.notice.qqGroupSuffix') }}
              </p>
              <p>{{ t('home.notice.trust') }}</p>
              <p>
                {{ t('home.notice.status') }}
                <a
                  href="/"
                  class="font-medium text-sky-600 hover:text-sky-700 dark:text-sky-300 dark:hover:text-sky-200"
                >
                  {{ siteName }}
                </a>
              </p>
            </div>

            <div class="hidden justify-center md:flex">
              <div class="flex h-36 w-36 items-center justify-center rounded-lg border border-slate-200 bg-slate-50 p-4 dark:border-dark-700 dark:bg-dark-950">
                <img :src="companyIconUrl" alt="ISACAI" class="max-h-full max-w-full object-contain" />
              </div>
            </div>
          </div>

          <div class="flex flex-wrap items-center justify-end gap-3 border-t border-slate-200 px-5 py-4 dark:border-dark-700">
            <button
              @click="closeNoticeToday"
              class="rounded-lg bg-slate-100 px-4 py-2 text-sm font-semibold text-slate-700 transition-colors hover:bg-slate-200 dark:bg-dark-800 dark:text-dark-200 dark:hover:bg-dark-700"
            >
              {{ t('home.notice.closeToday') }}
            </button>
            <button
              @click="closeNoticePermanently"
              class="rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-sky-700"
            >
              {{ t('home.notice.close') }}
            </button>
          </div>
        </section>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import ModelPriceComparison from '@/components/common/ModelPriceComparison.vue'
import Icon from '@/components/icons/Icon.vue'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { CC_SWITCH_DOWNLOAD_LINKS } from '@/utils/ccswitchImport'
import { INTERNAL_TOKEN_PRICE_DIVISOR, PUBLIC_RECHARGE_USD_PER_CNY, formatCompactNumber } from '@/utils/pricing'
import { sanitizeUrl } from '@/utils/url'

const { t, locale } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const companyIconUrl = '/logo.png'

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const isRtl = computed(() => locale.value === 'ar')
const showNotice = ref(false)

const NOTICE_PERMANENT_KEY = 'isacai_home_notice_closed'
const NOTICE_DATE_KEY = 'isacai_home_notice_date'

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return 'I'
  return user.email.charAt(0).toUpperCase()
})

const integrationClients = [
  {
    name: 'Codex',
    model: 'gpt-4.1',
    icon: 'terminal',
    panelClass: 'bg-emerald-400/15 text-emerald-300',
    cardClass: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300',
    descriptionKey: 'home.integrations.codexDescription'
  },
  {
    name: 'Claude Code',
    model: 'claude-3-5-sonnet',
    icon: 'terminal',
    panelClass: 'bg-orange-400/15 text-orange-300',
    cardClass: 'bg-orange-50 text-orange-700 dark:bg-orange-950/40 dark:text-orange-300',
    descriptionKey: 'home.integrations.claudeCodeDescription'
  },
  {
    name: 'Gemini CLI',
    model: 'gemini-2.5-pro',
    icon: 'sparkles',
    panelClass: 'bg-blue-400/15 text-blue-300',
    cardClass: 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300',
    descriptionKey: 'home.integrations.geminiDescription'
  },
  {
    name: 'VS Code / Cursor / IDE',
    panelName: 'IDE / Terminal',
    icon: 'grid',
    panelClass: 'bg-violet-400/15 text-violet-300',
    cardClass: 'bg-violet-50 text-violet-700 dark:bg-violet-950/40 dark:text-violet-300',
    descriptionKey: 'home.integrations.ideDescription'
  }
] as const

const deploymentModes = [
  {
    titleKey: 'home.deployment.ccSwitch.title',
    descriptionKey: 'home.deployment.ccSwitch.description',
    icon: 'bolt',
    iconClass: 'bg-sky-400/15 text-sky-300',
  },
  {
    titleKey: 'home.deployment.terminal.title',
    descriptionKey: 'home.deployment.terminal.description',
    icon: 'terminal',
    iconClass: 'bg-violet-400/15 text-violet-300',
  },
  {
    titleKey: 'home.deployment.server.title',
    descriptionKey: 'home.deployment.server.description',
    icon: 'server',
    iconClass: 'bg-amber-400/15 text-amber-300',
  },
  {
    titleKey: 'home.deployment.client.title',
    descriptionKey: 'home.deployment.client.description',
    icon: 'monitor',
    iconClass: 'bg-emerald-400/15 text-emerald-300',
  },
] as const

const providerLogos = [
  { model: 'gpt-4.1', labelKey: 'home.providers.items.openai' },
  { model: 'claude-3-5-sonnet', labelKey: 'home.providers.items.claude' },
  { model: 'gemini-2.5-pro', labelKey: 'home.providers.items.gemini' },
  { model: 'grok-4', labelKey: 'home.providers.items.xai' },
  { model: 'deepseek-chat', labelKey: 'home.providers.items.deepseek' },
  { model: 'glm-4-plus', labelKey: 'home.providers.items.zhipu' }
] as const

const publicStatusEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.publicStatus))
const githubUrl = 'https://github.com/chiellini/ISACAPI'

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

function todayKey() {
  return new Date().toISOString().slice(0, 10)
}

function initNotice() {
  const closedPermanently = localStorage.getItem(NOTICE_PERMANENT_KEY) === '1'
  const closedToday = localStorage.getItem(NOTICE_DATE_KEY) === todayKey()
  showNotice.value = !closedPermanently && !closedToday
}

function closeNoticeToday() {
  localStorage.setItem(NOTICE_DATE_KEY, todayKey())
  showNotice.value = false
}

function closeNoticePermanently() {
  localStorage.setItem(NOTICE_PERMANENT_KEY, '1')
  showNotice.value = false
}

onMounted(() => {
  initTheme()
  initNotice()

  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.notice-fade-enter-active,
.notice-fade-leave-active {
  transition: opacity 180ms ease;
}

.notice-fade-enter-from,
.notice-fade-leave-to {
  opacity: 0;
}
</style>
