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
      <section class="border-b border-slate-200/80 bg-gradient-to-b from-white via-sky-50/40 to-slate-50 dark:border-dark-800 dark:from-dark-950 dark:via-dark-900 dark:to-dark-950">
        <div class="mx-auto grid max-w-6xl gap-10 px-4 py-12 md:px-6 md:py-16 lg:grid-cols-[1.02fr_0.98fr] lg:items-center lg:gap-14">
          <div :class="['text-center lg:text-left', isRtl ? 'lg:text-right' : '']">
            <div class="mb-5 inline-flex items-center gap-2 rounded-lg border border-sky-200 bg-white px-3 py-2 text-xs font-medium text-sky-700 shadow-sm dark:border-sky-900/60 dark:bg-dark-900 dark:text-sky-300">
              <Icon name="sparkles" size="sm" />
              {{ t('home.heroEyebrow') }}
            </div>

            <h1 class="mx-auto max-w-3xl text-4xl font-bold tracking-normal text-slate-950 dark:text-white md:text-5xl lg:mx-0 lg:text-6xl">
              {{ siteName }}
            </h1>
            <p class="mx-auto mt-5 max-w-2xl text-lg leading-8 text-slate-600 dark:text-dark-300 lg:mx-0">
              {{ siteSubtitle }}
            </p>

            <div class="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row lg:justify-start">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-primary-600 px-5 py-3 text-sm font-semibold text-white shadow-md shadow-primary-600/20 transition-colors hover:bg-primary-700 sm:w-auto"
              >
                <Icon name="play" size="sm" />
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
              </router-link>
              <a
                v-if="docUrl"
                :href="docUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-slate-200 bg-white px-5 py-3 text-sm font-semibold text-slate-700 transition-colors hover:bg-slate-50 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-200 dark:hover:bg-dark-800 sm:w-auto"
              >
                <Icon name="document" size="sm" />
                {{ t('home.docs') }}
              </a>
            </div>

            <dl class="mt-9 grid grid-cols-3 gap-3">
              <div
                v-for="stat in heroStats"
                :key="stat.labelKey"
                class="rounded-lg border border-slate-200 bg-white/70 px-3 py-4 text-center shadow-sm dark:border-dark-800 dark:bg-dark-900/70"
              >
                <dt class="text-xl font-bold text-slate-950 dark:text-white">{{ stat.value }}</dt>
                <dd class="mt-1 text-xs text-slate-500 dark:text-dark-400">{{ t(stat.labelKey) }}</dd>
              </div>
            </dl>
          </div>

          <div class="mx-auto w-full max-w-xl lg:max-w-none">
            <div class="rounded-lg border border-slate-200 bg-white shadow-xl shadow-slate-200/60 dark:border-dark-800 dark:bg-dark-900 dark:shadow-black/30">
              <div class="flex items-center justify-between border-b border-slate-200 px-4 py-3 dark:border-dark-800">
                <div class="flex items-center gap-2">
                  <span class="h-3 w-3 rounded-full bg-red-400"></span>
                  <span class="h-3 w-3 rounded-full bg-amber-400"></span>
                  <span class="h-3 w-3 rounded-full bg-emerald-400"></span>
                </div>
                <span class="text-xs font-medium text-slate-500 dark:text-dark-400">
                  {{ t('home.apiCard.protocol') }}
                </span>
              </div>

              <div class="space-y-5 p-5">
                <div class="flex items-start justify-between gap-4">
                  <div>
                    <p class="text-sm font-semibold text-slate-950 dark:text-white">
                      {{ t('home.apiCard.title') }}
                    </p>
                    <p class="mt-1 text-xs text-slate-500 dark:text-dark-400">
                      {{ t('home.apiCard.subtitle') }}
                    </p>
                  </div>
                  <span class="rounded-md bg-emerald-50 px-2 py-1 text-xs font-semibold text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300">
                    200 OK
                  </span>
                </div>

                <div class="rounded-lg border border-slate-200 bg-slate-950 p-4 text-left font-mono text-xs text-slate-200 shadow-inner dark:border-dark-700">
                  <div class="mb-3 flex items-center justify-between gap-3">
                    <span class="text-slate-400">{{ t('home.apiCard.endpoint') }}</span>
                    <button
                      @click="copyEndpoint"
                      class="inline-flex items-center gap-1 rounded-md bg-white/10 px-2 py-1 text-[11px] font-medium text-white transition-colors hover:bg-white/15"
                    >
                      <Icon name="copy" size="xs" />
                      {{ endpointCopied ? t('home.apiCard.copied') : t('home.apiCard.copy') }}
                    </button>
                  </div>
                  <div class="break-all text-sky-300">{{ chatEndpoint }}</div>
                  <pre class="mt-4 overflow-x-auto text-slate-300"><code>curl {{ chatEndpoint }} \
  -H "Authorization: Bearer sk-..." \
  -H "Content-Type: application/json"</code></pre>
                </div>

                <div class="grid gap-3 sm:grid-cols-3">
                  <div
                    v-for="tag in apiTags"
                    :key="tag.labelKey"
                    class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-3 dark:border-dark-800 dark:bg-dark-950/60"
                  >
                    <Icon :name="tag.icon" size="sm" :class="tag.iconClass" />
                    <p class="mt-2 text-xs font-medium text-slate-700 dark:text-dark-200">
                      {{ t(tag.labelKey) }}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="bg-white px-4 py-12 dark:bg-dark-950 md:px-6 md:py-14">
        <div class="mx-auto max-w-6xl">
          <div class="text-center">
            <h2 class="text-2xl font-semibold text-slate-950 dark:text-white md:text-3xl">
              {{ t('home.providers.title') }}
            </h2>
            <p class="mt-3 text-sm text-slate-500 dark:text-dark-400">
              {{ t('home.providers.description') }}
            </p>
          </div>

          <div class="mt-9 grid grid-cols-3 gap-4 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-10">
            <div
              v-for="provider in providerLogos"
              :key="provider.labelKey"
              class="group flex min-h-24 flex-col items-center justify-center gap-2 rounded-lg border border-slate-200 bg-white p-3 shadow-sm transition-all hover:-translate-y-0.5 hover:border-sky-300 hover:shadow-md dark:border-dark-800 dark:bg-dark-900 dark:hover:border-sky-700"
              :title="t(provider.labelKey)"
            >
              <div class="flex h-11 w-11 items-center justify-center rounded-lg bg-slate-50 text-slate-900 transition-colors group-hover:bg-sky-50 dark:bg-dark-950 dark:text-white dark:group-hover:bg-sky-950/40">
                <ModelIcon :model="provider.model" size="30px" />
              </div>
              <span class="max-w-full truncate text-xs font-medium text-slate-600 dark:text-dark-300">
                {{ t(provider.labelKey) }}
              </span>
            </div>
          </div>

          <div class="mt-8 rounded-lg border border-amber-200 bg-amber-50 p-4 dark:border-amber-900/60 dark:bg-amber-950/20">
            <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
              <div>
                <h3 class="text-sm font-semibold text-amber-900 dark:text-amber-200">
                  {{ t('home.domestic.title') }}
                </h3>
                <p class="mt-1 text-xs leading-6 text-amber-800/80 dark:text-amber-200/80">
                  {{ t('home.domestic.description') }}
                </p>
              </div>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="provider in domesticProviderLogos"
                  :key="provider.labelKey"
                  class="rounded-md border border-amber-200 bg-white px-2.5 py-1 text-xs font-medium text-amber-900 dark:border-amber-800 dark:bg-dark-900 dark:text-amber-200"
                >
                  {{ t(provider.labelKey) }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="border-y border-slate-200 bg-slate-50 px-4 py-12 dark:border-dark-800 dark:bg-dark-900/40 md:px-6 md:py-14">
        <div class="mx-auto max-w-6xl">
          <div class="grid gap-5 md:grid-cols-4">
            <article
              v-for="feature in featureCards"
              :key="feature.titleKey"
              class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-dark-800 dark:bg-dark-900"
            >
              <div :class="['mb-4 flex h-10 w-10 items-center justify-center rounded-lg', feature.iconClass]">
                <Icon :name="feature.icon" size="md" />
              </div>
              <h3 class="text-base font-semibold text-slate-950 dark:text-white">
                {{ t(feature.titleKey) }}
              </h3>
              <p class="mt-2 text-sm leading-6 text-slate-500 dark:text-dark-400">
                {{ t(feature.descKey) }}
              </p>
            </article>
          </div>
        </div>
      </section>
    </main>

    <footer class="relative z-10 bg-white px-4 py-5 dark:bg-dark-950" dir="ltr">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-2 text-center text-sm text-slate-500 dark:text-dark-400">
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
                <img :src="siteLogo || '/logo.png'" alt="ISACAI" class="max-h-full max-w-full object-contain" />
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
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'

const { t, locale } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const isRtl = computed(() => locale.value === 'ar')
const endpointCopied = ref(false)
const apiBaseUrl = ref('')
const showNotice = ref(false)

const NOTICE_PERMANENT_KEY = 'isacai_home_notice_closed'
const NOTICE_DATE_KEY = 'isacai_home_notice_date'

const chatEndpoint = computed(() => `${apiBaseUrl.value || 'https://api.isacai.cn'}/v1/chat/completions`)

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return 'I'
  return user.email.charAt(0).toUpperCase()
})

