import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ChatPolicyView from '../ChatPolicyView.vue'

const {
  getPolicy,
  updatePolicy,
  getGroups,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getPolicy: vi.fn(),
  updatePolicy: vi.fn(),
  getGroups: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    chatPolicy: {
      get: getPolicy,
      update: updatePolicy,
    },
    groups: {
      getAll: getGroups,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

const storedPolicy = {
  enabled: true,
  profiles: [{
    id: 'gpt',
    name: 'GPT',
    provider: 'openai',
    public_model: 'gpt',
    upstream_model: 'gpt-5',
    group_id: 1,
    system_prompt: 'trusted',
    skill_ids: [],
    enabled: true,
    default: true,
    capabilities: { vision: false, image: false, web_search: false },
  }],
  skills: [],
}

function mountView() {
  return mount(ChatPolicyView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
      },
    },
  })
}

describe('admin ChatPolicyView load guard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getGroups.mockResolvedValue([])
    updatePolicy.mockImplementation(async (policy: unknown) => policy)
  })

  it('keeps save disabled after a failed load and enables it only after retry succeeds', async () => {
    getPolicy
      .mockRejectedValueOnce(new Error('load failed'))
      .mockResolvedValueOnce(storedPolicy)

    const wrapper = mountView()
    await flushPromises()

    const save = wrapper.get('[data-testid="save-chat-policy"]')
    expect(save.attributes('disabled')).toBeDefined()
    await save.trigger('click')
    expect(updatePolicy).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('load failed')

    await wrapper.get('[data-testid="retry-chat-policy"]').trigger('click')
    await flushPromises()

    expect(getPolicy).toHaveBeenCalledTimes(2)
    expect(save.attributes('disabled')).toBeUndefined()
    await save.trigger('click')
    await flushPromises()
    expect(updatePolicy).toHaveBeenCalledTimes(1)
  })
})
