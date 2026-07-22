import { apiClient } from './client'
import type { OpenAITokenInfo } from '@/composables/useOpenAIOAuth'
import type { GeminiTokenInfo } from '@/composables/useGeminiOAuth'
import type { AntigravityTokenInfo } from '@/api/admin/antigravity'
import type { GrokTokenInfo } from '@/api/admin/grok'

// Provider self-service OAuth flows. Same request/response shapes as the admin
// endpoints, but scoped under /provider/oauth and never carrying a proxy_id.

export interface ProviderAuthUrlResult {
  auth_url: string
  session_id: string
  state?: string
}

export interface AnthropicTokenInfo {
  access_token?: string
  refresh_token?: string
  expires_at?: number
  email_address?: string
  [key: string]: unknown
}

const post = async <T>(path: string, payload: Record<string, unknown> = {}): Promise<T> => {
  const { data } = await apiClient.post<T>(`/provider/oauth${path}`, payload)
  return data
}

export async function generateAnthropicAuthUrl(setupToken: boolean): Promise<ProviderAuthUrlResult> {
  return post<ProviderAuthUrlResult>(
    setupToken ? '/anthropic/generate-setup-token-url' : '/anthropic/generate-auth-url'
  )
}

export async function exchangeAnthropicCode(
  payload: { session_id: string; code: string },
  setupToken: boolean
): Promise<AnthropicTokenInfo> {
  return post<AnthropicTokenInfo>(
    setupToken ? '/anthropic/exchange-setup-token-code' : '/anthropic/exchange-code',
    payload
  )
}

export async function generateOpenAIAuthUrl(): Promise<ProviderAuthUrlResult> {
  return post<ProviderAuthUrlResult>('/openai/generate-auth-url')
}

export async function exchangeOpenAICode(payload: {
  session_id: string
  code: string
  state: string
}): Promise<OpenAITokenInfo> {
  return post<OpenAITokenInfo>('/openai/exchange-code', payload)
}

export async function generateGeminiAuthUrl(payload: {
  project_id?: string
  oauth_type?: string
  tier_id?: string
}): Promise<ProviderAuthUrlResult> {
  return post<ProviderAuthUrlResult>('/gemini/auth-url', payload)
}

export async function exchangeGeminiCode(payload: {
  session_id: string
  state: string
  code: string
  oauth_type?: string
  tier_id?: string
}): Promise<GeminiTokenInfo> {
  return post<GeminiTokenInfo>('/gemini/exchange-code', payload)
}

export async function generateAntigravityAuthUrl(): Promise<ProviderAuthUrlResult> {
  return post<ProviderAuthUrlResult>('/antigravity/auth-url')
}

export async function exchangeAntigravityCode(payload: {
  session_id: string
  state: string
  code: string
}): Promise<AntigravityTokenInfo> {
  return post<AntigravityTokenInfo>('/antigravity/exchange-code', payload)
}

export async function generateGrokAuthUrl(): Promise<ProviderAuthUrlResult> {
  return post<ProviderAuthUrlResult>('/grok/auth-url')
}

export async function exchangeGrokCode(payload: {
  session_id: string
  code: string
  state: string
}): Promise<GrokTokenInfo> {
  return post<GrokTokenInfo>('/grok/exchange-code', payload)
}

export const providerOAuthAPI = {
  generateAnthropicAuthUrl,
  exchangeAnthropicCode,
  generateOpenAIAuthUrl,
  exchangeOpenAICode,
  generateGeminiAuthUrl,
  exchangeGeminiCode,
  generateAntigravityAuthUrl,
  exchangeAntigravityCode,
  generateGrokAuthUrl,
  exchangeGrokCode,
}

export default providerOAuthAPI
