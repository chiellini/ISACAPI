import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub, type VueWrapper } from '@vue/test-utils'
import { nextTick } from 'vue'

import HomeView from '../HomeView.vue'
import homeViewSource from '../HomeView.vue?raw'

const { authStore, appStore } = vi.hoisted(() => ({
  authStore: {
    isAuthenticated: false,
    isAdmin: false,
    user: null,
    checkAuth: vi.fn(),
  },
  appStore: {
    cachedPublicSettings: null as { balance_recharge_multiplier?: number } | null,
    siteName: 'ISACAI',
    siteLogo: '',
    docUrl: '',
    publicSettingsLoaded: true,
    fetchPublicSettings: vi.fn(),
  },
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => authStore,
  useAppStore: () => appStore,
}))

vi.mock('@/utils/featureFlags', () => ({
  FeatureFlags: { publicStatus: {} },
  isFeatureFlagEnabled: () => false,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) =>
        key === 'home.pricing.rechargeValue' ? `${key}:${String(params?.usd ?? '')}` : key,
      locale: { value: 'zh' },
    }),
  }
})

const NOTICE_VERSION = '2026-07-pricing-v2'
const NOTICE_SESSION_KEY = `isacai_home_notice_session_${NOTICE_VERSION}`
const NOTICE_DATE_KEY = `isacai_home_notice_date_${NOTICE_VERSION}`

function mountHome(): VueWrapper {
  return mount(HomeView, {
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        LocaleSwitcher: true,
        ModelIcon: true,
        ModelPriceComparison: true,
        Icon: true,
      },
    },
  })
}

function hasNotice(wrapper: VueWrapper): boolean {
  return wrapper.findAll('h2').some((heading) => heading.text() === 'home.notice.title')
}

function buttonWithText(wrapper: VueWrapper, text: string) {
  const button = wrapper.findAll('button').find((candidate) => candidate.text().trim() === text)
  if (!button) {
    throw new Error(`Expected button with text: ${text}`)
  }
  return button
}

function localTodayKey(): string {
  const now = new Date()
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}

describe('HomeView public navigation', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
    authStore.isAuthenticated = false
    authStore.isAdmin = false
    authStore.user = null
    authStore.checkAuth.mockReset()
    appStore.cachedPublicSettings = null
    appStore.fetchPublicSettings.mockReset()

    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: vi.fn().mockReturnValue({ matches: false }),
    })
  })

  it('exposes pricing, authentication, and CC-Switch journeys to signed-out visitors', () => {
    const wrapper = mountHome()
    const routerDestinations = wrapper
      .findAllComponents(RouterLinkStub)
      .map((link) => link.props('to'))
    const externalDestinations = wrapper
      .findAll('a[href]')
      .map((link) => link.attributes('href'))

    expect(routerDestinations).toContain('/pricing')
    expect(routerDestinations).toContain('/login')
    expect(routerDestinations).toContain('/register')
    expect(routerDestinations).toContain('/cc-switch')
    expect(externalDestinations).toContain('https://ccswitch.io/')

    wrapper.unmount()
  })

  it('does not introduce a rankings route that the application does not provide', () => {
    const wrapper = mountHome()
    const renderedDestinations = [
      ...wrapper.findAllComponents(RouterLinkStub).map((link) => link.props('to')),
      ...wrapper.findAll('a[href]').map((link) => link.attributes('href')),
    ]

    expect(renderedDestinations).not.toContain('/rankings')
    expect(homeViewSource).not.toContain('/rankings')

    wrapper.unmount()
  })

  it('uses the configured public recharge multiplier across the homepage', () => {
    appStore.cachedPublicSettings = { balance_recharge_multiplier: 5 }
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('1 : 5')
    expect(wrapper.text()).toContain('home.pricing.rechargeValue:5')
    expect(wrapper.getComponent({ name: 'ModelPriceComparison' }).props('usdPerCny')).toBe(5)

    wrapper.unmount()
  })
})

describe('HomeView notice persistence', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
    authStore.isAuthenticated = false
    authStore.isAdmin = false
    authStore.user = null
    appStore.cachedPublicSettings = null

    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: vi.fn().mockReturnValue({ matches: false }),
    })
  })

  it('lets the header bell reopen a notice closed for the current session', async () => {
    sessionStorage.setItem(NOTICE_SESSION_KEY, '1')
    const wrapper = mountHome()
    await nextTick()
    expect(hasNotice(wrapper)).toBe(false)

    await wrapper.get('[data-testid="home-notice-bell"]').trigger('click')
    expect(hasNotice(wrapper)).toBe(true)

    wrapper.unmount()
  })

  it('keeps a session dismissal effective after the view is mounted again', async () => {
    const wrapper = mountHome()
    await nextTick()
    expect(hasNotice(wrapper)).toBe(true)

    await buttonWithText(wrapper, 'home.notice.close').trigger('click')
    expect(sessionStorage.getItem(NOTICE_SESSION_KEY)).toBe('1')
    wrapper.unmount()

    const remountedWrapper = mountHome()
    await nextTick()
    expect(hasNotice(remountedWrapper)).toBe(false)
    remountedWrapper.unmount()
  })

  it('persists a same-day dismissal and respects it on the next mount', async () => {
    const wrapper = mountHome()
    await nextTick()

    await buttonWithText(wrapper, 'home.notice.closeToday').trigger('click')
    expect(localStorage.getItem(NOTICE_DATE_KEY)).toBe(localTodayKey())
    wrapper.unmount()

    const remountedWrapper = mountHome()
    await nextTick()
    expect(hasNotice(remountedWrapper)).toBe(false)
    remountedWrapper.unmount()
  })
})
