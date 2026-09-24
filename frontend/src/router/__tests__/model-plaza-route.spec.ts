import { describe, expect, it } from 'vitest'
import routerSource from '@/router/index.ts?raw'

describe('model plaza route wiring', () => {
  it('registers the public model plaza page and its settings guard', () => {
    expect(routerSource).toContain("path: '/model-plaza'")
    expect(routerSource).toContain("component: () => import('@/views/ModelPlazaView.vue')")
    expect(routerSource).toContain("titleKey: 'modelPlaza.title'")
    expect(routerSource).toContain("if (to.path === '/model-plaza')")
  })

  it('redirects the retired pricing page to the model plaza', () => {
    expect(routerSource).toContain("path: '/pricing'")
    expect(routerSource).toContain("redirect: '/model-plaza'")
    expect(routerSource).not.toContain("component: () => import('@/views/public/PricingView.vue')")
  })
})
