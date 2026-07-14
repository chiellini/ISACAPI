import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import ResearchGroupFundingCard from '../ResearchGroupFundingCard.vue'
import type { ResearchGroupContext } from '@/types'

const mocks = vi.hoisted(() => ({
  accept: vi.fn(),
  reject: vi.fn(),
  refreshUser: vi.fn(),
  showSuccess: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/api/researchGroup', () => ({
  default: {
    acceptInvitation: mocks.accept,
    rejectInvitation: mocks.reject,
  },
}))

vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: mocks.showSuccess, showError: mocks.showError }) }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ refreshUser: mocks.refreshUser }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

function context(status: 'pending' | 'active' | 'paused'): ResearchGroupContext {
  return {
    role: 'member',
    group: {
      id: 3,
      name: 'AI Lab',
      status: 'active',
      owner_user_id: 1,
      owner_email: 'owner@example.com',
      owner_username: 'Professor',
      created_at: '2026-07-01T00:00:00Z',
      updated_at: '2026-07-01T00:00:00Z',
    },
    member: {
      id: 9,
      research_group_id: 3,
      user_id: 2,
      email: 'student@example.com',
      username: 'Student',
      status,
      monthly_limit_usd: 40,
      monthly_usage_usd: 15,
      monthly_reserved_usd: 5,
      monthly_remaining_usd: 20,
      usage_window_start: '2026-07-01T00:00:00Z',
      resets_at: '2026-08-01T00:00:00Z',
    },
  }
}

describe('ResearchGroupFundingCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.accept.mockResolvedValue({})
    mocks.reject.mockResolvedValue(undefined)
    mocks.refreshUser.mockResolvedValue(undefined)
  })

  it('accepts a pending invitation and refreshes the auth context', async () => {
    const wrapper = mount(ResearchGroupFundingCard, {
      props: { context: context('pending'), personalBalance: 12 },
      global: { stubs: { RouterLink: RouterLinkStub } },
    })

    const accept = wrapper.findAll('button').find((button) => button.text() === 'researchGroup.invitations.accept')
    expect(accept).toBeDefined()
    await accept!.trigger('click')
    await flushPromises()

    expect(mocks.accept).toHaveBeenCalledWith(9)
    expect(mocks.refreshUser).toHaveBeenCalledOnce()
    expect(wrapper.emitted('updated')).toHaveLength(1)
  })

  it('shows both monthly group funding and personal balance for an active member', () => {
    const wrapper = mount(ResearchGroupFundingCard, {
      props: { context: context('active'), personalBalance: 12 },
      global: { stubs: { RouterLink: RouterLinkStub } },
    })

    expect(wrapper.text()).toContain('researchGroup.dashboard.monthlyRemaining')
    expect(wrapper.text()).toContain('researchGroup.dashboard.personalBalance')
    expect(wrapper.text()).toContain('$20.00')
    expect(wrapper.text()).toContain('$12.00')
    expect(wrapper.get('[style]').attributes('style')).toContain('50%')
  })
})
