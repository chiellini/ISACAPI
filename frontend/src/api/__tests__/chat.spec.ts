import { beforeEach, describe, expect, it, vi } from 'vitest'

const apiClient = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient,
}))

import { generateImage } from '@/api/chat'

describe('chat api', () => {
  const fetchMock = vi.fn()

  beforeEach(() => {
    localStorage.clear()
    fetchMock.mockReset()
    vi.stubGlobal('fetch', fetchMock)
  })

  it('requests image generation as a stream and parses JSON fallback responses', async () => {
    localStorage.setItem('auth_token', 'jwt-token')
    fetchMock.mockResolvedValue(new Response(
      JSON.stringify({ data: [{ b64_json: 'aGVsbG8=' }] }),
      { status: 200, headers: { 'Content-Type': 'application/json' } },
    ))

    const images = await generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' })

    expect(images).toEqual(['data:image/png;base64,aGVsbG8='])
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/chat/v1/images/generations',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          Authorization: 'Bearer jwt-token',
        }),
      }),
    )
    const requestBody = JSON.parse(String(fetchMock.mock.calls[0][1].body))
    expect(requestBody).toEqual({
      model: 'gpt-image-2',
      prompt: 'draw a cat',
      stream: true,
      response_format: 'b64_json',
    })
  })

  it('parses image generation SSE and ignores partial images', async () => {
    fetchMock.mockResolvedValue(new Response(
      [
        'event: image_generation.partial_image',
        'data: {"type":"image_generation.partial_image","b64_json":"cGFydGlhbA=="}',
        '',
        'event: image_generation.completed',
        'data: {"type":"image_generation.completed","b64_json":"ZmluYWw="}',
        '',
        'data: [DONE]',
        '',
      ].join('\n'),
      { status: 200, headers: { 'Content-Type': 'text/event-stream' } },
    ))

    const images = await generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' })

    expect(images).toEqual(['data:image/png;base64,ZmluYWw='])
  })

  it('normalizes HTML gateway timeout errors', async () => {
    fetchMock.mockResolvedValue(new Response(
      '<html><body><h1>504 Gateway Time-out</h1></body></html>',
      { status: 504, headers: { 'Content-Type': 'text/html' } },
    ))

    await expect(generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' }))
      .rejects.toThrow('image request failed: 504')
  })
})
