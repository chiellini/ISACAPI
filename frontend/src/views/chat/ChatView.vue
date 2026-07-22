<template>
  <AppLayout>
    <div class="relative flex h-[calc(100vh-8rem)] gap-3">
      <!-- 移动端遮罩 -->
      <div v-if="showHistory" class="fixed inset-0 z-30 bg-black/40 sm:hidden" @click="showHistory = false"></div>

      <!-- 历史侧栏 -->
      <aside
        class="flex-col rounded-xl bg-gray-50 p-2 dark:bg-dark-800 sm:static sm:flex sm:w-56"
        :class="showHistory ? 'fixed inset-y-0 left-0 z-40 flex w-64' : 'hidden'"
      >
        <button class="btn btn-primary mb-2 w-full" :disabled="streaming" @click="newSession">
          {{ t('chat.newChat') }}
        </button>
        <div class="flex-1 space-y-1 overflow-y-auto">
          <div
            v-for="s in sessions"
            :key="s.id"
            class="group flex cursor-pointer items-center gap-1 rounded-lg px-2 py-1.5 text-sm"
            :class="s.id === currentId
              ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/40 dark:text-primary-300'
              : 'hover:bg-gray-100 dark:hover:bg-dark-700'"
            @click="switchSession(s.id)"
          >
            <span class="min-w-0 flex-1 truncate">{{ s.title || t('chat.untitled') }}</span>
            <button class="opacity-0 group-hover:opacity-100" :title="t('chat.rename')" @click.stop="renameSession(s.id)">✎</button>
            <button class="opacity-0 group-hover:opacity-100" :title="t('chat.delete')" @click.stop="deleteSession(s.id)">🗑</button>
          </div>
        </div>
      </aside>

      <!-- 主区 -->
      <div class="flex min-w-0 flex-1 flex-col">
        <!-- Header -->
        <div class="mb-3 flex items-center gap-3">
          <button class="btn btn-secondary sm:hidden" :title="t('chat.history')" @click="showHistory = true">☰</button>
          <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('chat.model') }}</span>
          <select v-model="selectedModel" class="input w-44 sm:w-56" :disabled="streaming">
            <option v-for="m in models" :key="m.id" :value="m.id">{{ m.label }}</option>
          </select>
          <!-- 联网搜索开关：仅文本模型 + 平台已配置搜索时出现 -->
          <button
            v-if="canWebSearch"
            type="button"
            class="flex h-9 items-center gap-1 rounded-lg border px-3 text-sm transition"
            :class="webSearchOn
              ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-500 dark:bg-primary-900/30 dark:text-primary-300'
              : 'border-gray-200 text-gray-500 hover:bg-gray-100 dark:border-dark-600 dark:hover:bg-dark-700'"
            :disabled="streaming"
            :title="t('chat.webSearchHint')"
            :aria-pressed="webSearchOn"
            @click="webSearchOn = !webSearchOn"
          >
            <span>🌐</span><span class="hidden sm:inline">{{ t('chat.webSearch') }}</span>
          </button>
        </div>

        <!-- 消息列表 -->
        <div ref="listEl" class="flex-1 space-y-4 overflow-y-auto rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
          <div v-if="messages.length === 0" class="flex h-full items-center justify-center text-gray-400">
            {{ t('chat.empty') }}
          </div>
          <div
            v-for="(msg, i) in messages"
            :key="i"
            class="flex flex-col"
            :class="msg.role === 'user' ? 'items-end' : 'items-start'"
          >
            <div
              class="max-w-[80%] break-words rounded-2xl px-4 py-2 text-sm"
              :class="msg.role === 'user'
                ? 'whitespace-pre-wrap bg-primary-600 text-white'
                : 'bg-white text-gray-800 shadow-sm dark:bg-dark-700 dark:text-gray-100'"
            >
              <!-- 生成的图片：点击可全屏预览 -->
              <div v-if="msg.images?.length" class="flex flex-wrap gap-2">
                <img
                  v-for="(src, j) in msg.images"
                  :key="j"
                  :src="src"
                  class="max-h-64 cursor-zoom-in rounded-lg transition hover:opacity-90"
                  :alt="t('chat.previewAlt')"
                  @click="openPreview(src)"
                />
              </div>
              <!-- 助手文本：Markdown 渲染（点击其中图片也可全屏） -->
              <template v-else-if="msg.role === 'assistant'">
                <!-- 联网搜索进行中的状态提示 -->
                <div
                  v-if="msg.searchStatus"
                  class="mb-1 flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400"
                >
                  <span class="animate-pulse">🔍</span><span class="truncate">{{ msg.searchStatus }}</span>
                </div>
                <div
                  class="markdown-body"
                  v-html="msg.content ? renderMarkdown(msg.content) : (streaming && !msg.searchStatus ? '…' : '')"
                  @click="onMarkdownClick"
                ></div>
                <!-- 搜索来源引用 -->
                <div
                  v-if="msg.sources?.length"
                  class="mt-2 border-t border-gray-200 pt-1.5 dark:border-dark-600"
                >
                  <div class="mb-1 text-xs font-medium text-gray-500 dark:text-gray-400">
                    {{ t('chat.sources') }}
                  </div>
                  <ol class="space-y-0.5 text-xs">
                    <li v-for="(s, k) in msg.sources" :key="k" class="flex gap-1">
                      <span class="text-gray-400">{{ k + 1 }}.</span>
                      <a
                        :href="s.url"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="truncate text-primary-600 hover:underline dark:text-primary-400"
                        :title="s.title || s.url"
                      >{{ s.title || s.url }}</a>
                    </li>
                  </ol>
                </div>
              </template>
              <!-- 用户文本 + 附件 -->
              <template v-else>
                {{ msg.content }}
                <div v-if="msg.attachments?.length" class="mt-2 flex flex-wrap gap-2">
                  <img
                    v-for="(a, j) in msg.attachments.filter((x) => x.kind === 'image')"
                    :key="'img' + j"
                    :src="a.dataUrl"
                    class="max-h-32 cursor-zoom-in rounded-lg transition hover:opacity-90"
                    alt="attachment"
                    @click="a.dataUrl && openPreview(a.dataUrl)"
                  />
                  <span
                    v-for="(a, j) in msg.attachments.filter((x) => x.kind === 'text')"
                    :key="'txt' + j"
                    class="rounded bg-black/20 px-2 py-0.5 text-xs"
                  >📄 {{ a.name }}</span>
                </div>
              </template>
            </div>
            <!-- 助手操作：复制 / 重新生成 -->
            <div v-if="msg.role === 'assistant' && (msg.content || msg.images?.length)" class="mt-1 flex gap-2 px-1 text-xs text-gray-400">
              <button v-if="msg.content" class="hover:text-gray-600 dark:hover:text-gray-200" @click="copyText(msg.content)">
                {{ t('chat.copy') }}
              </button>
              <button
                v-if="i === messages.length - 1 && !streaming"
                class="hover:text-gray-600 dark:hover:text-gray-200"
                @click="regenerate"
              >
                {{ t('chat.regenerate') }}
              </button>
            </div>
          </div>
        </div>

        <p v-if="errorMsg" class="mt-2 text-sm text-red-500">{{ errorMsg }}</p>

        <!-- 待发送附件 -->
        <div v-if="pending.length" class="mt-2 flex flex-wrap gap-2">
          <span
            v-for="(a, i) in pending"
            :key="i"
            class="flex items-center gap-1 rounded-lg bg-gray-100 px-2 py-1 text-xs dark:bg-dark-700"
          >
            <img v-if="a.kind === 'image'" :src="a.dataUrl" class="h-6 w-6 rounded object-cover" alt="" />
            <span v-else>📄</span>
            <span class="max-w-[8rem] truncate">{{ a.name }}</span>
            <button class="text-gray-400 hover:text-red-500" @click="pending.splice(i, 1)">✕</button>
          </span>
        </div>

        <!-- 输入框 -->
        <div class="mt-3 flex items-end gap-2">
          <button
            v-if="!isImageModel(selectedModel)"
            class="btn btn-secondary h-10"
            :disabled="streaming"
            :title="t('chat.attach')"
            @click="fileInput?.click()"
          >📎</button>
          <input
            ref="fileInput"
            type="file"
            class="hidden"
            multiple
            accept="image/png,image/jpeg,image/webp,image/gif,text/plain,text/markdown,.md,.txt"
            @change="onFiles"
          />
          <textarea
            v-model="input"
            :placeholder="t('chat.placeholder')"
            rows="2"
            class="input flex-1 resize-none"
            @keydown.enter.exact.prevent="send"
          ></textarea>
          <button v-if="!streaming" class="btn btn-primary h-10" :disabled="!canSend" @click="send">
            {{ t('chat.send') }}
          </button>
          <button v-else class="btn btn-secondary h-10" @click="stop">{{ t('chat.stop') }}</button>
        </div>
      </div>
    </div>

    <!-- 全屏图片预览 -->
    <Teleport to="body">
      <div
        v-if="previewSrc"
        class="fixed inset-0 z-[60] flex items-center justify-center bg-black/80 p-4"
        @click="closePreview"
      >
        <img
          :src="previewSrc"
          class="max-h-full max-w-full rounded-lg object-contain shadow-2xl"
          :alt="t('chat.previewAlt')"
          @click.stop
        />
        <button
          class="absolute right-4 top-4 rounded-full bg-white/15 px-3 py-1 text-lg leading-none text-white hover:bg-white/30"
          :title="t('chat.closePreview')"
          @click.stop="closePreview"
        >✕</button>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { onKeyStroke } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { OPENAI_CODEX_DEFAULT_MODEL } from '@/constants/codex'
