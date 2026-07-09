<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <section class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="grid gap-6 p-6 lg:grid-cols-[1.35fr_0.65fr] lg:p-8">
          <div class="space-y-5">
            <div class="inline-flex items-center gap-2 rounded-full border border-primary-200 bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-700 dark:border-primary-900/60 dark:bg-primary-900/20 dark:text-primary-300">
              <Icon name="book" size="sm" />
              {{ t('ccSwitchGuide.heroBadge') }}
            </div>
            <div>
              <h1 class="text-2xl font-bold tracking-normal text-gray-900 dark:text-white sm:text-3xl">
                {{ t('ccSwitchGuide.title') }}
              </h1>
              <p class="mt-3 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-300">
                {{ t('ccSwitchGuide.description') }}
              </p>
            </div>
            <div class="flex flex-wrap gap-3">
              <RouterLink to="/keys" class="btn btn-primary">
                <Icon name="key" size="sm" />
                {{ t('ccSwitchGuide.actions.openKeys') }}
              </RouterLink>
              <a
                :href="CC_SWITCH_DOWNLOAD_LINKS.releases"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-secondary"
              >
                <Icon name="externalLink" size="sm" />
                {{ t('ccSwitchGuide.actions.releasePage') }}
              </a>
            </div>
          </div>

          <div class="rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-900/60 dark:bg-amber-900/20">
            <div class="flex items-start gap-3">
              <Icon name="exclamationTriangle" size="lg" class="mt-0.5 flex-shrink-0 text-amber-600 dark:text-amber-300" />
              <div>
                <h2 class="text-sm font-semibold text-amber-900 dark:text-amber-100">
                  {{ t('ccSwitchGuide.remoteWarning.title') }}
                </h2>
                <p class="mt-2 text-sm leading-6 text-amber-800 dark:text-amber-100/80">
                  {{ t('ccSwitchGuide.remoteWarning.body') }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="grid gap-4 lg:grid-cols-2">
        <article class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-start gap-3">
            <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-300">
              <Icon name="download" size="md" />
            </div>
            <div>
              <p class="text-xs font-semibold uppercase text-blue-600 dark:text-blue-300">
                {{ t('ccSwitchGuide.scenarios.local.badge') }}
              </p>
              <h2 class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('ccSwitchGuide.scenarios.local.title') }}
              </h2>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                {{ t('ccSwitchGuide.scenarios.local.body') }}
              </p>
            </div>
          </div>
          <ul class="mt-4 space-y-2 text-sm text-gray-600 dark:text-gray-300">
            <li v-for="item in localPoints" :key="item" class="flex gap-2">
              <Icon name="checkCircle" size="sm" class="mt-0.5 flex-shrink-0 text-emerald-500" />
              <span>{{ item }}</span>
            </li>
          </ul>
        </article>

        <article class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-start gap-3">
            <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-violet-50 text-violet-600 dark:bg-violet-900/20 dark:text-violet-300">
              <Icon name="server" size="md" />
            </div>
            <div>
              <p class="text-xs font-semibold uppercase text-violet-600 dark:text-violet-300">
                {{ t('ccSwitchGuide.scenarios.remote.badge') }}
              </p>
              <h2 class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('ccSwitchGuide.scenarios.remote.title') }}
              </h2>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                {{ t('ccSwitchGuide.scenarios.remote.body') }}
              </p>
            </div>
          </div>
          <ul class="mt-4 space-y-2 text-sm text-gray-600 dark:text-gray-300">
            <li v-for="item in remotePoints" :key="item" class="flex gap-2">
              <Icon name="checkCircle" size="sm" class="mt-0.5 flex-shrink-0 text-emerald-500" />
              <span>{{ item }}</span>
            </li>
          </ul>
        </article>
      </section>

      <section class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-col justify-between gap-4 md:flex-row md:items-start">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('ccSwitchGuide.download.title') }}
            </h2>
            <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ t('ccSwitchGuide.download.body') }}
            </p>
          </div>
          <div class="grid w-full gap-2 sm:grid-cols-2 md:w-auto md:grid-cols-4">
            <a
              v-for="link in downloadLinks"
              :key="link.label"
              :href="link.href"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center justify-center gap-2 rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm font-medium text-gray-700 transition-colors hover:border-primary-300 hover:bg-primary-50 hover:text-primary-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:border-primary-700 dark:hover:bg-primary-900/20 dark:hover:text-primary-200"
            >
              <Icon :name="link.icon" size="sm" />
              {{ link.label }}
            </a>
          </div>
        </div>
      </section>

      <GuideSection
        :eyebrow="t('ccSwitchGuide.local.eyebrow')"
        :title="t('ccSwitchGuide.local.title')"
        :body="t('ccSwitchGuide.local.body')"
        :steps="localSteps"
        icon="download"
        tone="blue"
      />

      <GuideSection
        :eyebrow="t('ccSwitchGuide.remote.eyebrow')"
        :title="t('ccSwitchGuide.remote.title')"
        :body="t('ccSwitchGuide.remote.body')"
        :steps="remoteSteps"
        icon="server"
        tone="violet"
      >
        <template #extra>
          <div class="grid gap-4 lg:grid-cols-2">
            <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-900/60">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('ccSwitchGuide.remote.filesTitle') }}
              </h3>
              <ul class="mt-3 space-y-2 text-sm text-gray-600 dark:text-gray-300">
                <li v-for="item in remoteFiles" :key="item" class="flex gap-2">
                  <Icon name="document" size="sm" class="mt-0.5 flex-shrink-0 text-gray-400" />
                  <span>{{ item }}</span>
                </li>
              </ul>
            </div>
            <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-900/60">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('ccSwitchGuide.remote.verifyTitle') }}
              </h3>
              <pre class="mt-3 overflow-x-auto rounded-lg bg-gray-950 p-3 text-xs leading-5 text-gray-100"><code>{{ remoteVerifyCommand }}</code></pre>
            </div>
          </div>
        </template>
      </GuideSection>

      <section class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('ccSwitchGuide.troubleshooting.title') }}
        </h2>
        <div class="mt-4 grid gap-3 lg:grid-cols-2">
          <article
            v-for="item in troubleshooting"
            :key="item.title"
            class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-900/60"
          >
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ item.title }}
            </h3>
            <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ item.body }}
            </p>
          </article>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { CC_SWITCH_DOWNLOAD_LINKS } from '@/utils/ccswitchImport'

