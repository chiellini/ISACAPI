<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
      <section class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="grid gap-6 p-6 lg:grid-cols-[1.25fr_0.75fr] lg:p-8">
          <div class="space-y-5">
            <div class="inline-flex items-center gap-2 rounded-full border border-primary-200 bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-700 dark:border-primary-900/60 dark:bg-primary-900/20 dark:text-primary-300">
              <Icon name="terminal" size="sm" />
              {{ t('codexGuide.heroBadge') }}
            </div>

            <div>
              <h1 class="text-2xl font-bold tracking-normal text-gray-900 dark:text-white sm:text-3xl">
                {{ t('codexGuide.title') }}
              </h1>
              <p class="mt-3 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-300">
                {{ t('codexGuide.description') }}
              </p>
            </div>

            <div class="flex flex-wrap gap-3">
              <RouterLink to="/keys" class="btn btn-primary">
                <Icon name="key" size="sm" />
                {{ t('codexGuide.actions.openKeys') }}
              </RouterLink>
              <a
                :href="CODEX_DOCS_LINKS.cli"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-secondary"
              >
                <Icon name="externalLink" size="sm" />
                {{ t('codexGuide.actions.officialDocs') }}
              </a>
            </div>
          </div>

          <aside class="rounded-xl border border-primary-200 bg-primary-50/70 p-5 dark:border-primary-900/60 dark:bg-primary-900/15">
            <h2 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('codexGuide.quickStart.title') }}
            </h2>
            <ol class="mt-4 space-y-3">
              <li
                v-for="(item, index) in quickStartItems"
                :key="item"
                class="flex items-center gap-3 text-sm text-gray-700 dark:text-gray-200"
              >
                <span class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-white text-xs font-bold text-primary-700 shadow-sm dark:bg-dark-800 dark:text-primary-300">
                  {{ index + 1 }}
                </span>
                <span>{{ item }}</span>
              </li>
            </ol>
          </aside>
        </div>
      </section>

      <section class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <GuideHeading
          :eyebrow="t('codexGuide.install.eyebrow')"
          :title="t('codexGuide.install.title')"
          :body="t('codexGuide.install.body')"
          icon="download"
          tone="blue"
        />

        <div class="mt-5 grid gap-4 lg:grid-cols-2">
          <article class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-900/60">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('codexGuide.install.prerequisiteTitle') }}
            </h3>
            <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ t('codexGuide.install.prerequisiteBody') }}
            </p>
            <CommandBlock class="mt-4" :command="nodeCheckCommand" />
            <a
              href="https://nodejs.org/en/download"
              target="_blank"
              rel="noopener noreferrer"
              class="mt-3 inline-flex items-center gap-1 text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-300"
            >
              Node.js
              <Icon name="externalLink" size="xs" />
            </a>
          </article>

          <article class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-900/60">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('codexGuide.install.cliTitle') }}
            </h3>
            <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ t('codexGuide.install.cliBody') }}
            </p>
            <CommandBlock class="mt-4" :command="cliInstallCommand" />
          </article>
        </div>

        <article class="mt-4 rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-900/60">
          <div class="flex items-start gap-3">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-violet-50 text-violet-600 dark:bg-violet-900/20 dark:text-violet-300">
              <Icon name="terminal" size="md" />
            </div>
            <div class="min-w-0">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('codexGuide.install.ideTitle') }}
              </h3>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                {{ t('codexGuide.install.ideBody') }}
              </p>
            </div>
          </div>
          <ul class="mt-4 grid gap-2 text-sm text-gray-600 dark:text-gray-300 md:grid-cols-3">
            <li v-for="item in idePoints" :key="item" class="flex gap-2">
              <Icon name="checkCircle" size="sm" class="mt-0.5 flex-shrink-0 text-emerald-500" />
              <span>{{ item }}</span>
            </li>
          </ul>
        </article>
      </section>

      <section class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <GuideHeading
          :eyebrow="t('codexGuide.configure.eyebrow')"
          :title="t('codexGuide.configure.title')"
          :body="t('codexGuide.configure.body')"
          icon="key"
          tone="emerald"
        />

        <div class="mt-5 rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-900/60 dark:bg-amber-900/20">
          <div class="flex items-start gap-3">
            <Icon name="exclamationTriangle" size="lg" class="mt-0.5 flex-shrink-0 text-amber-600 dark:text-amber-300" />
            <div>
              <h3 class="text-sm font-semibold text-amber-900 dark:text-amber-100">
                {{ t('codexGuide.configure.warningTitle') }}
              </h3>
              <p class="mt-1 text-sm leading-6 text-amber-800 dark:text-amber-100/80">
                {{ t('codexGuide.configure.warningBody') }}
              </p>
            </div>
          </div>
        </div>

        <div class="mt-4 grid gap-4 lg:grid-cols-3">
          <article
            v-for="(step, index) in configureSteps"
            :key="step.title"
            class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-900/60"
          >
            <div class="flex items-start gap-3">
              <span class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-white text-xs font-bold text-gray-700 shadow-sm dark:bg-dark-800 dark:text-gray-200">
                {{ index + 1 }}
              </span>
              <div>
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ step.title }}
                </h3>
                <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                  {{ step.body }}
                </p>
              </div>
            </div>
          </article>
        </div>

        <div class="mt-4 grid gap-4 lg:grid-cols-[1fr_1.1fr]">
          <div class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-900/60">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('codexGuide.configure.filesTitle') }}
            </h3>
            <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ t('codexGuide.configure.filesBody') }}
            </p>
            <RouterLink
              to="/cc-switch"
              class="mt-4 inline-flex items-center gap-2 text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-300"
            >
              {{ t('codexGuide.configure.ccSwitchAction') }}
              <Icon name="chevronRight" size="sm" />
            </RouterLink>
          </div>
          <CommandBlock :command="configFiles" />
        </div>
      </section>

      <section class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <GuideHeading
          :eyebrow="t('codexGuide.use.eyebrow')"
          :title="t('codexGuide.use.title')"
          :body="t('codexGuide.use.body')"
          icon="play"
          tone="violet"
        />

        <div class="mt-5 grid gap-4 lg:grid-cols-2">
          <article
            v-for="item in usageCommands"
            :key="item.command"
            class="overflow-hidden rounded-xl border border-gray-200 bg-gray-50 dark:border-dark-600 dark:bg-dark-900/60"
          >
            <div class="p-4">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ item.title }}
              </h3>
              <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-300">
                {{ item.body }}
              </p>
            </div>
            <div class="border-t border-gray-200 bg-gray-950 dark:border-dark-600">
              <div class="flex items-center justify-between gap-3 border-b border-white/10 px-4 py-2">
                <span class="text-xs font-medium text-gray-400">terminal</span>
                <button
                  type="button"
                  class="inline-flex items-center gap-1.5 text-xs font-medium text-gray-300 transition-colors hover:text-white"
                  @click="copyCommand(item.command)"
                >
                  <Icon name="copy" size="xs" />
                  {{ t('codexGuide.use.copy') }}
                </button>
              </div>
              <pre class="overflow-x-auto p-4 text-xs leading-5 text-gray-100"><code>{{ item.command }}</code></pre>
            </div>
          </article>
        </div>
      </section>

      <section class="grid gap-4 lg:grid-cols-[0.85fr_1.15fr]">
        <article class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('codexGuide.bestPractices.title') }}
          </h2>
          <ul class="mt-4 space-y-3 text-sm text-gray-600 dark:text-gray-300">
            <li v-for="item in bestPractices" :key="item" class="flex gap-2">
              <Icon name="checkCircle" size="sm" class="mt-0.5 flex-shrink-0 text-emerald-500" />
              <span class="leading-6">{{ item }}</span>
            </li>
          </ul>
        </article>

        <article class="rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('codexGuide.troubleshooting.title') }}
          </h2>
          <div class="mt-4 grid gap-3 md:grid-cols-2">
            <div
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
            </div>
          </div>
        </article>
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
import { useClipboard } from '@/composables/useClipboard'

