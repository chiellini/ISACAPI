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
              <!-- 生成的图片 -->
              <div v-if="msg.images?.length" class="flex flex-wrap gap-2">
                <img v-for="(src, j) in msg.images" :key="j" :src="src" class="max-h-64 rounded-lg" alt="generated" />
              </div>
              <!-- 助手文本：Markdown 渲染 -->
              <div
                v-else-if="msg.role === 'assistant'"
                class="markdown-body"
                v-html="msg.content ? renderMarkdown(msg.content) : (streaming ? '…' : '')"
              ></div>
              <!-- 用户文本 + 附件 -->
              <template v-else>
                {{ msg.content }}
                <div v-if="msg.attachments?.length" class="mt-2 flex flex-wrap gap-2">
                  <img
                    v-for="(a, j) in msg.attachments.filter((x) => x.kind === 'image')"
                    :key="'img' + j"
                    :src="a.dataUrl"
                    class="max-h-32 rounded-lg"
                    alt="attachment"
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
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { renderMarkdown } from '@/utils/markdown'
import {
  generateImage,
  streamChat,
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
  { id: 'gpt-5.5', label: 'gpt-5.5' },
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
  attachments?: Attachment[]
}
// 侧栏只保留会话元数据；消息按需从服务端拉取。
interface SessionMeta {
  id: number
  title: string
  model: string
}

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
let controller: AbortController | null = null

const canSend = computed(
  () => (!!input.value.trim() || pending.value.length > 0) && !!selectedModel.value && !streaming.value,
)

function isImageModel(id: string): boolean {
  return id.startsWith('gpt-image-')
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
  if (hay.includes('chat_unavailable') || hay.includes('no_available_chat_group')) return t('chat.errUnavailable')
  if (hay.includes('not_found') || hay.includes('not supported') || hay.includes('model')) return t('chat.errModel')
  return msg || t('chat.errGeneric')
}

// ───────── 历史持久化（服务端，跨设备同步；图片不落库） ─────────

// 仅保存文本消息（占位/纯图片消息跳过），与服务端"图片不持久化"一致。
function toServerMessages(): ServerMessage[] {
  return messages.value
    .filter((m) => m.content)
    .map((m) => ({ role: m.role, content: m.content }))
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
    messages.value = (s.messages || []).map((m) => ({ role: m.role, content: m.content }))
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

// 末尾消息须为空的 assistant 占位；前面构成历史。
async function runAssistant() {
  const assistant = messages.value[messages.value.length - 1]
  const history = messages.value.slice(0, -1)
  streaming.value = true
  controller = new AbortController()

  if (isImageModel(selectedModel.value)) {
    const lastUser = [...history].reverse().find((m) => m.role === 'user')
    try {
      const imgs = await generateImage({ model: selectedModel.value, prompt: lastUser?.content || '' }, controller.signal)
      assistant.images = imgs
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

  const apiMessages: ChatMessage[] = history.map((m) => ({ role: m.role, content: toApiContent(m) }))
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
