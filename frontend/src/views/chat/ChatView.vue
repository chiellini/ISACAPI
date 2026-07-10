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
              <div
                v-else-if="msg.role === 'assistant'"
                class="markdown-body"
                v-html="msg.content ? renderMarkdown(msg.content) : (streaming ? '…' : '')"
                @click="onMarkdownClick"
              ></div>
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
  streamChat,
  completeChat,
  listSessions as apiListSessions,
  getSession as apiGetSession,
  createSession as apiCreateSession,
  updateSession as apiUpdateSession,
  deleteSession as apiDeleteSession,
  type ChatMessage,
  type ContentPart,
  type ServerMessage,
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
}
interface UiMessage {
  role: 'user' | 'assistant'
  content: string
  images?: string[]
  imageRefs?: string[]
  attachments?: Attachment[]
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

interface StoredImageMessagePayload {
  type: typeof IMAGE_MESSAGE_PAYLOAD_TYPE
  version: 1
  content?: string
  imageRefs?: string[]
}

const IMAGE_MESSAGE_PAYLOAD_TYPE = 'isac-chat-local-images'
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

async function saveLocalImages(sessionId: number, images: string[]): Promise<string[]> {
  const refs: string[] = []
  for (const src of images) {
    try {
      const ref = await putLocalImage(sessionId, src)
      if (ref) refs.push(ref)
    } catch {
      if (src.trim()) refs.push(src.trim())
    }
  }
  return refs
}

async function loadLocalImages(refs: string[] = []): Promise<string[]> {
  const images: string[] = []
  for (const ref of refs) {
    try {
      const src = await getLocalImage(ref)
      if (src) images.push(src)
    } catch {
      if (ref.startsWith('data:image/') || /^https?:\/\//i.test(ref)) images.push(ref)
    }
  }
  return images
}

function parseStoredImageMessage(content: string): StoredImageMessagePayload | null {
  if (!content.trim().startsWith('{')) return null
  try {
    const payload = JSON.parse(content) as Partial<StoredImageMessagePayload>
    if (payload.type !== IMAGE_MESSAGE_PAYLOAD_TYPE || payload.version !== 1) return null
    return {
      type: IMAGE_MESSAGE_PAYLOAD_TYPE,
      version: 1,
      content: typeof payload.content === 'string' ? payload.content : '',
      imageRefs: Array.isArray(payload.imageRefs)
        ? payload.imageRefs.filter((ref): ref is string => typeof ref === 'string' && !!ref.trim())
        : [],
    }
  } catch {
    return null
  }
}

function serializeServerContent(message: UiMessage): string {
  if (message.imageRefs?.length) {
    const payload: StoredImageMessagePayload = {
      type: IMAGE_MESSAGE_PAYLOAD_TYPE,
      version: 1,
      content: message.content,
      imageRefs: message.imageRefs,
    }
    return JSON.stringify(payload)
  }
  return message.content
}

async function fromServerMessage(message: ServerMessage): Promise<UiMessage> {
  const payload = parseStoredImageMessage(message.content)
  if (!payload) return { role: message.role, content: message.content }

  const images = await loadLocalImages(payload.imageRefs)
  return {
    role: message.role,
    content: payload.content || (images.length ? '' : t('chat.noImage')),
    images: images.length ? images : undefined,
    imageRefs: payload.imageRefs,
  }
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

// 服务端只保存文本和本地图片引用；图片 data URL 放在浏览器 IndexedDB。
function toServerMessages(): ServerMessage[] {
  return messages.value
    .filter((m) => m.content || m.imageRefs?.length)
    .map((m) => ({ role: m.role, content: serializeServerContent(m) }))
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
  try {
    await apiUpdateSession(meta.id, { title: meta.title, model: meta.model, messages: toServerMessages() })
  } catch {
    /* 静默：保存失败不打断聊天 */
  }
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

// ───────── Agent：生图提示词规划（多轮记忆） ─────────

// gpt-image 系列没有原生多轮记忆：不带上下文时，「字太多了」这类追加指令会被
// 当成全新请求。这里的兜底把本会话历史里的「用户指令」按顺序拼进 prompt。
const IMAGE_CONTEXT_MAX_TURNS = 8
const IMAGE_CONTEXT_MAX_CHARS = 1800
const PROMPT_AGENT_MODEL = OPENAI_CODEX_DEFAULT_MODEL

function userPromptsOf(history: UiMessage[]): string[] {
  return history.filter((m) => m.role === 'user' && m.content.trim()).map((m) => m.content.trim())
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
 * Agent 步骤：用文本模型读完整对话，产出「一条完整、自洽」的生图 prompt，
 * 把之前的画面细节延续下来再套用最新指令（如「字太多了」→ 减少文字）。
 * 首轮无历史时直出原始 prompt；失败/无文本模型时回退到 buildImagePrompt。
 */
async function agentImagePrompt(history: UiMessage[], signal: AbortSignal): Promise<string> {
  const prompts = userPromptsOf(history)
  const current = prompts[prompts.length - 1] || ''
  if (prompts.length <= 1) return current // 首轮：无需规划，直出

  const convo: ChatMessage[] = []
  for (const m of history) {
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

// 末尾消息须为空的 assistant 占位；前面构成历史。
async function runAssistant() {
  const assistant = messages.value[messages.value.length - 1]
  const history = messages.value.slice(0, -1)
  streaming.value = true
  controller = new AbortController()

  if (isImageModel(selectedModel.value)) {
    const signal = controller.signal
    try {
      // Agent 先根据整段对话规划出「带记忆」的完整 prompt，再交给生图模型。
      const prompt = await agentImagePrompt(history, signal)
      const imgs = await generateImage({ model: selectedModel.value, prompt }, signal)
      assistant.images = imgs
      assistant.imageRefs = await saveLocalImages(currentId.value, imgs)
      if (!imgs.length) assistant.content = t('chat.noImage')
    } catch (err) {
      errorMsg.value = friendlyError(err as Error)
    } finally {
      streaming.value = false
      controller = null
      saveCurrent()
      scrollToBottom()
    }
    return
  }

  // 首条 system 明确要求「带记忆」：结合完整对话上下文、理解指代、保持多轮连贯。
  const apiMessages: ChatMessage[] = [
    {
      role: 'system',
      content:
        '你是内置聊天助手。请始终结合本次对话的完整上下文作答：记住用户先前提供的信息、偏好与结论，'
        + '在多轮之间保持连贯；当用户使用「它 / 上面那个 / 继续 / 再来一个」等指代或省略时，'
        + '依据历史消息推断其真实指向，不要要求用户重复已经说过的内容。',
    },
    ...history
      .filter((m) => m.content || m.attachments?.length)
      .map((m): ChatMessage => ({ role: m.role, content: toApiContent(m) })),
  ]
  await streamChat(
    { model: selectedModel.value, messages: apiMessages },
    {
      signal: controller.signal,
      onDelta: (delta) => {
        assistant.content += delta
        scrollToBottom()
      },
      onDone: () => {
        streaming.value = false
        controller = null
        saveCurrent()
      },
      onError: (err) => {
        streaming.value = false
        controller = null
        errorMsg.value = friendlyError(err)
        saveCurrent()
      },
    },
  )
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

onMounted(loadSessions)
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
