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

export interface StreamHandlers {
  signal?: AbortSignal
  onDelta: (text: string) => void
  onDone: () => void
  onError: (err: Error) => void
}

function authToken(): string {
  return localStorage.getItem('auth_token') || ''
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

export async function updateSession(
  id: number,
  payload: { title: string; model: string; messages?: ServerMessage[] },
): Promise<void> {
  await apiClient.put(`/chat/sessions/${id}`, payload)
}

export async function deleteSession(id: number): Promise<void> {
  await apiClient.delete(`/chat/sessions/${id}`)
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
    body: JSON.stringify(body),
    signal,
  })
  if (!res.ok) {
    const detail = await res.text().catch(() => '')
    throw new Error(detail || `image request failed: ${res.status}`)
  }
  const data = await res.json()
  const items: Array<{ b64_json?: string; url?: string }> = data?.data ?? []
  return items
    .map((it) => (it.b64_json ? `data:image/png;base64,${it.b64_json}` : it.url || ''))
    .filter((s): s is string => !!s)
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
      throw new Error(detail || `chat request failed: ${res.status}`)
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
