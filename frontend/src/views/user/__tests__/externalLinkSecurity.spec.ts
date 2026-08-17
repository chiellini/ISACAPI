import { describe, expect, it } from 'vitest'
import chatViewSource from '@/views/chat/ChatView.vue?raw'
import customPageSource from '@/views/user/CustomPageView.vue?raw'
import paymentQrSource from '@/views/user/PaymentQRCodeView.vue?raw'
import { sanitizeUrl } from '@/utils/url'

describe('external-link security boundaries', () => {
  it('rejects executable and non-web URL schemes', () => {
    expect(sanitizeUrl('javascript:alert(1)')).toBe('')
    expect(sanitizeUrl('data:text/html,<script>alert(1)</script>')).toBe('')
    expect(sanitizeUrl('//attacker.example/path')).toBe('')
    expect(sanitizeUrl('https://safe.example/path')).toBe('https://safe.example/path')
  })

  it('sandboxes custom-page embeds without forwarding the user bearer token', () => {
    const iframe = customPageSource.slice(
      customPageSource.indexOf('<iframe'),
      customPageSource.indexOf('</iframe>'),
    )
    const embeddedUrlBuilder = customPageSource.slice(
      customPageSource.indexOf('return buildEmbeddedUrl('),
      customPageSource.indexOf('\n  )', customPageSource.indexOf('return buildEmbeddedUrl(')),
    )

    expect(iframe).toContain('sandbox="allow-downloads allow-forms allow-popups allow-scripts"')
    expect(iframe).toContain('referrerpolicy="no-referrer"')
    expect(iframe).not.toContain('allow-same-origin')
    expect(embeddedUrlBuilder).not.toContain('authStore.token')
  })

  it('sanitizes search result URLs before rendering or persisting them', () => {
    expect(chatViewSource).toContain("const url = typeof s.url === 'string' ? sanitizeUrl(s.url) : ''")
    expect(chatViewSource).toContain('.map(normalizeStoredSource)')
  })

  it('sanitizes payment redirect query parameters before binding them to href', () => {
    expect(paymentQrSource).toContain("payUrl.value = sanitizeUrl(String(route.query.pay_url || ''))")
  })
})
