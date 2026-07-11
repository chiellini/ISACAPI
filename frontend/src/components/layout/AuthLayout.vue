<template>
  <div
    class="relative min-h-screen overflow-hidden bg-slate-50 px-4 py-8 text-slate-950 dark:bg-dark-950 dark:text-white sm:px-6 lg:px-8"
  >
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div
        class="absolute inset-0 bg-[linear-gradient(135deg,rgba(14,165,233,0.14)_0%,transparent_34%),linear-gradient(225deg,rgba(16,185,129,0.12)_0%,transparent_38%),linear-gradient(180deg,#f8fafc_0%,#edf2f7_100%)] dark:bg-[linear-gradient(135deg,rgba(14,165,233,0.16)_0%,transparent_34%),linear-gradient(225deg,rgba(16,185,129,0.10)_0%,transparent_38%),linear-gradient(180deg,#020617_0%,#111827_100%)]"
      ></div>
      <div
        class="absolute inset-0 bg-[linear-gradient(rgba(15,23,42,0.05)_1px,transparent_1px),linear-gradient(90deg,rgba(15,23,42,0.05)_1px,transparent_1px)] bg-[size:48px_48px] opacity-70 dark:bg-[linear-gradient(rgba(148,163,184,0.08)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.08)_1px,transparent_1px)]"
      ></div>
      <div
        class="absolute -left-28 top-16 h-[520px] w-[58vw] rotate-[-8deg] border-y border-sky-200/60 bg-white/35 dark:border-sky-900/40 dark:bg-white/[0.03]"
      ></div>
      <div
        class="absolute bottom-0 right-0 h-64 w-full bg-[linear-gradient(0deg,rgba(15,23,42,0.08),transparent)] dark:bg-[linear-gradient(0deg,rgba(8,47,73,0.24),transparent)]"
      ></div>
    </div>

    <div
      class="relative z-10 mx-auto grid min-h-[calc(100vh-4rem)] w-full max-w-6xl items-center gap-8 lg:grid-cols-[minmax(0,1fr)_440px]"
    >
      <section class="hidden lg:block">
        <div class="max-w-xl">
          <template v-if="settingsLoaded">
            <div class="mb-6 flex items-center">
              <div
                class="inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-lg bg-white/80 shadow-lg shadow-sky-500/10 ring-1 ring-slate-200/80 dark:bg-white/10 dark:ring-white/10"
              >
                <img :src="siteLogo || '/logo.png'" alt="ISACAI" class="h-full w-full object-contain" />
              </div>
            </div>
            <h1 class="text-gradient mb-3 text-4xl font-bold">
              {{ siteName }}
            </h1>
            <p class="max-w-lg text-base leading-7 text-slate-600 dark:text-slate-300">
              {{ siteSubtitle }}
            </p>
            <div
              class="mt-6 max-w-sm overflow-hidden rounded-lg border border-white/70 bg-white/60 p-3 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5"
            >
              <img :src="companyIconUrl" alt="ISACAI" class="mx-auto h-28 w-full object-contain" />
            </div>
          </template>

          <div class="mt-10 grid max-w-lg grid-cols-3 gap-3">
            <div
              v-for="item in authHighlights"
              :key="item.label"
              class="rounded-lg border border-white/70 bg-white/60 p-4 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5"
            >
              <p class="text-xs text-slate-500 dark:text-slate-400">
                {{ item.label }}
              </p>
              <p class="mt-2 text-sm font-semibold text-slate-900 dark:text-white">
                {{ item.value }}
              </p>
            </div>
          </div>

          <div
            class="mt-8 max-w-lg border-l-4 border-sky-500 bg-white/50 px-5 py-4 text-sm leading-6 text-slate-600 backdrop-blur dark:bg-white/[0.04] dark:text-slate-300"
          >
            支持多模型供应商接入，提供统一 API 管理、用量统计和软件开发交流支持。
          </div>
        </div>
      </section>

      <div class="mx-auto w-full max-w-md">
        <div class="mb-8 text-center lg:hidden">
          <template v-if="settingsLoaded">
            <div class="mb-4 flex justify-center">
              <div
                class="inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-lg bg-white/80 shadow-lg shadow-sky-500/10 ring-1 ring-slate-200/80 dark:bg-white/10 dark:ring-white/10"
              >
                <img :src="siteLogo || '/logo.png'" alt="ISACAI" class="h-full w-full object-contain" />
              </div>
            </div>
            <h1 class="text-gradient mb-2 text-3xl font-bold">
              {{ siteName }}
            </h1>
            <p class="text-sm text-slate-500 dark:text-slate-400">
              {{ siteSubtitle }}
            </p>
            <img :src="companyIconUrl" alt="ISACAI" class="mx-auto mt-4 h-20 w-20 rounded-lg object-contain" />
          </template>
        </div>

        <div class="card-glass rounded-lg p-8 shadow-glass">
          <slot />
        </div>

        <div class="mt-6 text-center text-sm">
          <slot name="footer" />
        </div>

        <div
          class="mt-5 flex items-center justify-center gap-3 rounded-lg border border-slate-200/80 bg-white/75 px-4 py-3 text-center text-xs text-slate-600 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/[0.06] dark:text-slate-300"
        >
          <img :src="companyIconUrl" alt="ISACAI" class="h-8 w-8 shrink-0 rounded-md object-contain" />
          <span>
            <span>软件开发交流 QQ 群：</span>
            <span class="font-semibold text-slate-900 dark:text-white">{{ qqGroup }}</span>
          </span>
        </div>

        <div
          class="mt-5 space-y-2 text-center text-xs text-slate-500 dark:text-slate-400"
          dir="ltr"
        >
          <p>&copy; 2026 ISACAI. All rights reserved. 软件开发交流联系方式：{{ qqGroup }}</p>
          <a
            href="https://beian.miit.gov.cn"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex text-slate-500 transition-colors hover:text-primary-600 dark:text-slate-400 dark:hover:text-primary-300"
          >
            粤ICP备2026050877号
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'ISACAI')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Service Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)
const companyIconUrl = '/logo.png'

const qqGroup = '1027890648'
const authHighlights = [
  { label: '统一接口', value: 'OpenAI Compatible' },
  { label: '模型供应商', value: '30+' },
  { label: '开发交流', value: `QQ ${qqGroup}` },
]

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.text-gradient {
  @apply bg-gradient-to-r from-primary-600 to-primary-500 bg-clip-text text-transparent;
}
</style>
