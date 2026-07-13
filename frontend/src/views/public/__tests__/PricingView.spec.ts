import { beforeEach, describe, expect, it } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import PricingView from '../PricingView.vue'
import zh from '@/i18n/locales/zh'
import { useAppStore } from '@/stores/app'
import type { PublicSettings } from '@/types'

function mountView(rate?: number): {
  wrapper: VueWrapper
  appStore: ReturnType<typeof useAppStore>
} {
  const pinia = createPinia()
  setActivePinia(pinia)
  const appStore = useAppStore(pinia)
  appStore.publicSettingsLoaded = true
  if (rate !== undefined) {
    appStore.cachedPublicSettings = {
      balance_recharge_multiplier: rate,
    } as PublicSettings
  }

  const i18n = createI18n({
    legacy: false,
    locale: 'zh',
    fallbackLocale: 'zh',
    messages: { zh },
  })

  const wrapper = mount(PricingView, {
    global: {
      plugins: [pinia, i18n],
      stubs: {
        RouterLink: { template: '<a><slot /></a>' },
        LocaleSwitcher: true,
        ModelIcon: true,
        Icon: true,
      },
    },
  })

  return { wrapper, appStore }
}

describe('PricingView', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
  })

  it('shows every public model as RMB per million using the default multiplier', () => {
    const { wrapper } = mountView()

    expect(wrapper.findAll('[data-testid^="model-card-"]')).toHaveLength(7)
    expect(wrapper.get('[data-testid="recharge-rate"]').text()).toBe('6')
    expect(wrapper.get('[data-testid="input-price-gpt-5.6-sol"]').text()).toContain(
      '¥0.833333',
    )
    expect(wrapper.get('[data-testid="output-price-gpt-5.6-sol"]').text()).toContain('¥5')
    expect(wrapper.get('[data-testid="cache-price-gpt-5.6-sol"]').text()).toContain(
      '¥0.083333',
    )
  })

  it('uses the configured balance recharge multiplier for all displayed prices', () => {
    const { wrapper } = mountView(5)

    expect(wrapper.get('[data-testid="recharge-rate"]').text()).toBe('5')
    expect(wrapper.get('[data-testid="input-price-gpt-5.6-sol"]').text()).toContain('¥1')
    expect(wrapper.get('[data-testid="output-price-gpt-5.6-sol"]').text()).toContain('¥6')
    expect(wrapper.get('[data-testid="cache-price-gpt-5.6-sol"]').text()).toContain('¥0.1')
  })

  it('updates displayed RMB prices when public settings change', async () => {
    const { wrapper, appStore } = mountView(6)

    appStore.cachedPublicSettings = {
      ...(appStore.cachedPublicSettings || {}),
      balance_recharge_multiplier: 4,
    } as PublicSettings
    await nextTick()

    expect(wrapper.get('[data-testid="recharge-rate"]').text()).toBe('4')
    expect(wrapper.get('[data-testid="input-price-gpt-5.6-sol"]').text()).toContain('¥1.25')
    expect(wrapper.get('[data-testid="output-price-gpt-5.6-sol"]').text()).toContain('¥7.5')
  })

  it('does not expose currency, unit, family, search, or sort controls', () => {
    const { wrapper } = mountView()

    expect(wrapper.text()).not.toContain('$')
    expect(wrapper.find('[data-testid="mode-balance"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="mode-cash"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="unit-million"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="unit-thousand"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid^="family-filter-"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="sort-select"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="pricing-search"]').exists()).toBe(false)
  })

  it('renders the correct currency and copyright characters', () => {
    const { wrapper } = mountView()

    expect(wrapper.text()).toContain('¥')
    expect(wrapper.text()).toContain('© 2026 ISACAI')
    expect(wrapper.text()).not.toContain('Â¥')
    expect(wrapper.text()).not.toContain('Â©')
  })
})