import { renderMarkdown } from '@/utils/markdown'
import {
  generateImage,
  streamChatCompletion,
  completeChat,
  chatSearch,
  getChatCapabilities,
  listSessions as apiListSessions,
  getSession as apiGetSession,
  createSession as apiCreateSession,
  updateSession as apiUpdateSession,
  deleteSession as apiDeleteSession,
  uploadChatImage,
  fetchChatImageDataUrl,
  type ChatMessage,
  type ChatCompletionMessage,
  type AssistantToolCall,
  type ContentPart,
  type ServerMessage,
  type ChatSource,
} from '@/api/chat'

const { t } = useI18n()

const models: Array<{ id: string; label: string }> = [
  { id: OPENAI_CODEX_DEFAULT_MODEL, label: OPENAI_CODEX_DEFAULT_MODEL },
  { id: 'gpt-image-2', label: 'GPT Image 2' },
]
const selectedModel = ref(models[0].id)

const MAX_IMAGE_BYTES = 10 * 1024 * 1024
const MAX_TEXT_BYTES = 1 * 1024 * 1024

interface Attachment {
  kind: 'image' | 'text'
  name: string
  mime: string
  dataUrl?: string
  text?: string
  // 服务端图片引用（srv:<id>）；上传落库后回填，用于跨设备回读。
  ref?: string
}
interface UiMessage {
  role: 'user' | 'assistant'
  content: string
  images?: string[]      // 展示用（data URL）
  imageRefs?: string[]   // 持久化引用（srv:<id> / 兼容旧的 IndexedDB key）
  attachments?: Attachment[]
  sources?: ChatSource[] // 联网搜索引用来源
  searchStatus?: string  // 搜索进行中的临时状态（不持久化）
}
// 侧栏只保留会话元数据；消息按需从服务端拉取。
interface SessionMeta {
  id: number
  title: string
  model: string
}

interface StoredChatImage {
  id: string
  sessionId: number
  dataUrl: string
  createdAt: number
}

interface StoredAttachment {
  kind: 'image' | 'text'
  name: string
  mime: string
  ref?: string
}

// v2：统一承载「助手生成图（imageRefs）」「用户上传附件（attachments）」与「搜索来源（sources）」。
interface StoredMessagePayload {
  type: typeof CHAT_MESSAGE_PAYLOAD_TYPE
  version: 2
  content?: string
  imageRefs?: string[]
  attachments?: StoredAttachment[]
  sources?: ChatSource[]
}

