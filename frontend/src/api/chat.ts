/**
 * 内置聊天 Playground API。
 *
 * 对话走会话代理端点 /api/v1/chat/v1/chat/completions：前端用 JWT 登录态请求，
 * 后端桥接成「用户用自己内置 Key 调一次网关」，复用全部计费链路。
 *
 * 流式输出用原生 fetch + ReadableStream 解析 SSE（axios 不适合流式）。
 */

import { apiClient } from './client'

const CHAT_BASE = '/api/v1/chat'

// OpenAI 多模态 content part（文本 / 图片直传）。
export type ContentPart =
  | { type: 'text'; text: string }
  | { type: 'image_url'; image_url: { url: string } }

export interface ChatMessage {
  role: 'system' | 'user' | 'assistant'
  content: string | ContentPart[]
}

// 工具调用循环里会用到的更宽消息类型（assistant 带 tool_calls / role:tool 回灌）。
export interface AssistantToolCall {
  id: string
  type: 'function'
  function: { name: string; arguments: string }
}
export type ChatCompletionMessage =
  | ChatMessage
  | { role: 'assistant'; content: string; tool_calls: AssistantToolCall[] }
  | { role: 'tool'; tool_call_id: string; content: string }

// 一条联网搜索结果（供工具回灌与来源引用展示）。
export interface ChatSource {
  title: string
  url: string
  snippet: string
}

export interface StreamToolCall {
  id: string
  name: string
  arguments: string
}
export interface CompletionResult {
  content: string
  toolCalls: StreamToolCall[]
  finishReason: string
}

export interface StreamHandlers {
  signal?: AbortSignal
  onDelta: (text: string) => void
  onDone: () => void
  onError: (err: Error) => void
}

interface ImageGenerationItem {
  b64_json?: string
  result?: string
  base64?: string
  image_base64?: string
  url?: string
  mime_type?: string
  mimeType?: string
  content_type?: string
  output_format?: string
}

type UnknownRecord = Record<string, unknown>

function authToken(): string {
  return localStorage.getItem('auth_token') || ''
}

function responseErrorMessage(detail: string, status: number, fallback: string): string {
  const trimmed = detail.trim()
  if (!trimmed) return `${fallback}: ${status}`
  if (trimmed.startsWith('<') || trimmed.toLowerCase().includes('<html')) {
    return `${fallback}: ${status}`
  }
  return trimmed
}

function firstTrimmedString(...values: unknown[]): string {
  for (const value of values) {
    if (typeof value !== 'string') continue
    const trimmed = value.trim()
    if (trimmed) return trimmed
  }
  return ''
}

function imageMimeType(item: ImageGenerationItem): string {
  const explicit = firstTrimmedString(item.mime_type, item.mimeType, item.content_type)
  if (explicit.toLowerCase().startsWith('image/')) return explicit

  const format = firstTrimmedString(item.output_format).toLowerCase()
  if (!format) return 'image/png'
  if (format === 'jpg') return 'image/jpeg'
  return `image/${format}`
}

function imageItemToSrc(item: ImageGenerationItem): string {
  const b64 = firstTrimmedString(item.b64_json, item.result, item.base64, item.image_base64)
  if (b64) {
    if (b64.startsWith('data:image/')) return b64
    return `data:${imageMimeType(item)};base64,${b64}`
  }
  return item.url?.trim() || ''
}

function addImageSrc(out: string[], item: ImageGenerationItem) {
  const src = imageItemToSrc(item)
  if (src && !out.includes(src)) out.push(src)
}

function asRecord(value: unknown): UnknownRecord | null {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as UnknownRecord : null
}

function collectImagesFromPayload(payload: unknown, out: string[], eventName = '') {
  const obj = asRecord(payload)
  if (!obj) return

  if (Array.isArray(obj.data)) {
    for (const item of obj.data) {
      const image = asRecord(item)
      if (image) addImageSrc(out, image as ImageGenerationItem)
    }
  }

  const type = typeof obj.type === 'string' ? obj.type : ''
  const streamKind = `${eventName} ${type}`.trim().toLowerCase()
  if (streamKind.includes('partial_image')) return
  if (streamKind && !streamKind.includes('completed') && !streamKind.includes('output_item.done')) return
  addImageSrc(out, obj as ImageGenerationItem)

  const item = asRecord(obj.item)
  if (item) addImageSrc(out, item as ImageGenerationItem)

  const response = asRecord(obj.response)
  const output = Array.isArray(obj.output)
    ? obj.output
    : Array.isArray(response?.output)
      ? response.output
      : []
  for (const candidate of output) {
    const image = asRecord(candidate)
    if (image) addImageSrc(out, image as ImageGenerationItem)
  }
}