interface GuideStep {
  title: string
  body: string
  bullets: string[]
}

interface InfoBlock {
  title: string
  body: string
}

const { t, tm, rt } = useI18n()

function asText(value: unknown): string {
  return typeof value === 'string' ? value : rt(value as never)
}

function list(path: string): string[] {
  const value = tm(path) as unknown
  return Array.isArray(value) ? value.map((item) => asText(item)) : []
}

function blocks(path: string): InfoBlock[] {
  const value = tm(path) as unknown
  if (!Array.isArray(value)) return []
  return value.map((item) => {
    const block = item as Record<string, unknown>
    return {
      title: asText(block.title),
      body: asText(block.body),
    }
  })
}

function steps(path: string): GuideStep[] {
  const value = tm(path) as unknown
  if (!Array.isArray(value)) return []
  return value.map((item) => {
    const step = item as Record<string, unknown>
    return {
      title: asText(step.title),
      body: asText(step.body),
      bullets: Array.isArray(step.bullets) ? step.bullets.map((bullet) => asText(bullet)) : [],
    }
  })
}

const localPoints = computed(() => list('ccSwitchGuide.scenarios.local.points'))
const remotePoints = computed(() => list('ccSwitchGuide.scenarios.remote.points'))
const localSteps = computed(() => steps('ccSwitchGuide.local.steps'))
const remoteSteps = computed(() => steps('ccSwitchGuide.remote.steps'))
const remoteFiles = computed(() => list('ccSwitchGuide.remote.files'))
const troubleshooting = computed(() => blocks('ccSwitchGuide.troubleshooting.items'))

const downloadLinks = computed(() => [
  { label: t('ccSwitchGuide.download.officialSite'), href: CC_SWITCH_DOWNLOAD_LINKS.officialSite, icon: 'globe' as const },
  { label: t('ccSwitchGuide.download.windows'), href: CC_SWITCH_DOWNLOAD_LINKS.windows, icon: 'download' as const },
  { label: t('ccSwitchGuide.download.macos'), href: CC_SWITCH_DOWNLOAD_LINKS.macos, icon: 'download' as const },
  { label: t('ccSwitchGuide.download.release'), href: CC_SWITCH_DOWNLOAD_LINKS.releases, icon: 'externalLink' as const },
])

const remoteVerifyCommand = `# Linux / macOS remote terminal
echo "$ANTHROPIC_BASE_URL"
echo "$OPENAI_BASE_URL"
cat ~/.claude/settings.json 2>/dev/null
cat ~/.codex/config.toml 2>/dev/null`

const GuideSection = defineComponent({
  name: 'GuideSection',
  props: {
    eyebrow: { type: String, required: true },
    title: { type: String, required: true },
    body: { type: String, required: true },
    steps: { type: Array as () => GuideStep[], required: true },
    icon: { type: String, required: true },
    tone: { type: String, required: true },
  },
  setup(props, { slots }) {
    const toneClass = props.tone === 'violet'
      ? 'bg-violet-50 text-violet-600 dark:bg-violet-900/20 dark:text-violet-300'
      : 'bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-300'

    return () => h('section', { class: 'rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800' }, [
      h('div', { class: 'flex items-start gap-3' }, [
        h('div', { class: `flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg ${toneClass}` }, [
          h(Icon, { name: props.icon as never, size: 'md' }),
        ]),
        h('div', [
          h('p', { class: 'text-xs font-semibold uppercase text-gray-500 dark:text-gray-400' }, props.eyebrow),
          h('h2', { class: 'mt-1 text-lg font-semibold text-gray-900 dark:text-white' }, props.title),
          h('p', { class: 'mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300' }, props.body),
        ]),
      ]),
      h('div', { class: 'mt-5 grid gap-4 lg:grid-cols-2' }, props.steps.map((step, index) =>
        h('article', { class: 'rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-900/60' }, [
          h('div', { class: 'flex items-start gap-3' }, [
            h('div', { class: 'flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-white text-sm font-semibold text-gray-700 shadow-sm dark:bg-dark-800 dark:text-gray-200' }, String(index + 1)),
            h('div', { class: 'min-w-0' }, [
              h('h3', { class: 'text-sm font-semibold text-gray-900 dark:text-white' }, step.title),
              h('p', { class: 'mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300' }, step.body),
              step.bullets.length > 0
                ? h('ul', { class: 'mt-3 space-y-2 text-sm text-gray-600 dark:text-gray-300' }, step.bullets.map((bullet) =>
                    h('li', { class: 'flex gap-2' }, [
                      h(Icon, { name: 'checkCircle', size: 'sm', class: 'mt-0.5 flex-shrink-0 text-emerald-500' }),
                      h('span', bullet),
                    ])
                  ))
                : null,
            ]),
          ]),
        ])
      )),
      slots.extra ? h('div', { class: 'mt-5' }, slots.extra()) : null,
    ])
  },
})
</script>