const CHAT_MESSAGE_PAYLOAD_TYPE = 'isac-chat-message'
// v1（历史遗留，仅读）：{ type, version:1, content, imageRefs(IndexedDB keys) }
const LEGACY_IMAGE_PAYLOAD_TYPE = 'isac-chat-local-images'
// 服务端图片引用前缀，区别于旧的浏览器 IndexedDB key 与外链 / 内联 data URL。
const SERVER_IMAGE_PREFIX = 'srv:'
const CHAT_IMAGE_DB_NAME = 'isac-chat-images'
const CHAT_IMAGE_STORE = 'images'
let chatImageDBPromise: Promise<IDBDatabase> | null = null

const sessions = ref<SessionMeta[]>([])
const currentId = ref(0)
const messages = ref<UiMessage[]>([])
const pending = ref<Attachment[]>([])
const input = ref('')
const streaming = ref(false)
const errorMsg = ref('')
const showHistory = ref(false)
const listEl = ref<HTMLElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const previewSrc = ref('')
let controller: AbortController | null = null

// ───────── 会话记忆（长/中/短期）状态，随会话切换加载、每轮保存 ─────────
// summary=中期滚动摘要；memory=长期稳定事实；summarizedCount=已折叠进摘要的前缀消息条数。
const summary = ref('')
const memory = ref('')
const summarizedCount = ref(0)

// ───────── 联网搜索开关（能力由后端探测，偏好本地持久化） ─────────
const WEB_SEARCH_PREF_KEY = 'isac-chat-web-search'
const webSearchAvailable = ref(false)
const webSearchOn = ref(localStorage.getItem(WEB_SEARCH_PREF_KEY) === '1')
// 仅文本模型 + 平台已配置搜索时，开关才生效。
const canWebSearch = computed(() => webSearchAvailable.value && !isImageModel(selectedModel.value))
watch(webSearchOn, (v) => {
  try {
    localStorage.setItem(WEB_SEARCH_PREF_KEY, v ? '1' : '0')
  } catch {
    /* 忽略持久化失败 */
  }
})

const canSend = computed(
  () => (!!input.value.trim() || pending.value.length > 0) && !!selectedModel.value && !streaming.value,
)

function isImageModel(id: string): boolean {
  return id.startsWith('gpt-image-')
}

function openChatImageDB(): Promise<IDBDatabase> {
  if (chatImageDBPromise) return chatImageDBPromise
  const promise = new Promise<IDBDatabase>((resolve, reject) => {
    if (typeof indexedDB === 'undefined') {
      reject(new Error('IndexedDB unavailable'))
      return
    }
    const req = indexedDB.open(CHAT_IMAGE_DB_NAME, 1)
    req.onupgradeneeded = () => {
      const db = req.result
      if (!db.objectStoreNames.contains(CHAT_IMAGE_STORE)) {
        const store = db.createObjectStore(CHAT_IMAGE_STORE, { keyPath: 'id' })
        store.createIndex('sessionId', 'sessionId', { unique: false })
      }
    }
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error || new Error('open image store failed'))
  }).catch((err) => {
    chatImageDBPromise = null
    throw err
  })
  chatImageDBPromise = promise
  return promise
}