const heroStats = [
  { value: '30+', labelKey: 'home.stats.providers' },
  { value: '4', labelKey: 'home.stats.languages' },
  { value: 'OpenAI', labelKey: 'home.stats.compatible' }
]

const apiTags = [
  { icon: 'swap', iconClass: 'text-sky-600 dark:text-sky-300', labelKey: 'home.tags.subscriptionToApi' },
  { icon: 'shield', iconClass: 'text-emerald-600 dark:text-emerald-300', labelKey: 'home.tags.stickySession' },
  { icon: 'chart', iconClass: 'text-amber-600 dark:text-amber-300', labelKey: 'home.tags.realtimeBilling' }
] as const

const providerLogos = [
  { model: 'gpt-4.1', labelKey: 'home.providers.items.openai' },
  { model: 'claude-3-5-sonnet', labelKey: 'home.providers.items.claude' },
  { model: 'gemini-2.5-pro', labelKey: 'home.providers.items.gemini' },
  { model: 'grok-4', labelKey: 'home.providers.items.xai' },
  { model: 'llama-4', labelKey: 'home.providers.items.meta' },
  { model: 'mistral-large', labelKey: 'home.providers.items.mistral' },
  { model: 'command-r-plus', labelKey: 'home.providers.items.cohere' },
  { model: 'midjourney', labelKey: 'home.providers.items.midjourney' },
  { model: 'perplexity', labelKey: 'home.providers.items.perplexity' },
  { model: 'openrouter', labelKey: 'home.providers.items.openrouter' },
  { model: 'deepseek-chat', labelKey: 'home.providers.items.deepseek', domestic: true },
  { model: 'qwen-max', labelKey: 'home.providers.items.qwen', domestic: true },
  { model: 'doubao-pro', labelKey: 'home.providers.items.doubao', domestic: true },
  { model: 'glm-4-plus', labelKey: 'home.providers.items.zhipu', domestic: true },
  { model: 'kimi-k2', labelKey: 'home.providers.items.moonshot', domestic: true },
  { model: 'ernie-4.5', labelKey: 'home.providers.items.baidu', domestic: true },
  { model: 'hunyuan', labelKey: 'home.providers.items.tencent', domestic: true },
  { model: 'spark-max', labelKey: 'home.providers.items.iflytek', domestic: true },
  { model: 'minimax-abab', labelKey: 'home.providers.items.minimax', domestic: true },
  { model: '360gpt', labelKey: 'home.providers.items.ai360', domestic: true }
] as const