function imageStreamError(payload: unknown, eventName = ''): string {
  const obj = asRecord(payload)
  if (!obj) return ''
  const type = typeof obj.type === 'string' ? obj.type.toLowerCase() : ''
  const isError = eventName.toLowerCase() === 'error' || type === 'error' || !!obj.error
  if (!isError) return ''
  const error = asRecord(obj.error)
  if (error && typeof error.message === 'string') return error.message
  return typeof obj.message === 'string' ? obj.message : 'image request failed'
}

function parseSSEEvent(raw: string): { eventName: string; data: string } | null {
  const dataLines: string[] = []
  let eventName = ''
  for (const line of raw.split(/\r?\n/)) {
    if (!line || line.startsWith(':')) continue
    const sep = line.indexOf(':')
    const field = sep >= 0 ? line.slice(0, sep) : line
    let value = sep >= 0 ? line.slice(sep + 1) : ''
    if (value.startsWith(' ')) value = value.slice(1)
    if (field === 'event') eventName = value
    if (field === 'data') dataLines.push(value)
  }
  if (!dataLines.length) return null
  return { eventName, data: dataLines.join('\n') }
}

function nextSSEBoundary(buffer: string): { index: number; length: number } | null {
  const crlf = buffer.indexOf('\r\n\r\n')
  const lf = buffer.indexOf('\n\n')
  if (crlf >= 0 && (lf < 0 || crlf < lf)) return { index: crlf, length: 4 }
  if (lf >= 0) return { index: lf, length: 2 }
  return null
}

function consumeImageStreamEvent(raw: string, out: string[]): string {
  const event = parseSSEEvent(raw)
  if (!event) return ''
  if (event.data === '[DONE]') return ''
  try {
    const payload = JSON.parse(event.data)
    const err = imageStreamError(payload, event.eventName)
    if (err) return err
    collectImagesFromPayload(payload, out, event.eventName)
  } catch {
    // 忽略上游心跳或非 JSON 诊断行。
  }
  return ''
}

async function readImageStream(res: Response): Promise<string[]> {
  const images: string[] = []
  const reader = res.body?.getReader()
  if (!reader) return images

  const decoder = new TextDecoder()
  let buffer = ''
  let streamError = ''

  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    let boundary: { index: number; length: number } | null
    while ((boundary = nextSSEBoundary(buffer))) {
      const raw = buffer.slice(0, boundary.index)
      buffer = buffer.slice(boundary.index + boundary.length)
      if (!streamError) streamError = consumeImageStreamEvent(raw, images)
    }
  }

  buffer += decoder.decode()
  if (buffer.trim()) {
    if (!streamError) streamError = consumeImageStreamEvent(buffer, images)
  }
  if (streamError) throw new Error(streamError)
  return images
}

async function parseImageResponse(res: Response): Promise<string[]> {
  const contentType = res.headers.get('Content-Type') || ''
  if (contentType.toLowerCase().includes('text/event-stream')) {
    return readImageStream(res)
  }

  const data = await res.json()
  const images: string[] = []
  collectImagesFromPayload(data, images)
  return images
}

// ───────── 服务端会话历史（跨设备同步，JWT 鉴权，走 apiClient） ─────────

export interface ServerMessage {
  role: 'user' | 'assistant'
  content: string
}
export interface ServerSession {
  id: number
  title: string
  model: string
  updated_at: string
  // 会话记忆：中期滚动摘要 / 长期稳定事实 / 已折叠进摘要的前缀消息条数。
  summary?: string
  memory?: string
  summarized_count?: number
  messages?: ServerMessage[]
}

export interface SessionUpdatePayload {
  title: string
  model: string
  summary?: string
  memory?: string
  summarized_count?: number
  messages?: ServerMessage[]
}

export async function listSessions(): Promise<ServerSession[]> {
  const r = await apiClient.get('/chat/sessions')
  return (r.data as ServerSession[]) || []
}

export async function getSession(id: number): Promise<ServerSession> {
  const r = await apiClient.get(`/chat/sessions/${id}`)
  return r.data as ServerSession
}

