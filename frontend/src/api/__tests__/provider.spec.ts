import { beforeEach, describe, expect, it, vi } from 'vitest'

const client = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('@/api/client', () => ({ apiClient: client }))

import { providerAPI, normalizeProviderUsage } from '@/api/provider'

describe('provider API', () => {
  beforeEach(() => vi.clearAllMocks())

  it('normalizes account pagination when pages is omitted', async () => {
    client.get.mockResolvedValue({ data: { items: [], total: 45, page: 2, page_size: 20 } })

    const result = await providerAPI.listAccounts({ page: 2, page_size: 20, search: 'lab' })

    expect(result.pages).toBe(3)
    expect(client.get).toHaveBeenCalledWith('/provider/accounts', {
      params: { page: 2, page_size: 20, search: 'lab' },
    })
  })

  it('sends only provider-managed fields on create and update', async () => {
    client.post.mockResolvedValue({ data: { id: 1 } })
    client.put.mockResolvedValue({ data: { id: 1 } })

    await providerAPI.createAccount({
      name: '  shared  ',
      platform: 'openai',
      type: 'oauth',
      credentials: { access_token: 'secret' },
      concurrency: 2,
      group_ids: [3, 3],
      priority: 999,
      load_factor: 99,
      provider_id: 7,
    } as never)
    await providerAPI.updateAccount(1, {
      name: ' updated ',
      group_ids: [4, 4],
      rate_multiplier: 9,
      provider_id: null,
    } as never)

    expect(client.post).toHaveBeenCalledWith('/provider/accounts', {
      name: 'shared',
      notes: null,
      platform: 'openai',
      type: 'oauth',
      credentials: { access_token: 'secret' },
      concurrency: 2,
      group_ids: [3],
    })
    expect(client.put).toHaveBeenCalledWith('/provider/accounts/1', {
      name: 'updated',
      group_ids: [4],
    })
  })

  it('normalizes current and legacy usage field names', () => {
    expect(normalizeProviderUsage({
      provider_id: '8',
      start_time: 'start',
      end_time: 'end',
      totals: { requests: 2, input_tokens: 3, output_tokens: 4, cache_creation_tokens: 5, cache_read_tokens: 6 },
      per_account: [{ account_id: 9, account_name: 'A', platform: 'anthropic', requests: 1, input_tokens: 2, output_tokens: 3 }],
    })).toEqual({
      provider_id: 8,
      start_time: 'start',
      end_time: 'end',
      totals: {
        total_requests: 2,
        total_input_tokens: 3,
        total_output_tokens: 4,
        total_cache_creation_tokens: 5,
        total_cache_read_tokens: 6,
        total_tokens: 18,
      },
      accounts: [{
        account_id: 9,
        account_name: 'A',
        platform: 'anthropic',
        total_requests: 1,
        total_input_tokens: 2,
        total_output_tokens: 3,
        total_cache_creation_tokens: 0,
        total_cache_read_tokens: 0,
        total_tokens: 5,
      }],
    })
  })

  it('uses separate self-service and admin usage endpoints', async () => {
    client.get.mockResolvedValue({ data: { totals: {}, accounts: [] } })
    const params = { start_time: 'start', end_time: 'end' }

    await providerAPI.getUsage(params)
    await providerAPI.getAdminUsage(12, params)

    expect(client.get).toHaveBeenNthCalledWith(1, '/provider/usage', { params })
    expect(client.get).toHaveBeenNthCalledWith(2, '/admin/providers/12/usage', { params })
  })
})

