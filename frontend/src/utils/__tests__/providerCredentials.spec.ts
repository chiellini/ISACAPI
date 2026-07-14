import { describe, expect, it } from 'vitest'
import { normalizeProviderCredentials, ProviderCredentialError } from '@/utils/providerCredentials'

describe('provider credential normalization', () => {
  it('normalizes Claude desktop credentials and millisecond expiry', () => {
    const result = normalizeProviderCredentials(JSON.stringify({
      claudeAiOauth: {
        accessToken: 'access',
        refreshToken: 'refresh',
        expiresAt: 1_800_000_000_000,
        subscriptionType: 'pro',
        rateLimitTier: 'tier-1',
      },
    }), 'oauth')

    expect(result).toMatchObject({
      access_token: 'access',
      refresh_token: 'refresh',
      expires_at: 1_800_000_000,
      subscription_type: 'pro',
      rate_limit_tier: 'tier-1',
    })
    expect(result).not.toHaveProperty('accessToken')
  })

  it('normalizes Codex auth.json token fields and account ownership hint', () => {
    const result = normalizeProviderCredentials(JSON.stringify({
      tokens: {
        access_token: 'access',
        refresh_token: 'refresh',
        id_token: 'id',
        account_id: 'acct_1',
      },
    }), 'oauth')

    expect(result).toMatchObject({
      access_token: 'access',
      refresh_token: 'refresh',
      id_token: 'id',
      chatgpt_account_id: 'acct_1',
    })
  })

  it('accepts raw API keys and rejects malformed JSON-looking input', () => {
    expect(normalizeProviderCredentials('sk-test', 'apikey')).toEqual({ api_key: 'sk-test' })
    expect(() => normalizeProviderCredentials('{bad', 'oauth')).toThrow(ProviderCredentialError)
  })

  it('wraps a Google service-account file without changing its JSON', () => {
    const source = { type: 'service_account', project_id: 'demo', client_email: 'svc@example.com', private_key: 'secret' }
    const result = normalizeProviderCredentials(JSON.stringify(source), 'service_account')

    expect(result).toMatchObject({ project_id: 'demo', client_email: 'svc@example.com' })
    expect(JSON.parse(String(result.service_account_json))).toEqual(source)
  })
})