export async function createSession(title = '', model = ''): Promise<number> {
  const r = await apiClient.post('/chat/sessions', { title, model })
  return (r.data as { id: number }).id
}

export async function updateSession(id: number, payload: SessionUpdatePayload): Promise<void> {
  await apiClient.put(`/chat/sessions/${id}`, payload)
}

export async function deleteSession(id: number): Promise<void> {
  await apiClient.delete(`/chat/sessions/${id}`)
}

// ───────── 服务端图片存储（生成图 / 上传图落库，跨设备回读） ─────────

function blobToDataUrl(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = () => reject(reader.error || new Error('read image failed'))
    reader.readAsDataURL(blob)
  })
}

/** 上传一张图片（data URL 或裸 base64）到当前会话，返回服务端图片 ID。 */
export async function uploadChatImage(sessionId: number, image: string): Promise<string> {
  const r = await apiClient.post(`/chat/sessions/${sessionId}/images`, { image })
  const id = (r.data as { id?: string }).id
  if (!id) throw new Error('upload returned no image id')
  return id
}

/** 按图片 ID 回读为可直接用于 <img src> 的 data URL（走 apiClient 以复用鉴权/刷新）。 */
export async function fetchChatImageDataUrl(id: string): Promise<string> {
  const r = await apiClient.get(`/chat/images/${id}`, { responseType: 'blob' })
  return blobToDataUrl(r.data as Blob)
}

/**
 * 生图（GPT Image 等）。返回可直接用于 <img src> 的数据 URL 列表。
 * 兼容上游返回 b64_json 或 url 两种形态。
 */