function newLocalImageId(sessionId: number): string {
  const random = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`
  return `chat-image:${sessionId}:${random}`
}

async function putLocalImage(sessionId: number, dataUrl: string): Promise<string> {
  const trimmed = dataUrl.trim()
  if (!trimmed) return ''
  if (!trimmed.startsWith('data:image/')) return trimmed
  const id = newLocalImageId(sessionId)
  const db = await openChatImageDB()
  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(CHAT_IMAGE_STORE, 'readwrite')
    tx.objectStore(CHAT_IMAGE_STORE).put({
      id,
      sessionId,
      dataUrl: trimmed,
      createdAt: Date.now(),
    } satisfies StoredChatImage)
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error || new Error('save image failed'))
    tx.onabort = () => reject(tx.error || new Error('save image aborted'))
  })
  return id
}

async function getLocalImage(ref: string): Promise<string> {
  const trimmed = ref.trim()
  if (!trimmed || trimmed.startsWith('data:image/') || /^https?:\/\//i.test(trimmed)) return trimmed
  const db = await openChatImageDB()
  return new Promise((resolve) => {
    const tx = db.transaction(CHAT_IMAGE_STORE, 'readonly')
    const req = tx.objectStore(CHAT_IMAGE_STORE).get(trimmed)
    req.onsuccess = () => resolve((req.result as StoredChatImage | undefined)?.dataUrl || '')
    req.onerror = () => resolve('')
  })
}

// 生成/上传图片：优先落服务端（跨设备）；失败回落浏览器 IndexedDB，保证不丢图。
async function saveServerImages(sessionId: number, images: string[]): Promise<string[]> {
  const refs: string[] = []
  for (const src of images) {
    const trimmed = (src || '').trim()
    if (!trimmed) continue
    // 外链（http/https）直接存原样，无需转存。
    if (!trimmed.startsWith('data:image/')) {
      refs.push(trimmed)
      continue
    }
    try {
      const id = await uploadChatImage(sessionId, trimmed)
      refs.push(`${SERVER_IMAGE_PREFIX}${id}`)
    } catch {
      try {
        const local = await putLocalImage(sessionId, trimmed)
        if (local) refs.push(local)
      } catch {
        refs.push(trimmed) // 最后兜底：内联（仍可展示，仅体积偏大）
      }
    }
  }
  return refs
}

// 把一个图片引用解析成可直接展示 / 重新发送的 data URL。
async function resolveImageRef(ref: string): Promise<string> {
  const trimmed = (ref || '').trim()
  if (!trimmed) return ''
  if (trimmed.startsWith(SERVER_IMAGE_PREFIX)) {
    try {
      return await fetchChatImageDataUrl(trimmed.slice(SERVER_IMAGE_PREFIX.length))
    } catch {
      return ''
    }
  }
  if (trimmed.startsWith('data:image/') || /^https?:\/\//i.test(trimmed)) return trimmed
  // 兼容旧数据：浏览器 IndexedDB key。
  try {
    return await getLocalImage(trimmed)
  } catch {
    return ''
  }
}

async function loadServerImages(refs: string[] = []): Promise<string[]> {
  const out: string[] = []
  for (const ref of refs) {
    const src = await resolveImageRef(ref)
    if (src) out.push(src)
  }
  return out
}

function normalizeStoredAttachment(value: unknown): StoredAttachment | null {
  if (!value || typeof value !== 'object') return null
  const a = value as Record<string, unknown>
  const kind = a.kind === 'image' ? 'image' : a.kind === 'text' ? 'text' : null
  if (!kind) return null
  return {
    kind,
    name: typeof a.name === 'string' ? a.name : '',
    mime: typeof a.mime === 'string' ? a.mime : '',
    ref: typeof a.ref === 'string' && a.ref.trim() ? a.ref : undefined,
  }
}

function normalizeStoredSource(value: unknown): ChatSource | null {
  if (!value || typeof value !== 'object') return null
  const s = value as Record<string, unknown>
  const url = typeof s.url === 'string' ? s.url.trim() : ''
  if (!url) return null
  return {
    url,
    title: typeof s.title === 'string' ? s.title : '',
    snippet: typeof s.snippet === 'string' ? s.snippet : '',
  }
}

function stringRefs(value: unknown): string[] {
  return Array.isArray(value)
    ? value.filter((r): r is string => typeof r === 'string' && !!r.trim())
    : []
}

// 解析持久化的消息内容：纯文本返回 null，结构化（v1/v2）返回统一后的 v2 形态。
function parseStoredMessage(content: string): StoredMessagePayload | null {
  const trimmed = content.trim()
  if (!trimmed.startsWith('{')) return null
  try {
    const raw = JSON.parse(trimmed) as Record<string, unknown>
    if (raw.type === LEGACY_IMAGE_PAYLOAD_TYPE && raw.version === 1) {
      return {
        type: CHAT_MESSAGE_PAYLOAD_TYPE,
        version: 2,
        content: typeof raw.content === 'string' ? raw.content : '',
        imageRefs: stringRefs(raw.imageRefs),
      }
    }
    if (raw.type === CHAT_MESSAGE_PAYLOAD_TYPE && raw.version === 2) {
      const attachments = Array.isArray(raw.attachments)
        ? raw.attachments.map(normalizeStoredAttachment).filter((a): a is StoredAttachment => !!a)
        : []
      const sources = Array.isArray(raw.sources)
        ? raw.sources.map(normalizeStoredSource).filter((s): s is ChatSource => !!s)
        : []
      return {
        type: CHAT_MESSAGE_PAYLOAD_TYPE,
        version: 2,
        content: typeof raw.content === 'string' ? raw.content : '',
        imageRefs: stringRefs(raw.imageRefs),
        attachments,
        sources,
      }
    }
    return null
  } catch {
    return null
  }
}

function serializeServerContent(message: UiMessage): string {
  const imageRefs = (message.imageRefs ?? []).filter(Boolean)
  const attachments: StoredAttachment[] = (message.attachments ?? [])
    .map((a): StoredAttachment =>
      a.kind === 'image'
        ? { kind: 'image', name: a.name, mime: a.mime, ref: a.ref }
        : { kind: 'text', name: a.name, mime: a.mime },
    )
    // 仅持久化已拿到服务端引用的图片附件；文本附件保留元数据（名称）用于回显。
    .filter((a) => a.kind === 'text' || !!a.ref)
  const sources = (message.sources ?? []).filter((s) => s.url)
  if (!imageRefs.length && !attachments.length && !sources.length) return message.content
  const payload: StoredMessagePayload = {
    type: CHAT_MESSAGE_PAYLOAD_TYPE,
    version: 2,
    content: message.content,
    imageRefs: imageRefs.length ? imageRefs : undefined,
    attachments: attachments.length ? attachments : undefined,
    sources: sources.length ? sources : undefined,
  }
  return JSON.stringify(payload)
}

async function fromServerMessage(message: ServerMessage): Promise<UiMessage> {
  const payload = parseStoredMessage(message.content)
  if (!payload) return { role: message.role, content: message.content }

  const msg: UiMessage = { role: message.role, content: payload.content || '' }

  if (payload.imageRefs?.length) {
    const images = await loadServerImages(payload.imageRefs)
    if (images.length) {
      msg.images = images
      msg.imageRefs = payload.imageRefs
    }
    if (!msg.content && !images.length) msg.content = t('chat.noImage')
  }

  if (payload.attachments?.length) {
    const atts: Attachment[] = []
    for (const a of payload.attachments) {
      if (a.kind === 'image' && a.ref) {
        const src = await resolveImageRef(a.ref)
        atts.push({ kind: 'image', name: a.name, mime: a.mime, dataUrl: src || undefined, ref: a.ref })
      } else if (a.kind === 'text') {
        atts.push({ kind: 'text', name: a.name, mime: a.mime })
      }
    }
    if (atts.length) msg.attachments = atts
  }

  if (payload.sources?.length) msg.sources = payload.sources

  return msg
}

// ───────── 友好错误映射 ─────────

function friendlyError(err: Error): string {
  const raw = err?.message || ''
  let code = ''
  let msg = raw
  try {
    const j = JSON.parse(raw)
    code = j?.error?.code || j?.code || ''
    msg = j?.error?.message || j?.message || raw
  } catch {
    /* 非 JSON，按原文处理 */
  }
  const hay = `${code} ${msg}`.toLowerCase()
  if (hay.includes('insufficient') || hay.includes('balance')) return t('chat.errBalance')
  if (hay.includes('quota')) return t('chat.errQuota')
  if (hay.includes('429') || hay.includes('rate') || hay.includes('too many')) return t('chat.errRate')
  if (hay.includes('504') || hay.includes('gateway time-out') || hay.includes('gateway timeout') || hay.includes('timeout')) return t('chat.errTimeout')
  if (hay.includes('chat_unavailable') || hay.includes('no_available_chat_group')) return t('chat.errUnavailable')
  if (hay.includes('not_found') || hay.includes('not supported') || hay.includes('model')) return t('chat.errModel')
  if (raw.trim().startsWith('<') || hay.includes('<html')) return t('chat.errGeneric')
  return msg || t('chat.errGeneric')
}

// ───────── 历史持久化（服务端保存文本，本地保存生成图） ─────────

// 服务端保存文本 + 图片引用（生成图与用户上传图均已转存服务端，跨设备可回读）。
function toServerMessages(): ServerMessage[] {
  return messages.value
    .filter(
      (m) =>
        m.content ||
        m.imageRefs?.length ||
        m.attachments?.some((a) => (a.kind === 'image' && a.ref) || a.kind === 'text'),
    )
    .map((m) => ({ role: m.role, content: serializeServerContent(m) }))
}

// 保存前把用户上传的图片附件转存服务端（幂等：已有 ref 的跳过），使其可跨设备回读。
async function persistUserAttachments() {
  const sid = currentId.value
  if (!sid) return
  for (const m of messages.value) {
    if (m.role !== 'user' || !m.attachments) continue
    for (const a of m.attachments) {
      if (a.kind !== 'image' || a.ref || !a.dataUrl) continue
      if (!a.dataUrl.startsWith('data:image/')) continue
      try {
        a.ref = `${SERVER_IMAGE_PREFIX}${await uploadChatImage(sid, a.dataUrl)}`
      } catch {
        /* 忽略：本轮未转存成功，下次保存再试 */
      }
    }
  }
}

function currentMeta(): SessionMeta | undefined {
  return sessions.value.find((s) => s.id === currentId.value)
}

async function loadSessions() {
  try {
    const list = await apiListSessions()
    sessions.value = list.map((s) => ({ id: s.id, title: s.title, model: s.model }))
  } catch {
    sessions.value = []
  }
  if (sessions.value.length) await switchSession(sessions.value[0].id)
  else await newSession()
}

// 每轮对话结束后保存当前会话（自动生成标题、置顶、整体覆盖消息）。
async function saveCurrent() {
  const meta = currentMeta()
  if (!meta) return
  if (!meta.title && messages.value.length) {
    const firstUser = messages.value.find((m) => m.role === 'user')
    if (firstUser) meta.title = firstUser.content.slice(0, 24) || t('chat.untitled')
  }
  meta.model = selectedModel.value
  sessions.value = [meta, ...sessions.value.filter((s) => s.id !== meta.id)]
  await persistUserAttachments()
  try {
    await apiUpdateSession(meta.id, {
      title: meta.title,
      model: meta.model,
      summary: summary.value,
      memory: memory.value,
      summarized_count: summarizedCount.value,
      messages: toServerMessages(),
    })
  } catch {
    /* 静默：保存失败不打断聊天 */
  }
}

function resetMemory() {
  summary.value = ''
  memory.value = ''
  summarizedCount.value = 0
}

async function newSession() {
  if (streaming.value) return
  try {
    const id = await apiCreateSession('', selectedModel.value)
    sessions.value.unshift({ id, title: '', model: selectedModel.value })
    currentId.value = id
  } catch (e) {
    errorMsg.value = friendlyError(e as Error)
    return
  }
  messages.value = []
  resetMemory()
  errorMsg.value = ''
  pending.value = []
  showHistory.value = false
}

async function switchSession(id: number) {
  if (streaming.value) return
  try {
    const s = await apiGetSession(id)
    currentId.value = id
    messages.value = await Promise.all((s.messages || []).map(fromServerMessage))
    selectedModel.value = s.model || models[0].id
    summary.value = s.summary || ''
    memory.value = s.memory || ''
    // 折叠水位以服务端为准，并夹取到当前消息条数（消息经过滤后条数可能微调）。
    const stored = typeof s.summarized_count === 'number' ? s.summarized_count : 0
    summarizedCount.value = Math.min(Math.max(stored, 0), messages.value.length)
  } catch (e) {
    errorMsg.value = friendlyError(e as Error)
    return
  }
  errorMsg.value = ''
  pending.value = []
  showHistory.value = false
}

async function renameSession(id: number) {
  const meta = sessions.value.find((s) => s.id === id)
  if (!meta) return
  const name = window.prompt(t('chat.rename'), meta.title)
  if (name == null) return
  meta.title = name.trim()
  try {
    // 仅更新元数据（不传 messages → 服务端保留原消息）
    await apiUpdateSession(id, { title: meta.title, model: meta.model })
  } catch (e) {
    errorMsg.value = friendlyError(e as Error)
  }
}

async function deleteSession(id: number) {
  try {
    await apiDeleteSession(id)
  } catch {
    /* 忽略删除失败 */
  }
  sessions.value = sessions.value.filter((s) => s.id !== id)
  if (currentId.value === id) {
    if (sessions.value.length) await switchSession(sessions.value[0].id)
    else await newSession()
  }
}

// ───────── 附件 ─────────

function readAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const r = new FileReader()
    r.onload = () => resolve(r.result as string)
    r.onerror = () => reject(r.error)
    r.readAsDataURL(file)
  })
}
function readAsText(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const r = new FileReader()
    r.onload = () => resolve(r.result as string)
    r.onerror = () => reject(r.error)
    r.readAsText(file)
  })
}

async function onFiles(e: Event) {
  const files = (e.target as HTMLInputElement).files
  if (!files) return
  errorMsg.value = ''
  for (const f of Array.from(files)) {
    const isImage = f.type.startsWith('image/')
    if (isImage) {
      if (f.size > MAX_IMAGE_BYTES) {
        errorMsg.value = t('chat.fileTooLarge', { name: f.name })
        continue
      }
      pending.value.push({ kind: 'image', name: f.name, mime: f.type, dataUrl: await readAsDataURL(f) })
    } else if (f.type.startsWith('text/') || /\.(txt|md)$/i.test(f.name)) {
      if (f.size > MAX_TEXT_BYTES) {
        errorMsg.value = t('chat.fileTooLarge', { name: f.name })
        continue
      }
      pending.value.push({ kind: 'text', name: f.name, mime: f.type || 'text/plain', text: await readAsText(f) })
    } else {
      errorMsg.value = t('chat.unsupportedType', { name: f.name })
    }
  }
  if (fileInput.value) fileInput.value.value = ''
}

function toApiContent(m: UiMessage): string | ContentPart[] {
  if (!m.attachments?.length) return m.content
  let text = m.content
  for (const a of m.attachments) {
    if (a.kind === 'text' && a.text) text += `\n\n[${a.name}]\n${a.text}`
  }
  const parts: ContentPart[] = [{ type: 'text', text }]
  for (const a of m.attachments) {
    if (a.kind === 'image' && a.dataUrl) parts.push({ type: 'image_url', image_url: { url: a.dataUrl } })
  }
  return parts
}

// ───────── 发送 / 重新生成 ─────────

async function scrollToBottom() {
  await nextTick()
  if (listEl.value) listEl.value.scrollTop = listEl.value.scrollHeight
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    /* 忽略复制失败 */
  }
}

// ───────── 图片全屏预览 ─────────

function openPreview(src: string) {
  if (src) previewSrc.value = src
}
function closePreview() {
  previewSrc.value = ''
}
// 助手 Markdown 里的图片（模型可能以 Markdown 图片返回）点击也可全屏。
function onMarkdownClick(e: MouseEvent) {
  const target = e.target as HTMLElement | null
  if (target?.tagName === 'IMG') {
    const img = target as HTMLImageElement
    openPreview(img.currentSrc || img.src)
  }
}
onKeyStroke('Escape', () => {
  if (previewSrc.value) closePreview()
})

// ───────── 会话记忆（长/中/短期）+ 压缩（compact） ─────────
//
// 短期：最近 SHORT_TERM_KEEP 条消息按原文进入上下文（全保真）。
// 中期：更早的消息被压缩成滚动摘要 summary（叙述性梗概）。
// 长期：跨轮稳定的事实 memory（用户身份/偏好/结论/实体）。
// 当「未折叠的消息」超过 COMPACT_TRIGGER 条时触发一次压缩：把溢出的最早一批折叠进
// summary/memory，只保留最近 SHORT_TERM_KEEP 条原文，从而把上下文长度稳定在可控区间。

const PROMPT_AGENT_MODEL = OPENAI_CODEX_DEFAULT_MODEL
const SHORT_TERM_KEEP = 8
const COMPACT_TRIGGER = 16
const MEMORY_MAX_CHARS = 2000
const SUMMARY_MAX_CHARS = 2000
const IMAGE_CONTEXT_MAX_TURNS = 8
const IMAGE_CONTEXT_MAX_CHARS = 1800
const IMAGE_CONTEXT_RECENT = 12

const BASE_CHAT_SYSTEM =
  '你是内置聊天助手。请始终结合本次对话的完整上下文作答：记住用户先前提供的信息、偏好与结论，'
  + '在多轮之间保持连贯；当用户使用「它 / 上面那个 / 继续 / 再来一个」等指代或省略时，'
  + '依据历史消息推断其真实指向，不要要求用户重复已经说过的内容。'

function userPromptsOf(history: UiMessage[]): string[] {
  return history.filter((m) => m.role === 'user' && m.content.trim()).map((m) => m.content.trim())
}

function clampText(s: string, max: number): string {
  const t2 = (s || '').trim()
  return t2.length > max ? t2.slice(0, max) : t2
}

// 把一条消息压成给「摘要器」阅读的一行文本（含图片/附件的占位说明）。
function messageToText(m: UiMessage): string {
  const who = m.role === 'user' ? '用户' : '助手'
  let body = (m.content || '').trim()
  if (!body && m.role === 'assistant' && m.images?.length) body = '（生成了一张图片）'
  if (m.role === 'user' && m.attachments?.length) {
    const imgs = m.attachments.filter((a) => a.kind === 'image').length
    const txts = m.attachments.filter((a) => a.kind === 'text').map((a) => a.name)
    if (imgs) body += ` （附带 ${imgs} 张图片）`
    if (txts.length) body += ` （附件：${txts.join('、')}）`
  }
  body = body.trim()
  return body ? `${who}：${body}` : ''
}

// 组装文本对话的实际请求消息：系统提示（含长期记忆 + 中期摘要）+ 短期原文窗口。
function buildTextMessages(history: UiMessage[]): ChatMessage[] {
  const start = Math.min(Math.max(summarizedCount.value, 0), history.length)
  const shortTerm = history.slice(start)
  let system = BASE_CHAT_SYSTEM
  if (memory.value.trim()) system += `\n\n【长期记忆 · 用户与对话的稳定事实】\n${memory.value.trim()}`
  if (summary.value.trim()) system += `\n\n【早前对话摘要】\n${summary.value.trim()}`
  return [
    { role: 'system', content: system },
    ...shortTerm
      .filter((m) => m.content || m.attachments?.length)
      .map((m): ChatMessage => ({ role: m.role, content: toApiContent(m) })),
  ]
}

const COMPACT_SYSTEM =
  '你是对话记忆压缩器。给定「已有长期记忆」「已有阶段摘要」和「新增对话片段」，产出更新后的两段记忆，'
  + '并且只返回一个 JSON 对象：{"memory": string, "summary": string}。'
  + 'memory=长期稳定信息（用户身份/偏好/明确结论/关键实体/约定的待办），在已有基础上合并去重、删去过时项，用简洁要点罗列；'
  + 'summary=到目前为止的连续对话梗概（把已有阶段摘要与新增片段融合成一段连贯叙述，保留具体事实与决定，省略寒暄客套）。'
  + '两段都用与对话相同的语言书写、尽量精炼；只输出 JSON，不要额外解释，不要代码块围栏。'

function buildCompactInput(mem: string, sum: string, batch: string): string {
  return [
    `已有长期记忆：\n${mem.trim() || '（无）'}`,
    `已有阶段摘要：\n${sum.trim() || '（无）'}`,
    `新增对话片段（需并入）：\n${batch}`,
    '请输出更新后的 JSON。',
  ].join('\n\n')
}

function parseCompactOutput(out: string): { memory: string; summary: string } {
  const text = (out || '').trim()
  const start = text.indexOf('{')
  const end = text.lastIndexOf('}')
  if (start >= 0 && end > start) {
    try {
      const j = JSON.parse(text.slice(start, end + 1)) as Record<string, unknown>
      return {
        memory: typeof j.memory === 'string' ? j.memory.trim() : '',
        summary: typeof j.summary === 'string' ? j.summary.trim() : '',
      }
    } catch {
      /* fall through */
    }
  }
  // 兜底：整段当作摘要，至少不丢中期上下文。
  return { memory: '', summary: text }
}

// 触发条件满足时，把最早溢出的一批消息折叠进 summary/memory（一次额外的模型调用）。
// best-effort：失败则保持水位不动，下一轮再尝试。
async function maybeCompact() {
  const total = messages.value.length
  if (total - summarizedCount.value < COMPACT_TRIGGER) return
  const foldEnd = total - SHORT_TERM_KEEP
  if (foldEnd <= summarizedCount.value) return

  const batch = messages.value.slice(summarizedCount.value, foldEnd)
  const batchText = batch.map(messageToText).filter(Boolean).join('\n')
  if (!batchText.trim()) {
    summarizedCount.value = foldEnd
    return
  }

  const sessionAtStart = currentId.value
  try {
    const out = await completeChat({
      model: PROMPT_AGENT_MODEL,
      messages: [
        { role: 'system', content: COMPACT_SYSTEM },
        { role: 'user', content: buildCompactInput(memory.value, summary.value, batchText) },
      ],
    })
    if (currentId.value !== sessionAtStart) return // 会话已切换，丢弃结果
    const parsed = parseCompactOutput(out)
    let changed = false
    if (parsed.memory) {
      memory.value = clampText(parsed.memory, MEMORY_MAX_CHARS)
      changed = true
    }
    if (parsed.summary) {
      summary.value = clampText(parsed.summary, SUMMARY_MAX_CHARS)
      changed = true
    }
    if (!changed) return // 没拿到有效结果，保持不动，下轮再压缩
    summarizedCount.value = foldEnd
    await saveCurrent()
  } catch {
    /* best-effort：压缩失败不影响聊天 */
  }
}

// 纯文本兜底：把历史指令编号拼接，供 agent 失败时使用。
function buildImagePrompt(history: UiMessage[]): string {
  const prompts = userPromptsOf(history)
  const current = prompts[prompts.length - 1] || ''
  const prior = prompts.slice(0, -1).slice(-IMAGE_CONTEXT_MAX_TURNS)
  if (!prior.length) return current
  let context = prior.map((p, i) => `${i + 1}. ${p}`).join('\n')
  if (context.length > IMAGE_CONTEXT_MAX_CHARS) {
    context = `…${context.slice(context.length - IMAGE_CONTEXT_MAX_CHARS)}`
  }
  return [
    'You are iteratively refining a single image across turns.',
    'Earlier instructions in this conversation, in order:',
    context,
    'Now apply the latest instruction while staying consistent with the above:',
    current,
  ].join('\n')
}

/**
 * Agent 步骤：用文本模型读对话（近程原文 + 长期记忆/中期摘要），产出「一条完整、自洽」的生图
 * prompt，把之前的画面细节延续下来再套用最新指令（如「字太多了」→ 减少文字）。
 * 首轮且无记忆时直出原始 prompt；失败时回退到 buildImagePrompt。
 */
async function agentImagePrompt(history: UiMessage[], signal: AbortSignal): Promise<string> {
  const prompts = userPromptsOf(history)
  const current = prompts[prompts.length - 1] || ''
  if (prompts.length <= 1 && !summary.value.trim() && !memory.value.trim()) return current // 首轮：直出

  const convo: ChatMessage[] = []
  if (memory.value.trim()) convo.push({ role: 'system', content: `已知长期记忆：\n${memory.value.trim()}` })
  if (summary.value.trim()) convo.push({ role: 'system', content: `早前对话摘要：\n${summary.value.trim()}` })
  for (const m of history.slice(-IMAGE_CONTEXT_RECENT)) {
    if (m.role === 'user' && m.content.trim()) {
      convo.push({ role: 'user', content: m.content.trim() })
    } else if (m.role === 'assistant' && m.images?.length) {
      convo.push({ role: 'assistant', content: '[已按上一条指令生成了一张图片]' })
    }
  }
  const system: ChatMessage = {
    role: 'system',
    content:
      'You are an image-prompt engineer for an iterative image editor. ' +
      'The user is refining ONE image across multiple turns. Read the whole conversation and output a SINGLE, ' +
      'complete, self-contained prompt for the NEXT image: carry over every visual detail established in earlier ' +
      'turns (subject, composition, style, colours, text), then apply the latest instruction. ' +
      'Treat short follow-ups (e.g. "too much text", "make it night", "bigger") as edits to the previous image, ' +
      'NOT as brand-new unrelated images. Keep the language of any on-image text as the user specified. ' +
      'Output ONLY the final prompt text — no preamble, no quotes, no explanation.',
  }
  try {
    const rewritten = (await completeChat({ model: PROMPT_AGENT_MODEL, messages: [system, ...convo] }, signal)).trim()
    return rewritten || buildImagePrompt(history)
  } catch {
    return buildImagePrompt(history)
  }
}

// ───────── 联网搜索 agent 循环 ─────────

const MAX_SEARCH_ROUNDS = 3
const SEARCH_MAX_RESULTS = 5

// 以 OpenAI 函数工具形式暴露给模型；模型自行判断是否需要联网。
const WEB_SEARCH_TOOL = {
  type: 'function',
  function: {
    name: 'web_search',
    description:
      '联网搜索获取实时/最新信息（新闻、当前事件、价格、天气、发布时间、事实核查，或任何知识截止之后的内容）。'
      + '当回答需要最新或可引用的外部信息时调用；不需要时可直接作答。',
    parameters: {
      type: 'object',
      properties: {
        query: { type: 'string', description: '检索关键词或问题，使用与用户相同的语言' },
      },
      required: ['query'],
    },
  },
}

// 文本对话一轮：无搜索时单轮流式；开启搜索时按需进入「调用 web_search → 回灌结果 → 继续」的 agent 循环。
async function runTextTurn(
  assistant: UiMessage,
  base: ChatMessage[],
  signal: AbortSignal,
  withSearch: boolean,
) {
  const messages: ChatCompletionMessage[] = [...base]
  const sources: ChatSource[] = []
  const maxRounds = withSearch ? MAX_SEARCH_ROUNDS : 1

  for (let round = 0; round < maxRounds; round++) {
    const isLast = round === maxRounds - 1
    // 最后一轮不再给工具，逼模型给出最终答案。
    const tools = withSearch && !isLast ? [WEB_SEARCH_TOOL] : undefined
    const result = await streamChatCompletion(
      { model: selectedModel.value, messages, tools },
      { signal, onDelta: (delta) => { assistant.content += delta; scrollToBottom() } },
    )

    if (signal.aborted) break // 用户已停止：不再发起后续搜索/回合

    if (tools && result.toolCalls.length) {
      // 丢弃本轮可能的前言（如"让我搜一下"），最终答案在后续轮流式呈现。
      assistant.content = ''
      messages.push({
        role: 'assistant',
        content: result.content || '',
        tool_calls: result.toolCalls.map((tc): AssistantToolCall => ({
          id: tc.id,
          type: 'function',
          function: { name: tc.name, arguments: tc.arguments },
        })),
      })
      for (const tc of result.toolCalls) {
        let query = ''
        try {
          query = String(JSON.parse(tc.arguments || '{}').query || '')
        } catch {
          /* 参数非 JSON，按空查询处理 */
        }
        assistant.searchStatus = query ? t('chat.searching', { query }) : t('chat.searchingGeneric')
        await scrollToBottom()
        let results: ChatSource[] = []
        if (tc.name === 'web_search' && query.trim()) {
          try {
            results = await chatSearch(query, SEARCH_MAX_RESULTS)
          } catch {
            /* 搜索不可用/失败：让模型无搜索继续作答 */
          }
        }
        for (const r of results) {
          if (r.url && !sources.some((s) => s.url === r.url)) sources.push(r)
        }
        messages.push({
          role: 'tool',
          tool_call_id: tc.id,
          content: JSON.stringify(results.length ? results : { note: 'no results (search unavailable or empty)' }),
        })
      }
      assistant.searchStatus = ''
      continue
    }
    // 无工具调用：本轮内容即最终答案。
    break
  }
  assistant.searchStatus = ''
  if (sources.length) assistant.sources = sources
}

// 每轮结束后：保存会话，再按需压缩（压缩内部会再次保存）。压缩期间 streaming 仍为 true，
// 避免用户在记忆更新过程中并发发起新一轮造成竞态。
async function afterTurn() {
  await saveCurrent()
  await maybeCompact()
}

// 末尾消息须为空的 assistant 占位；前面构成历史。
async function runAssistant() {
  const assistant = messages.value[messages.value.length - 1]
  const history = messages.value.slice(0, -1)
  streaming.value = true
  controller = new AbortController()
  const signal = controller.signal

  try {
    if (isImageModel(selectedModel.value)) {
      // Agent 先根据对话（近程原文 + 记忆/摘要）规划出完整 prompt，再交给生图模型。
      const prompt = await agentImagePrompt(history, signal)
      const imgs = await generateImage({ model: selectedModel.value, prompt }, signal)
      assistant.images = imgs
      assistant.imageRefs = await saveServerImages(currentId.value, imgs)
      if (!imgs.length) assistant.content = t('chat.noImage')
    } else {
      await runTextTurn(assistant, buildTextMessages(history), signal, canWebSearch.value && webSearchOn.value)
    }
  } catch (err) {
    if ((err as Error).name !== 'AbortError') errorMsg.value = friendlyError(err as Error)
  } finally {
    controller = null
    await afterTurn()
    streaming.value = false
    scrollToBottom()
  }
}

async function send() {
  if (!canSend.value) return
  const text = input.value.trim()
  const atts = pending.value.slice()
  input.value = ''
  pending.value = []
  errorMsg.value = ''

  const userMsg: UiMessage = { role: 'user', content: text }
  if (atts.length) userMsg.attachments = atts
  messages.value.push(userMsg)
  messages.value.push({ role: 'assistant', content: '' })
  await scrollToBottom()
  await runAssistant()
}

async function regenerate() {
  if (streaming.value) return
  if (messages.value[messages.value.length - 1]?.role !== 'assistant') return
  messages.value.pop()
  messages.value.push({ role: 'assistant', content: '' })
  errorMsg.value = ''
  await scrollToBottom()
  await runAssistant()
}

function stop() {
  controller?.abort()
}

watch(selectedModel, (model) => {
  const meta = currentMeta()
  if (meta && meta.model !== model) {
    meta.model = model
    // 仅更新元数据，不动消息
    apiUpdateSession(meta.id, { title: meta.title, model }).catch(() => {})
  }
})

onMounted(async () => {
  getChatCapabilities()
    .then((caps) => {
      webSearchAvailable.value = caps.web_search
    })
    .catch(() => {
      webSearchAvailable.value = false
    })
  await loadSessions()
})
</script>

<style scoped>
.markdown-body :deep(p) {
  margin: 0.25rem 0;
}
.markdown-body :deep(pre) {
  margin: 0.5rem 0;
  padding: 0.75rem;
  border-radius: 0.5rem;
  overflow-x: auto;
  background: rgb(243 244 246);
  font-size: 0.8125rem;
}
.dark .markdown-body :deep(pre) {
  background: rgb(30 41 59);
}
.markdown-body :deep(code) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.markdown-body :deep(:not(pre) > code) {
  padding: 0.1rem 0.3rem;
  border-radius: 0.25rem;
  background: rgb(243 244 246);
}
.dark .markdown-body :deep(:not(pre) > code) {
  background: rgb(30 41 59);
}
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 0.25rem 0;
  padding-left: 1.25rem;
  list-style: revert;
}
.markdown-body :deep(a) {
  color: rgb(37 99 235);
  text-decoration: underline;
}
.markdown-body :deep(table) {
  border-collapse: collapse;
  margin: 0.5rem 0;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid rgb(209 213 219);
  padding: 0.25rem 0.5rem;
}
</style>
