import { describe, expect, it } from 'vitest'
import routerSource from '@/router/index.ts?raw'

describe('model plaza route wiring', () => {
  it('registers the public model plaza page and its settings guard', () => {
    expect(routerSource).toContain("path: '/model-plaza'")
    expect(routerSource).toContain("component: () => import('@/views/ModelPlazaView.vue')")
    expect(routerSource).toContain("titleKey: 'modelPlaza.title'")
    expect(routerSource).toContain("if (to.path === '/model-plaza')")
  })
})