export async function generateImage(
  body: { model: string; prompt: string },
  signal?: AbortSignal,
): Promise<string[]> {
  const res = await fetch(`${CHAT_BASE}/v1/images/generations`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${authToken()}`,
    },
    body: JSON.stringify({ ...body, stream: true, response_format: 'b64_json' }),
    signal,
  })
  if (!res.ok) {
    const detail = await res.text().catch(() => '')
    throw new Error(responseErrorMessage(detail, res.status, 'image request failed'))
  }
  return parseImageResponse(res)
}

/** 获取可用模型列表（复用网关 /v1/models）。 */
export async function listModels(): Promise<string[]> {
  const res = await fetch(`${CHAT_BASE}/v1/models`, {
    headers: { Authorization: `Bearer ${authToken()}` },
  })
  if (!res.ok) {
    throw new Error(`models request failed: ${res.status}`)
  }
  const data = await res.json()
  const items: Array<{ id?: string }> = data?.data ?? []
  return items.map((m) => m.id).filter((id): id is string => !!id)
}

/**
 * 发起流式对话。逐行解析 `data: {...}` / `data: [DONE]`，
 * 把 choices[0].delta.content 增量回调给调用方。
 */
export async function streamChat(
  body: { model: string; messages: ChatMessage[] },
  handlers: StreamHandlers,
): Promise<void> {
  const { signal, onDelta, onDone, onError } = handlers
  try {
    const res = await fetch(`${CHAT_BASE}/v1/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authToken()}`,
      },
      body: JSON.stringify({ ...body, stream: true }),
      signal,
    })

    if (!res.ok || !res.body) {
      const detail = await res.text().catch(() => '')
      throw new Error(responseErrorMessage(detail, res.status, 'chat request failed'))
    }

    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    for (;;) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })

      // SSE 事件以空行分隔；逐行取出 data: 负载。
      let nl: number
      while ((nl = buffer.indexOf('\n')) >= 0) {
        const line = buffer.slice(0, nl).trim()
        buffer = buffer.slice(nl + 1)
        if (!line.startsWith('data:')) continue
        const payload = line.slice(5).trim()
        if (payload === '' ) continue
        if (payload === '[DONE]') {
          onDone()
          return
        }
        try {
          const json = JSON.parse(payload)
          const delta: string = json?.choices?.[0]?.delta?.content ?? ''
          if (delta) onDelta(delta)
        } catch {
          // 忽略非 JSON 行（注释/心跳）
        }
      }
    }
    onDone()
  } catch (err) {
    if ((err as Error).name === 'AbortError') {
      onDone()
      return
    }
    onError(err as Error)
  }
}

/**
 * 非流式对话：内部复用 streamChat 的 SSE 解析，把增量拼成整段文本返回。
 * 用于「Agent 提示词规划」等一次性需要完整结果、不需要逐字上屏的场景。
 */
export function completeChat(
  body: { model: string; messages: ChatMessage[] },
  signal?: AbortSignal,
): Promise<string> {
  return new Promise((resolve, reject) => {
    let text = ''
    streamChat(body, {
      signal,
      onDelta: (delta) => {
        text += delta
      },
      onDone: () => resolve(text),
      onError: (err) => reject(err),
    })
  })
}

// ───────── 工具调用（联网搜索）循环支持 ─────────

function finalizeCompletion(
  content: string,
  toolAcc: Record<number, StreamToolCall>,
  finishReason: string,
): CompletionResult {
  const toolCalls = Object.keys(toolAcc)
    .map(Number)
    .sort((a, b) => a - b)
    .map((i) => toolAcc[i])
    .filter((tc) => tc.name)
  return { content, toolCalls, finishReason: finishReason || (toolCalls.length ? 'tool_calls' : 'stop') }
}

/**
 * 流式对话（工具感知版）：在回调增量文本的同时，累积模型的 tool_calls 并返回 finish_reason，
 * 供聊天页实现「模型调用 web_search → 前端执行 → 回灌结果 → 继续作答」的 agent 循环。
 * 无 tools 时行为等价于 streamChat（仅回调 content 增量）。
 */
export async function streamChatCompletion(
  body: { model: string; messages: ChatCompletionMessage[]; tools?: unknown[]; tool_choice?: unknown },
  handlers: { signal?: AbortSignal; onDelta?: (text: string) => void } = {},
): Promise<CompletionResult> {
  const res = await fetch(`${CHAT_BASE}/v1/chat/completions`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${authToken()}`,
    },
    body: JSON.stringify({ ...body, stream: true }),
    signal: handlers.signal,
  })
  if (!res.ok || !res.body) {
    const detail = await res.text().catch(() => '')
    throw new Error(responseErrorMessage(detail, res.status, 'chat request failed'))
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let content = ''
  let finishReason = ''
  const toolAcc: Record<number, StreamToolCall> = {}

  const consume = (payload: string): CompletionResult | null => {
    if (payload === '[DONE]') return finalizeCompletion(content, toolAcc, finishReason)
    try {
      const json = JSON.parse(payload)
      const choice = json?.choices?.[0]
      const delta = choice?.delta
      if (delta?.content) {
        content += delta.content
        handlers.onDelta?.(delta.content)
      }
      if (Array.isArray(delta?.tool_calls)) {
        for (const tc of delta.tool_calls) {
          const idx = typeof tc?.index === 'number' ? tc.index : 0
          const acc = toolAcc[idx] ?? (toolAcc[idx] = { id: '', name: '', arguments: '' })
          if (tc?.id) acc.id = tc.id
          if (tc?.function?.name) acc.name = tc.function.name
          if (typeof tc?.function?.arguments === 'string') acc.arguments += tc.function.arguments
        }
      }
      if (choice?.finish_reason) finishReason = choice.finish_reason
    } catch {
      // 忽略非 JSON 行（注释/心跳）
    }
    return null
  }

  try {
    for (;;) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      let nl: number
      while ((nl = buffer.indexOf('\n')) >= 0) {
        const line = buffer.slice(0, nl).trim()
        buffer = buffer.slice(nl + 1)
        if (!line.startsWith('data:')) continue
        const payload = line.slice(5).trim()
        if (!payload) continue
        const result = consume(payload)
        if (result) return result
      }
    }
  } catch (err) {
    if ((err as Error).name === 'AbortError') return finalizeCompletion(content, toolAcc, finishReason)
    throw err
  }
  return finalizeCompletion(content, toolAcc, finishReason)
}

/** 执行一次联网搜索（复用网关侧搜索提供方与配额）。未配置时后端返回 503。 */
export async function chatSearch(query: string, maxResults = 5): Promise<ChatSource[]> {
  const r = await apiClient.post('/chat/search', { query, max_results: maxResults })
  return (r.data as { results?: ChatSource[] }).results ?? []
}

/** 查询聊天页可用能力（当前：是否已配置联网搜索）。失败按不可用处理。 */
export async function getChatCapabilities(): Promise<{ web_search: boolean }> {
  try {
    const r = await apiClient.get('/chat/capabilities')
    return { web_search: !!(r.data as { web_search?: boolean }).web_search }
  } catch {
    return { web_search: false }
  }
}
