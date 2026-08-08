import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import PricingView from '../PricingView.vue'
import zh from '@/i18n/locales/zh'
import { useAppStore } from '@/stores/app'
import type { PublicPricingModel, PublicSettings } from '@/types'
import { PUBLIC_MODEL_PRICES, type PublicModelPrice } from '@/utils/pricing'

const { getPublicPricingModels } = vi.hoisted(() => ({
  getPublicPricingModels: vi.fn(),
}))

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth')
  return {
    ...actual,
    authAPI: {
      ...(actual.authAPI || {}),
      getPublicPricingModels,
    },
  }
})

const currency = '\u00A5'

function createPricingModelsResponse(
  models: PublicModelPrice[],
): { models: PublicPricingModel[] } {
  return {
    models: models.map((model) => ({
      id: model.id,
      name: model.name,
      family: model.family,
      benchmark_input_usd_per_million: model.benchmarkInputUsdPerMillion,
      benchmark_output_usd_per_million: model.benchmarkOutputUsdPerMillion,
      benchmark_cache_read_usd_per_million: model.benchmarkCacheReadUsdPerMillion,
    })),
  }
}

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
    getPublicPricingModels.mockReset()
    getPublicPricingModels.mockResolvedValue(createPricingModelsResponse(PUBLIC_MODEL_PRICES))
  })

  it('shows every public model as RMB per million using the default multiplier', async () => {
    const { wrapper } = mountView()
    await flushPromises()

    expect(wrapper.findAll('[data-testid^="model-card-"]')).toHaveLength(7)
    expect(wrapper.get('[data-testid="recharge-rate"]').text()).toBe('6')
    expect(wrapper.get('[data-testid="input-price-gpt-5.6-sol"]').text()).toContain(
      `${currency}0.833333`,
    )
    expect(wrapper.get('[data-testid="output-price-gpt-5.6-sol"]').text()).toContain(`${currency}5`)
    expect(wrapper.get('[data-testid="cache-price-gpt-5.6-sol"]').text()).toContain(
      `${currency}0.083333`,
    )
  })

  it('uses the configured balance recharge multiplier for all displayed prices', async () => {
    const { wrapper } = mountView(5)
    await flushPromises()

    expect(wrapper.get('[data-testid="recharge-rate"]').text()).toBe('5')
    expect(wrapper.get('[data-testid="input-price-gpt-5.6-sol"]').text()).toContain(`${currency}1`)
    expect(wrapper.get('[data-testid="output-price-gpt-5.6-sol"]').text()).toContain(`${currency}6`)
    expect(wrapper.get('[data-testid="cache-price-gpt-5.6-sol"]').text()).toContain(`${currency}0.1`)
  })

  it('updates displayed RMB prices when public settings change', async () => {
    const { wrapper, appStore } = mountView(6)
    await flushPromises()

    appStore.cachedPublicSettings = {
      ...(appStore.cachedPublicSettings || {}),
      balance_recharge_multiplier: 4,
    } as PublicSettings
    await nextTick()

    expect(wrapper.get('[data-testid="recharge-rate"]').text()).toBe('4')
    expect(wrapper.get('[data-testid="input-price-gpt-5.6-sol"]').text()).toContain(`${currency}1.25`)
    expect(wrapper.get('[data-testid="output-price-gpt-5.6-sol"]').text()).toContain(`${currency}7.5`)
  })

  it('loads model prices from backend pricing API', async () => {
    getPublicPricingModels.mockResolvedValue({
      models: [
        {
          id: 'gpt-5.7',
          name: 'GPT 5.7',
          family: 'gpt',
          benchmark_input_usd_per_million: 2,
          benchmark_output_usd_per_million: 10,
          benchmark_cache_read_usd_per_million: 0.5,
        },
      ],
    })

    const { wrapper } = mountView()
    await flushPromises()

    expect(wrapper.findAll('[data-testid^="model-card-"]')).toHaveLength(1)
    expect(wrapper.get('[data-testid="model-card-gpt-5.7"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="input-price-gpt-5.7"]').text()).toContain(`${currency}0.333333`)
  })

  it('falls back to static pricing list when backend pricing API fails', async () => {
    getPublicPricingModels.mockRejectedValueOnce(new Error('pricing api failed'))

    const { wrapper } = mountView()
    await flushPromises()

    expect(wrapper.findAll('[data-testid^="model-card-"]')).toHaveLength(PUBLIC_MODEL_PRICES.length)
    expect(wrapper.get('[data-testid="input-price-gpt-5.6-sol"]').text()).toContain(
      `${currency}0.833333`,
    )
  })

  it('does not expose currency, unit, family, search, or sort controls', async () => {
    const { wrapper } = mountView()
    await flushPromises()

    expect(wrapper.text()).not.toContain('$')
    expect(wrapper.find('[data-testid="mode-balance"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="mode-cash"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="unit-million"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="unit-thousand"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid^="family-filter-"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="sort-select"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="pricing-search"]').exists()).toBe(false)
  })

  it('renders the correct currency and copyright characters', async () => {
    const { wrapper } = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain(currency)
    expect(wrapper.text()).toContain('© 2026 ISACAI')
    expect(wrapper.text()).not.toContain('Ã‚Â¥')
    expect(wrapper.text()).not.toContain('Ã‚Â©')
  })
})
