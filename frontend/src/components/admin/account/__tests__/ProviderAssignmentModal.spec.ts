import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProviderAssignmentModal from '../ProviderAssignmentModal.vue'

const { listUsers, assignProvider, showError, showSuccess } = vi.hoisted(() => ({
  listUsers: vi.fn(),
  assignProvider: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: { list: listUsers },
    accounts: { assignProvider }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

function pageResult(items: Array<Record<string, unknown>>, overrides: Record<string, unknown> = {}) {
  return { items, total: items.length, page: 1, page_size: 100, pages: 1, ...overrides }
}

function mountModal() {
  return mount(ProviderAssignmentModal, {
    props: {
      show: false,
      account: { id: 4, name: 'gpt my d2y', provider_id: null } as any
    },
    global: {
      stubs: {
        BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
      }
    }
  })
}

describe('ProviderAssignmentModal', () => {
  beforeEach(() => {
    listUsers.mockReset()
    assignProvider.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('lists active users of both provider and admin_provider roles', async () => {
    listUsers.mockImplementation(async (_page: number, _size: number, filters: { role?: string }) => {
      if (filters?.role === 'provider') {
        return pageResult([{ id: 7, username: 'pure-provider', email: 'p@example.com' }])
      }
      if (filters?.role === 'admin_provider') {
        return pageResult([{ id: 9, username: 'admin-provider', email: 'ap@example.com' }])
      }
      return pageResult([])
    })

    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    const queriedRoles = listUsers.mock.calls.map(([, , filters]) => filters?.role)
    expect(queriedRoles).toContain('provider')
    expect(queriedRoles).toContain('admin_provider')
    expect(listUsers.mock.calls.every(([, , filters]) => filters?.status === 'active')).toBe(true)

    const options = wrapper.findAll('option').map((o) => o.text())
    expect(options.some((text) => text.includes('pure-provider'))).toBe(true)
    expect(options.some((text) => text.includes('admin-provider'))).toBe(true)
    expect(showError).not.toHaveBeenCalled()
  })

  it('collects every page of a paginated role query', async () => {
    listUsers.mockImplementation(async (page: number, _size: number, filters: { role?: string }) => {
      if (filters?.role === 'provider') {
        return page === 1
          ? pageResult([{ id: 1, username: 'p1', email: 'p1@example.com' }], { pages: 2, total: 2 })
          : pageResult([{ id: 2, username: 'p2', email: 'p2@example.com' }], { pages: 2, total: 2, page: 2 })
      }
      return pageResult([])
    })

    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    const options = wrapper.findAll('option').map((o) => o.text())
    expect(options.some((text) => text.includes('p1'))).toBe(true)
    expect(options.some((text) => text.includes('p2'))).toBe(true)
  })
})
