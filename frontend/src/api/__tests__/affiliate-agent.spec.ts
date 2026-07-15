import { beforeEach, describe, expect, it, vi } from 'vitest'

const client = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('@/api/client', () => ({ apiClient: client }))

import userAPI from '@/api/user'
import affiliatesAPI from '@/api/admin/affiliates'
import zhAffiliate from '@/i18n/locales/zh/affiliate'
import zhAdminAffiliate from '@/i18n/locales/zh/admin/affiliates'

describe('affiliate agent and withdrawal api', () => {
  beforeEach(() => vi.clearAllMocks())

  it('submits a user withdrawal with an idempotency key', async () => {
    const payload = { payment_account_id: 12, amount: 10 }
    client.post.mockResolvedValueOnce({ data: { id: 91, status: 'submitted' } })

    await userAPI.createAffiliateWithdrawal(payload, 'withdrawal-key')

    expect(client.post).toHaveBeenCalledWith('/user/aff/withdrawals', payload, {
      headers: { 'Idempotency-Key': 'withdrawal-key' },
    })
  })

  it('uses an idempotency key when a super administrator changes agent status', async () => {
    client.put.mockResolvedValueOnce({ data: { user_id: 7, status: 'active' } })

    await affiliatesAPI.updateAgentStatus(7, 'active', 'agent-status-key')

    expect(client.put).toHaveBeenCalledWith(
      '/admin/affiliates/agents/7/status',
      { status: 'active' },
      { headers: { 'Idempotency-Key': 'agent-status-key' } },
    )
  })

  it('records the completed offline transfer with the required settlement fields', async () => {
    const settlement = {
      actual_currency: 'CNY',
      actual_amount: 72.5,
      exchange_rate: 7.25,
      external_reference: 'ALIPAY-20260715-001',
    }
    client.post.mockResolvedValueOnce({ data: { id: 91, status: 'paid', ...settlement } })

    await affiliatesAPI.markWithdrawalPaid(91, settlement, 'mark-paid-key')

    expect(client.post).toHaveBeenCalledWith(
      '/admin/affiliates/withdrawals/91/mark-paid',
      settlement,
      { headers: { 'Idempotency-Key': 'mark-paid-key' } },
    )
  })

  it('keeps the requested Chinese transfer wording exact', () => {
    expect(zhAdminAffiliate.withdrawals.markPaid).toBe('标记为已经转账')
    expect(zhAdminAffiliate.withdrawals.statuses.paid).toBe('已经转账')
    expect(zhAffiliate.affiliate.withdrawal.statuses.paid).toBe('已经转账')
  })
})
