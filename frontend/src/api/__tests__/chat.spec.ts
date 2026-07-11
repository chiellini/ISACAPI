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

import { completeChat, generateImage } from '@/api/chat'

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

  it('parses Responses image results nested under response output', async () => {
    fetchMock.mockResolvedValue(new Response(
      [
        'data: {"type":"response.completed","response":{"output":[{"type":"image_generation_call","result":"d2VicA==","output_format":"webp"}]}}',
        '',
        'data: [DONE]',
        '',
      ].join('\n'),
      { status: 200, headers: { 'Content-Type': 'text/event-stream' } },
    ))

    const images = await generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' })

    expect(images).toEqual(['data:image/webp;base64,d2VicA=='])
  })

  it('parses Responses output_item.done image results', async () => {
    fetchMock.mockResolvedValue(new Response(
      [
        'data: {"type":"response.output_item.done","item":{"type":"image_generation_call","result":"anBn","output_format":"jpg"}}',
        '',
        'data: [DONE]',
        '',
      ].join('\n'),
      { status: 200, headers: { 'Content-Type': 'text/event-stream' } },
    ))

    const images = await generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' })

    expect(images).toEqual(['data:image/jpeg;base64,anBn'])
  })

  it('normalizes HTML gateway timeout errors', async () => {
    fetchMock.mockResolvedValue(new Response(
      '<html><body><h1>504 Gateway Time-out</h1></body></html>',
      { status: 504, headers: { 'Content-Type': 'text/html' } },
    ))

    await expect(generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' }))
      .rejects.toThrow('image request failed: 504')
  })

  it('completeChat accumulates streamed deltas into the full reply', async () => {
    localStorage.setItem('auth_token', 'jwt-token')
    fetchMock.mockResolvedValue(new Response(
      [
        'data: {"choices":[{"delta":{"content":"Hello"}}]}',
        '',
        'data: {"choices":[{"delta":{"content":", "}}]}',
        '',
        'data: {"choices":[{"delta":{"content":"world"}}]}',
        '',
        'data: [DONE]',
        '',
      ].join('\n'),
      { status: 200, headers: { 'Content-Type': 'text/event-stream' } },
    ))

    const text = await completeChat({
      model: 'gpt-5.5',
      messages: [{ role: 'user', content: 'hi' }],
    })

    expect(text).toBe('Hello, world')
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/chat/v1/chat/completions',
      expect.objectContaining({ method: 'POST' }),
    )
    const requestBody = JSON.parse(String(fetchMock.mock.calls[0][1].body))
    expect(requestBody.stream).toBe(true)
    expect(requestBody.messages).toEqual([{ role: 'user', content: 'hi' }])
  })
})
