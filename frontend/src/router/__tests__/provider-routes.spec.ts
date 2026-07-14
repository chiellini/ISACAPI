import { describe, expect, it } from 'vitest'
import routerSource from '@/router/index.ts?raw'
import sidebarSource from '@/components/layout/AppSidebar.vue?raw'
import formSource from '@/components/provider/ProviderAccountFormModal.vue?raw'
import en from '@/i18n/locales/en/provider'
import zh from '@/i18n/locales/zh/provider'

describe('provider navigation and isolation', () => {
  it('registers provider self-service and admin usage routes with role guards', () => {
    expect(routerSource).toContain("path: '/provider/accounts'")
    expect(routerSource).toContain("path: '/provider/usage'")
    expect(routerSource).toContain("path: '/admin/providers/:id/usage'")
    expect(routerSource).toContain('requiresProvider: true')
    expect(routerSource).toContain("if (authStore.isProvider && !requiresProvider && to.path !== '/profile')")
  })

  it('shows a dedicated provider sidebar and keeps scheduling controls out of the form', () => {
    expect(sidebarSource).toContain('v-else-if="isProvider"')
    expect(sidebarSource).toContain("path: '/provider/accounts'")
    expect(sidebarSource).toContain("path: '/provider/usage'")
    expect(formSource).not.toContain('priority')
    expect(formSource).not.toContain('load_factor')
    expect(formSource).not.toContain('rate_multiplier')
    expect(formSource).not.toContain('provider_id')
  })

  it('ships matching English and Chinese provider usage notices', () => {
    expect(en.provider.usage.platformOnlyNotice).toContain('official Claude or Codex desktop apps')
    expect(zh.provider.usage.platformOnlyNotice).toContain('官方 Claude 或 Codex 桌面端')
    expect(en.provider.admin.viewUsage).toBeTruthy()
    expect(zh.provider.admin.viewUsage).toBeTruthy()
  })
})

