import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  generateImage,
  listImageModels,
  referenceImageLimitForModel
} from '@/api/imageGeneration'

const okJSON = (payload: unknown) => new Response(JSON.stringify(payload), {
  status: 200,
  headers: { 'Content-Type': 'application/json' }
})

const baseRequest = {
  apiKey: 'sk-test',
  model: 'gpt-image-2',
  prompt: 'Draw a precise product image',
  size: '1024x1024',
  quality: 'high' as const,
  n: 1
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('imageGeneration API', () => {
  it('uses the OpenAI generations endpoint when no reference image is supplied', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(okJSON({
      data: [{ b64_json: 'aW1hZ2U=' }]
    }))

    const result = await generateImage(baseRequest)

    expect(fetchMock).toHaveBeenCalledOnce()
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/v1/images/generations')
    expect(init?.headers).toMatchObject({ 'Content-Type': 'application/json' })
    expect(JSON.parse(String(init?.body))).toMatchObject({ model: 'gpt-image-2', n: 1 })
    expect(result.images[0].url).toBe('data:image/png;base64,aW1hZ2U=')
  })

  it('uses multipart OpenAI edits and preserves every reference image', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(okJSON({
      data: [{ b64_json: 'ZWRpdA==', output_format: 'webp' }]
    }))
    const first = new File(['first'], 'first.png', { type: 'image/png' })
    const second = new File(['second'], 'second.webp', { type: 'image/webp' })

    const result = await generateImage({
      ...baseRequest,
      referenceImages: [first, second]
    })

    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/v1/images/edits')
    expect(init?.body).toBeInstanceOf(FormData)
    const form = init?.body as FormData
    expect(form.get('model')).toBe('gpt-image-2')
    expect(form.getAll('image[]')).toHaveLength(2)
    expect(result.images[0]).toMatchObject({
      url: 'data:image/webp;base64,ZWRpdA==',
      mimeType: 'image/webp'
    })
  })

  it('sends Gemini reference images as inlineData and parses image parts', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(okJSON({
      candidates: [{
        content: {
          parts: [
            { text: 'Done' },
            { inlineData: { data: 'Z2VtaW5p', mimeType: 'image/jpeg' } }
          ]
        }
      }]
    }))
    const file = new File(['reference'], 'reference.jpg', { type: 'image/jpeg' })
    Object.defineProperty(file, 'arrayBuffer', {
      value: async () => new TextEncoder().encode('reference').buffer
    })

    const result = await generateImage({
      ...baseRequest,
      model: 'gemini-2.5-flash-image',
      platform: 'gemini',
      referenceImages: [file]
    })

    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/v1beta/models/gemini-2.5-flash-image:generateContent')
    const payload = JSON.parse(String(init?.body))
    expect(payload.contents[0].parts).toEqual([
      { text: baseRequest.prompt },
      { inlineData: { data: 'cmVmZXJlbmNl', mimeType: 'image/jpeg' } }
    ])
    expect(payload.generationConfig.imageConfig).toEqual({ aspectRatio: '1:1' })
    expect(result.images[0]).toMatchObject({
      url: 'data:image/jpeg;base64,Z2VtaW5p',
      mimeType: 'image/jpeg'
    })
  })

  it('uses the Antigravity Gemini endpoint and discovers native model names', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(okJSON({ models: [{ name: 'models/gemini-3.1-flash-image' }] }))
      .mockResolvedValueOnce(okJSON({ candidates: [] }))

    await expect(listImageModels('sk-test', undefined, 'antigravity')).resolves.toEqual([
      'gemini-3.1-flash-image'
    ])
    await generateImage({
      ...baseRequest,
      model: 'gemini-3.1-flash-image',
      platform: 'antigravity',
      size: '3840x2160'
    })

    expect(fetchMock.mock.calls[0][0]).toBe('/antigravity/v1beta/models')
    expect(fetchMock.mock.calls[1][0]).toBe('/antigravity/v1beta/models/gemini-3.1-flash-image:generateContent')
    const payload = JSON.parse(String(fetchMock.mock.calls[1][1]?.body))
    expect(payload.generationConfig.imageConfig).toEqual({ aspectRatio: '16:9', imageSize: '4K' })
  })

  it('applies provider-specific reference image limits', () => {
    expect(referenceImageLimitForModel('gpt-image-2')).toBe(16)
    expect(referenceImageLimitForModel('gemini-2.5-flash-image')).toBe(3)
    expect(referenceImageLimitForModel('gemini-3.1-flash-image')).toBe(3)
    expect(referenceImageLimitForModel('gemini-3-pro-image')).toBe(14)
    expect(referenceImageLimitForModel('gemini-3-pro-image', 'antigravity')).toBe(3)
  })
})
