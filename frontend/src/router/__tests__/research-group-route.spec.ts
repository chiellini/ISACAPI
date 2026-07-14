import { describe, expect, it } from 'vitest'
import routerSource from '@/router/index.ts?raw'
import sidebarSource from '@/components/layout/AppSidebar.vue?raw'
import researchGroupViewSource from '@/views/user/ResearchGroupView.vue?raw'

describe('research group navigation', () => {
  it('registers an authenticated user route with translated metadata', () => {
    expect(routerSource).toContain("path: '/research-group'")
    expect(routerSource).toContain("name: 'ResearchGroup'")
    expect(routerSource).toContain("import('@/views/user/ResearchGroupView.vue')")
    expect(routerSource).toContain("titleKey: 'researchGroup.title'")
    expect(routerSource).toContain("'/research-group'")
  })

  it('adds the route to regular user navigation but hides it for admins and simple mode', () => {
    expect(sidebarSource).toContain("path: '/research-group'")
    expect(sidebarSource).toContain("label: t('nav.researchGroup')")
    expect(sidebarSource).toContain('hideInSimpleMode: true')
    expect(sidebarSource).toContain('featureFlag: flagResearchGroup')
    expect(sidebarSource).toContain('const flagResearchGroup = () => !authStore.isAdmin')
  })

  it('does not require student contexts to expose the owner balance', () => {
    expect(researchGroupViewSource).toContain('context.group.owner_balance ?? 0')
  })
})
