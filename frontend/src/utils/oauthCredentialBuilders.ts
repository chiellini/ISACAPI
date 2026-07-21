import type { OpenAITokenInfo } from '@/composables/useOpenAIOAuth'
import type { GeminiTokenInfo } from '@/composables/useGeminiOAuth'
import type { AntigravityTokenInfo } from '@/api/admin/antigravity'
import type { GrokTokenInfo } from '@/api/admin/grok'

// Pure token-info → account-credentials builders, shared by the admin
// create-account flow and the provider self-service flow. Keep aligned with
// the backend BuildAccountCredentials implementations.

export function buildOpenAIOAuthCredentials(tokenInfo: OpenAITokenInfo): Record<string, unknown> {
  const creds: Record<string, unknown> = {
    access_token: tokenInfo.access_token,
    expires_at: tokenInfo.expires_at
  }

  // 仅在返回了新的 refresh_token 时才写入，防止用空值覆盖已有令牌
  if (tokenInfo.refresh_token) {
    creds.refresh_token = tokenInfo.refresh_token
  }
  if (tokenInfo.id_token) {
    creds.id_token = tokenInfo.id_token
  }
  if (tokenInfo.email) {
    creds.email = tokenInfo.email
  }
  if (tokenInfo.chatgpt_account_id) {
    creds.chatgpt_account_id = tokenInfo.chatgpt_account_id
  }
  if (tokenInfo.chatgpt_user_id) {
    creds.chatgpt_user_id = tokenInfo.chatgpt_user_id
  }
  if (tokenInfo.organization_id) {
    creds.organization_id = tokenInfo.organization_id
  }
  if (tokenInfo.plan_type) {
    creds.plan_type = tokenInfo.plan_type
  }
  if (tokenInfo.subscription_expires_at) {
    creds.subscription_expires_at = tokenInfo.subscription_expires_at
  }
  if (tokenInfo.client_id) {
    creds.client_id = tokenInfo.client_id
  }

  return creds
}

const normalizeExpiresAt = (expiresAt: number | string | undefined): string | undefined => {
  if (typeof expiresAt === 'number' && Number.isFinite(expiresAt)) {
    return Math.floor(expiresAt).toString()
  }
  if (typeof expiresAt === 'string' && expiresAt.trim()) {
    return expiresAt.trim()
  }
  return undefined
}

export function buildGeminiOAuthCredentials(tokenInfo: GeminiTokenInfo): Record<string, unknown> {
  return {
    access_token: tokenInfo.access_token,
    refresh_token: tokenInfo.refresh_token,
    token_type: tokenInfo.token_type,
    expires_at: normalizeExpiresAt(tokenInfo.expires_at),
    scope: tokenInfo.scope,
    project_id: tokenInfo.project_id,
    oauth_type: tokenInfo.oauth_type,
    tier_id: tokenInfo.tier_id
  }
}

export function buildAntigravityOAuthCredentials(
  tokenInfo: AntigravityTokenInfo,
  fallbackRefreshToken?: string
): Record<string, unknown> {
  const refreshToken = tokenInfo.refresh_token?.trim()
    ? tokenInfo.refresh_token
    : fallbackRefreshToken

  return {
    access_token: tokenInfo.access_token,
    refresh_token: refreshToken,
    token_type: tokenInfo.token_type,
    expires_at: normalizeExpiresAt(tokenInfo.expires_at),
    project_id: tokenInfo.project_id,
    email: tokenInfo.email
  }
}

export function buildGrokOAuthCredentials(tokenInfo: GrokTokenInfo): Record<string, unknown> {
  const credentials: Record<string, unknown> = {
    access_token: tokenInfo.access_token,
    token_type: tokenInfo.token_type,
    expires_at: tokenInfo.expires_at,
    client_id: tokenInfo.client_id,
    scope: tokenInfo.scope,
    email: tokenInfo.email,
    sub: tokenInfo.sub,
    team_id: tokenInfo.team_id,
    subscription_tier: tokenInfo.subscription_tier,
    entitlement_status: tokenInfo.entitlement_status,
    base_url: 'https://cli-chat-proxy.grok.com/v1'
  }
  if (tokenInfo.refresh_token) credentials.refresh_token = tokenInfo.refresh_token
  if (tokenInfo.id_token) credentials.id_token = tokenInfo.id_token
  return Object.fromEntries(Object.entries(credentials).filter(([, value]) => value !== undefined && value !== ''))
}
