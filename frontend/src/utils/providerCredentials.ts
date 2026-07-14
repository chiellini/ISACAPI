import type { AccountType } from '@/types'

export class ProviderCredentialError extends Error {
  constructor(message = 'Invalid provider credentials') {
    super(message)
    this.name = 'ProviderCredentialError'
  }
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value)

const normalizeExpiry = (value: unknown): unknown => {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return Math.trunc(value > 100_000_000_000 ? value / 1000 : value)
  }
  if (typeof value === 'string') {
    const trimmed = value.trim()
    if (/^\d+$/.test(trimmed)) {
      const numeric = Number(trimmed)
      return Math.trunc(numeric > 100_000_000_000 ? numeric / 1000 : numeric)
    }
    const timestamp = Date.parse(trimmed)
    if (Number.isFinite(timestamp)) return Math.trunc(timestamp / 1000)
  }
  return value
}

const rawCredential = (value: string, accountType: AccountType): Record<string, unknown> => {
  if (accountType === 'apikey' || accountType === 'upstream' || accountType === 'bedrock') {
    return { api_key: value }
  }
  if (accountType === 'service_account') {
    return { service_account_json: value }
  }
  return { access_token: value }
}

const normalizeKnownFields = (source: Record<string, unknown>): Record<string, unknown> => {
  const result: Record<string, unknown> = { ...source }
  const aliases: Record<string, string> = {
    accessToken: 'access_token',
    refreshToken: 'refresh_token',
    idToken: 'id_token',
    expiresAt: 'expires_at',
    accountId: 'chatgpt_account_id',
    subscriptionType: 'subscription_type',
    rateLimitTier: 'rate_limit_tier',
  }

  for (const [from, to] of Object.entries(aliases)) {
    if (result[to] === undefined && result[from] !== undefined) result[to] = result[from]
    delete result[from]
  }
  if (result.chatgpt_account_id === undefined && result.account_id !== undefined) {
    result.chatgpt_account_id = result.account_id
  }
  if (result.expires_at !== undefined) result.expires_at = normalizeExpiry(result.expires_at)
  return result
}

/**
 * Normalizes direct tokens, Claude credentials JSON, Codex auth.json and
 * generic credential objects into the snake_case shape expected by accounts.
 */
export function normalizeProviderCredentials(input: string, accountType: AccountType): Record<string, unknown> {
  const trimmed = input.trim()
  if (!trimmed) throw new ProviderCredentialError('Credentials are required')

  let parsed: unknown
  try {
    parsed = JSON.parse(trimmed)
  } catch {
    if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
      throw new ProviderCredentialError()
    }
    return rawCredential(trimmed, accountType)
  }

  if (typeof parsed === 'string') return rawCredential(parsed.trim(), accountType)
  if (!isRecord(parsed)) throw new ProviderCredentialError()

  if (accountType === 'service_account' && parsed.type === 'service_account') {
    return {
      service_account_json: JSON.stringify(parsed),
      ...(typeof parsed.project_id === 'string' ? { project_id: parsed.project_id } : {}),
      ...(typeof parsed.client_email === 'string' ? { client_email: parsed.client_email } : {}),
    }
  }

  let source = parsed
  if (isRecord(source.credentials)) source = source.credentials
  if (isRecord(source.claudeAiOauth)) source = source.claudeAiOauth
  if (isRecord(source.tokens)) {
    source = {
      ...source.tokens,
      ...(source.OPENAI_API_KEY ? { api_key: source.OPENAI_API_KEY } : {}),
    }
  }

  const normalized = normalizeKnownFields(source)
  if (Object.keys(normalized).length === 0) throw new ProviderCredentialError()
  return normalized
}