const domesticProviderLogos = providerLogos.filter(
  (provider) => 'domestic' in provider && provider.domestic
)

const featureCards = [
  {
    icon: 'server',
    iconClass: 'bg-sky-50 text-sky-600 dark:bg-sky-950/40 dark:text-sky-300',
    titleKey: 'home.highlights.gateway.title',
    descKey: 'home.highlights.gateway.desc'
  },
  {
    icon: 'grid',
    iconClass: 'bg-violet-50 text-violet-600 dark:bg-violet-950/40 dark:text-violet-300',
    titleKey: 'home.highlights.providers.title',
    descKey: 'home.highlights.providers.desc'
  },
  {
    icon: 'calculator',
    iconClass: 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40 dark:text-emerald-300',
    titleKey: 'home.highlights.billing.title',
    descKey: 'home.highlights.billing.desc'
  },
  {
    icon: 'globe',
    iconClass: 'bg-rose-50 text-rose-600 dark:bg-rose-950/40 dark:text-rose-300',
    titleKey: 'home.highlights.languages.title',
    descKey: 'home.highlights.languages.desc'
  }
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

async function copyEndpoint() {
  try {
    await navigator.clipboard.writeText(chatEndpoint.value)
    endpointCopied.value = true
    window.setTimeout(() => {
      endpointCopied.value = false
    }, 1800)
  } catch {
    endpointCopied.value = false
  }
}

onMounted(() => {
  initTheme()
  initNotice()
  apiBaseUrl.value = window.location.origin

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
