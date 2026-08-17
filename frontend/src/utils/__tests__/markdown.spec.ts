import { describe, expect, it } from 'vitest'
import { renderMarkdown } from '@/utils/markdown'

function renderContainer(markdown: string): HTMLDivElement {
  const container = document.createElement('div')
  container.innerHTML = renderMarkdown(markdown)
  return container
}

describe('chat Markdown security', () => {
  it('keeps safe prose, code blocks, and tables', () => {
    const container = renderContainer([
      '## Result',
      '',
      '```ts',
      'const answer = 42',
      '```',
      '',
      '| key | value |',
      '| --- | --- |',
      '| a | b |',
    ].join('\n'))

    expect(container.querySelector('h2')?.textContent).toBe('Result')
    expect(container.querySelector('pre code')?.textContent).toContain('const answer = 42')
    expect(container.querySelector('table td')?.textContent).toBe('a')
  })

  it('removes active UI, inline styles, and executable navigation', () => {
    const container = renderContainer([
      '<form action="https://attacker.example"><input><button>Continue</button></form>',
      '<iframe src="https://attacker.example"></iframe>',
      '<a href="javascript:alert(1)" style="position:fixed;inset:0">unsafe</a>',
    ].join('\n'))

    expect(container.querySelector('form, input, button, iframe')).toBeNull()
    expect(container.querySelector('[style]')).toBeNull()
    expect(container.querySelector('a')?.hasAttribute('href')).toBe(false)
  })

  it('drops external tracking images but keeps permitted chat image sources', () => {
    const origin = window.location.origin
    const container = renderContainer([
      '![external](https://tracker.example/pixel.png)',
      '![relative](/api/v1/images/1)',
      `![same-origin](${origin}/api/v1/images/2)`,
      '![data](data:image/png;base64,iVBORw0KGgo=)',
      `![blob](blob:${origin}/8b62fd65-9499-4ad4-a05e-3e10b5eb17ab)`,
      '![svg](data:image/svg+xml,<svg onload=alert(1)></svg>)',
    ].join('\n'))

    const images = [...container.querySelectorAll('img')]
    expect(images.map(image => image.alt)).toEqual(['relative', 'same-origin', 'data', 'blob'])
    expect(images.every(image => !image.src.includes('tracker.example'))).toBe(true)
  })
})
