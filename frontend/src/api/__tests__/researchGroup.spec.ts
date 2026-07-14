import { beforeEach, describe, expect, it, vi } from 'vitest'

const client = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  patch: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('@/api/client', () => ({ apiClient: client }))

import researchGroupAPI from '@/api/researchGroup'

describe('research group api', () => {
  beforeEach(() => vi.clearAllMocks())

  it('uses the user research-group context and lifecycle endpoints', async () => {
    client.get.mockResolvedValueOnce({ data: null })
    client.post.mockResolvedValue({ data: { role: 'owner' } })
    client.patch.mockResolvedValue({ data: { role: 'owner' } })
    client.delete.mockResolvedValue({ data: undefined })

    await expect(researchGroupAPI.getContext()).resolves.toBeNull()
    await researchGroupAPI.create({ name: 'AI Lab' })
    await researchGroupAPI.update({ status: 'paused' })
    await researchGroupAPI.dissolve()

    expect(client.get).toHaveBeenCalledWith('/user/research-group')
    expect(client.post).toHaveBeenCalledWith('/user/research-group', { name: 'AI Lab' })
    expect(client.patch).toHaveBeenCalledWith('/user/research-group', { status: 'paused' })
    expect(client.delete).toHaveBeenCalledWith('/user/research-group')
  })

  it('uses stable member, invitation, reset, leave, and usage endpoints', async () => {
    client.post.mockResolvedValue({ data: {} })
    client.patch.mockResolvedValue({ data: {} })
    client.delete.mockResolvedValue({ data: undefined })
    client.get.mockResolvedValue({ data: { items: [], total: 0, page: 1, page_size: 20, pages: 0 } })

    await researchGroupAPI.inviteMember({ email: 'student@example.com', monthly_limit_usd: 20 })
    await researchGroupAPI.updateMember(7, { monthly_limit_usd: 30, status: 'active' })
    await researchGroupAPI.resetMemberUsage(7)
    await researchGroupAPI.removeMember(7)
    await researchGroupAPI.acceptInvitation(7)
    await researchGroupAPI.rejectInvitation(8)
    await researchGroupAPI.leave()
    await researchGroupAPI.getUsage({ page: 2, page_size: 20, member_id: 7 })

    expect(client.post).toHaveBeenCalledWith('/user/research-group/members', { email: 'student@example.com', monthly_limit_usd: 20 })
    expect(client.patch).toHaveBeenCalledWith('/user/research-group/members/7', { monthly_limit_usd: 30, status: 'active' })
    expect(client.post).toHaveBeenCalledWith('/user/research-group/members/7/reset')
    expect(client.delete).toHaveBeenCalledWith('/user/research-group/members/7')
    expect(client.post).toHaveBeenCalledWith('/user/research-group/invitations/7/accept')
    expect(client.post).toHaveBeenCalledWith('/user/research-group/invitations/8/reject')
    expect(client.post).toHaveBeenCalledWith('/user/research-group/leave')
    expect(client.get).toHaveBeenCalledWith('/user/research-group/usage', { params: { page: 2, page_size: 20, member_id: 7 } })
  })
})