interface GuideStep {
  title: string
  body: string
}

interface UsageCommand extends GuideStep {
  command: string
}

const CODEX_DOCS_LINKS = {
  cli: 'https://learn.chatgpt.com/docs/codex/cli',
  ide: 'https://learn.chatgpt.com/docs/codex/ide',
  configuration: 'https://learn.chatgpt.com/docs/config-file/config-basic'
} as const

const { t, tm, rt } = useI18n()
const { copyToClipboard } = useClipboard()

function asText(value: unknown): string {
  return typeof value === 'string' ? value : rt(value as never)
}

function list(path: string): string[] {
  const value = tm(path) as unknown
  return Array.isArray(value) ? value.map((item) => asText(item)) : []
}

function blocks(path: string): GuideStep[] {
  const value = tm(path) as unknown
  if (!Array.isArray(value)) return []
  return value.map((item) => {
    const block = item as Record<string, unknown>
    return {
      title: asText(block.title),
      body: asText(block.body)
    }
  })
}

const quickStartItems = computed(() => list('codexGuide.quickStart.items'))
const idePoints = computed(() => list('codexGuide.install.idePoints'))
const configureSteps = computed(() => blocks('codexGuide.configure.steps'))
const bestPractices = computed(() => list('codexGuide.bestPractices.items'))
const troubleshooting = computed(() => blocks('codexGuide.troubleshooting.items'))

