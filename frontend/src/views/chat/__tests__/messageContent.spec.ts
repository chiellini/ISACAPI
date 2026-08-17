import { describe, expect, it } from 'vitest'
import { buildApiMessageContent } from '../messageContent'

describe('chat message content policy', () => {
  const attachments = [
    { kind: 'text' as const, name: 'notes.txt', text: 'trusted notes' },
    { kind: 'image' as const, name: 'diagram.png', dataUrl: 'data:image/png;base64,AA==' },
  ]

  it('includes image parts only when the selected profile supports vision', () => {
    expect(buildApiMessageContent('hello', attachments, true)).toEqual([
      { type: 'text', text: 'hello\n\n[notes.txt]\ntrusted notes' },
      { type: 'image_url', image_url: { url: 'data:image/png;base64,AA==' } },
    ])
  })

  it('replaces historical images with an honest marker for non-vision profiles', () => {
    const content = buildApiMessageContent('hello', attachments, false)
    expect(content).toBe(
      'hello\n\n[notes.txt]\ntrusted notes\n\n[Image attachment not sent to this model: diagram.png]',
    )
    expect(JSON.stringify(content)).not.toContain('data:image')
  })
})
