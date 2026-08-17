import { describe, expect, it } from 'vitest'
import { sanitizeSvg } from '@/utils/sanitize'

describe('sanitizeSvg', () => {
  it('keeps inert vector geometry', () => {
    const clean = sanitizeSvg('<svg viewBox="0 0 24 24"><path d="M1 1h4" stroke="currentColor"/></svg>')

    expect(clean).toContain('<svg')
    expect(clean).toContain('<path')
    expect(clean).toContain('stroke="currentColor"')
  })

  it('removes executable and external-resource SVG content', () => {
    const clean = sanitizeSvg(`
      <svg xmlns="http://www.w3.org/2000/svg">
        <script>alert(1)</script>
        <a href="https://attacker.example/phish"><path d="M0 0h1"/></a>
        <image href="https://attacker.example/track" />
        <use xlink:href="https://attacker.example/icons.svg#x" />
        <path style="fill:url(https://attacker.example/pixel)" d="M0 0h2" />
        <path fill="url(https://attacker.example/gradient.svg#x)" d="M0 0h3" />
      </svg>
    `)

    expect(clean).not.toMatch(/script|attacker\.example|xlink:href|style=/i)
    expect(clean).toContain('<path')
  })
})