const nodeCheckCommand = `node -v
npm -v`

const cliInstallCommand = `npm install -g @openai/codex
codex --version`

const configFiles = `# macOS / Linux
~/.codex/config.toml
~/.codex/auth.json

# Windows
%USERPROFILE%\\.codex\\config.toml
%USERPROFILE%\\.codex\\auth.json`

const commandValues = [
  'cd /path/to/project\ncodex',
  'codex resume --last',
  'codex exec "Review the current changes and run relevant tests"',
  'codex --help'
]

const usageCommands = computed<UsageCommand[]>(() =>
  blocks('codexGuide.use.commands').map((item, index) => ({
    ...item,
    command: commandValues[index] ?? ''
  }))
)

async function copyCommand(command: string): Promise<void> {
  await copyToClipboard(command, t('codexGuide.use.copied'))
}

const GuideHeading = defineComponent({
  name: 'GuideHeading',
  props: {
    eyebrow: { type: String, required: true },
    title: { type: String, required: true },
    body: { type: String, required: true },
    icon: { type: String, required: true },
    tone: { type: String, required: true }
  },
  setup(props) {
    const toneClasses: Record<string, string> = {
      blue: 'bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-300',
      emerald: 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/20 dark:text-emerald-300',
      violet: 'bg-violet-50 text-violet-600 dark:bg-violet-900/20 dark:text-violet-300'
    }

    return () =>
      h('div', { class: 'flex items-start gap-3' }, [
        h(
          'div',
          {
            class: `flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg ${toneClasses[props.tone] ?? toneClasses.blue}`
          },
          [h(Icon, { name: props.icon as never, size: 'md' })]
        ),
        h('div', [
          h('p', { class: 'text-xs font-semibold uppercase text-gray-500 dark:text-gray-400' }, props.eyebrow),
          h('h2', { class: 'mt-1 text-lg font-semibold text-gray-900 dark:text-white' }, props.title),
          h('p', { class: 'mt-2 max-w-4xl text-sm leading-6 text-gray-600 dark:text-gray-300' }, props.body)
        ])
      ])
  }
})

const CommandBlock = defineComponent({
  name: 'CommandBlock',
  props: {
    command: { type: String, required: true }
  },
  setup(props, { attrs }) {
    return () =>
      h(
        'div',
        {
          ...attrs,
          class: `overflow-hidden rounded-lg border border-gray-800 bg-gray-950 ${String(attrs.class ?? '')}`
        },
        [
          h('div', { class: 'flex items-center justify-between gap-3 border-b border-white/10 px-3 py-2' }, [
            h('span', { class: 'text-xs font-medium text-gray-400' }, 'terminal'),
            h(
              'button',
              {
                type: 'button',
                class: 'inline-flex items-center gap-1.5 text-xs font-medium text-gray-300 transition-colors hover:text-white',
                onClick: () => copyCommand(props.command)
              },
              [
                h(Icon, { name: 'copy', size: 'xs' }),
                h('span', t('codexGuide.use.copy'))
              ]
            )
          ]),
          h('pre', { class: 'overflow-x-auto p-3 text-xs leading-5 text-gray-100' }, [
            h('code', props.command)
          ])
        ]
      )
  }
})
</script>
