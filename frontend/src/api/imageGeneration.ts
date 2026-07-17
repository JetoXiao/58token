export type ImageGenerationQuality = 'low' | 'medium' | 'high' | 'auto'
export type ImageGenerationSize = string
export type ImageGenerationPlatform = 'openai' | 'gemini' | 'antigravity'

export const openAIImageModels = ['gpt-image-2', 'gpt-image-1.5', 'gpt-image-1'] as const
export const geminiImageModels = ['gemini-3.1-flash-image', 'gemini-3-pro-image', 'gemini-2.5-flash-image'] as const

const GEMINI_FLASH_REFERENCE_LIMIT = 3
const GEMINI_PRO_REFERENCE_LIMIT = 14
const OPENAI_REFERENCE_LIMIT = 16

export interface GenerateImageRequest {
  apiKey: string
  model: string
  prompt: string
  size: ImageGenerationSize
  quality: ImageGenerationQuality
  n: number
  platform?: ImageGenerationPlatform | string
  referenceImages?: File[]
  clientRequestId?: string
  signal?: AbortSignal
}

export interface GeneratedImage {
  url: string
  mimeType: string
  revisedPrompt?: string
}

export interface GenerateImageResponse {
  images: GeneratedImage[]
  created?: number
}

const extractGatewayError = async (response: Response): Promise<string> => {
  const fallback = `${response.status} ${response.statusText}`.trim()
  try {
    const payload = await response.json()
    const message =
      payload?.error?.message ||
      payload?.message ||
      payload?.detail ||
      payload?.error ||
      fallback
    return typeof message === 'string' ? message : fallback
  } catch {
    return fallback
  }
}

const normalizeBase64ImageUrl = (raw: string, mimeType = 'image/png'): string => {
  const trimmed = raw.trim()
  if (trimmed.startsWith('data:image/')) return trimmed
  return `data:${mimeType};base64,${trimmed}`
}

export const isGeminiImageModel = (model: string): boolean =>
  /^gemini-.+-image(?:-preview)?$/i.test(model.trim())

export const referenceImageLimitForModel = (model: string, platform?: ImageGenerationPlatform | string): number => {
  const normalized = model.trim().toLowerCase()
  if (
    platform !== 'antigravity' &&
    (normalized === 'gemini-3-pro-image' || normalized === 'gemini-3-pro-image-preview')
  ) {
    return GEMINI_PRO_REFERENCE_LIMIT
  }
  if (isGeminiImageModel(normalized)) return GEMINI_FLASH_REFERENCE_LIMIT
  return OPENAI_REFERENCE_LIMIT
}

const parseOpenAIImageResponse = (payload: any): GenerateImageResponse => {
  const images: GeneratedImage[] = Array.isArray(payload?.data)
    ? payload.data
        .map((item: any): GeneratedImage | null => {
          const mimeType = typeof item?.mime_type === 'string' && item.mime_type.trim()
            ? item.mime_type.trim()
            : typeof item?.output_format === 'string' && item.output_format.trim()
              ? `image/${item.output_format.trim().toLowerCase()}`
              : 'image/png'
          if (typeof item?.b64_json === 'string' && item.b64_json.trim()) {
            return {
              url: normalizeBase64ImageUrl(item.b64_json, mimeType),
              mimeType,
              revisedPrompt: typeof item.revised_prompt === 'string' ? item.revised_prompt : undefined
            }
          }
          if (typeof item?.url === 'string' && item.url.trim()) {
            const url = item.url.trim()
            const separatorIndex = url.indexOf(';')
            return {
              url,
              mimeType: url.startsWith('data:image/') && separatorIndex > 5 ? url.slice(5, separatorIndex) : 'image/*',
              revisedPrompt: typeof item.revised_prompt === 'string' ? item.revised_prompt : undefined
            }
          }
          return null
        })
        .filter((item: GeneratedImage | null): item is GeneratedImage => item !== null)
    : []

  return {
    images,
    created: typeof payload?.created === 'number' ? payload.created : undefined
  }
}

const generateOpenAIImage = async (request: GenerateImageRequest): Promise<GenerateImageResponse> => {
  const referenceImages = request.referenceImages || []
  const editing = referenceImages.length > 0
  const headers: Record<string, string> = {
    Authorization: `Bearer ${request.apiKey}`,
    ...(request.clientRequestId ? { 'X-Client-Request-ID': request.clientRequestId } : {})
  }
  let body: BodyInit

  if (editing) {
    const form = new FormData()
    form.append('model', request.model.trim())
    form.append('prompt', request.prompt.trim())
    form.append('n', String(request.n))
    form.append('size', request.size)
    form.append('quality', request.quality)
    form.append('response_format', 'b64_json')
    referenceImages.forEach((file) => {
      form.append(referenceImages.length === 1 ? 'image' : 'image[]', file, file.name)
    })
    body = form
  } else {
    headers['Content-Type'] = 'application/json'
    body = JSON.stringify({
      model: request.model.trim(),
      prompt: request.prompt.trim(),
      n: request.n,
      size: request.size,
      quality: request.quality,
      response_format: 'b64_json'
    })
  }

  const response = await fetch(editing ? '/v1/images/edits' : '/v1/images/generations', {
    method: 'POST',
    headers,
    body,
    signal: request.signal
  })

  if (!response.ok) {
    throw new Error(await extractGatewayError(response))
  }

  return parseOpenAIImageResponse(await response.json())
}

const fileToBase64 = async (file: File): Promise<string> => {
  const buffer = await file.arrayBuffer()
  const bytes = new Uint8Array(buffer)
  const chunkSize = 0x8000
  let binary = ''
  for (let offset = 0; offset < bytes.length; offset += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(offset, offset + chunkSize))
  }
  return btoa(binary)
}

const geminiImageConfigForSize = (model: string, size: ImageGenerationSize): Record<string, string> => {
  const match = /^(\d+)x(\d+)$/i.exec(size.trim())
  if (!match) return {}
  const width = Number(match[1])
  const height = Number(match[2])
  if (!width || !height) return {}

  const ratios = [
    ['1:1', 1], ['5:4', 5 / 4], ['4:5', 4 / 5], ['4:3', 4 / 3], ['3:4', 3 / 4],
    ['3:2', 3 / 2], ['2:3', 2 / 3], ['16:9', 16 / 9], ['9:16', 9 / 16], ['21:9', 21 / 9]
  ] as const
  const target = width / height
  const aspectRatio = ratios.reduce((best, candidate) =>
    Math.abs(Math.log(candidate[1] / target)) < Math.abs(Math.log(best[1] / target)) ? candidate : best
  )[0]
  const maxSide = Math.max(width, height)
  const imageSize = maxSide <= 1024 ? '1K' : maxSide <= 2048 ? '2K' : '4K'
  const config: Record<string, string> = { aspectRatio }
  if (!/^gemini-2\.5-flash-image(?:-preview)?$/i.test(model.trim())) {
    config.imageSize = imageSize
  }
  return config
}

const generateGeminiImage = async (request: GenerateImageRequest): Promise<GenerateImageResponse> => {
  const imageParts = await Promise.all((request.referenceImages || []).map(async (file) => ({
    inlineData: {
      data: await fileToBase64(file),
      mimeType: file.type || 'image/png'
    }
  })))
  const platform = request.platform === 'antigravity' ? 'antigravity' : 'gemini'
  const prefix = platform === 'antigravity' ? '/antigravity/v1beta' : '/v1beta'
  const response = await fetch(`${prefix}/models/${encodeURIComponent(request.model.trim())}:generateContent`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${request.apiKey}`,
      'Content-Type': 'application/json',
      ...(request.clientRequestId ? { 'X-Client-Request-ID': request.clientRequestId } : {})
    },
    body: JSON.stringify({
      contents: [{
        role: 'user',
        parts: [{ text: request.prompt.trim() }, ...imageParts]
      }],
      generationConfig: {
        responseModalities: ['TEXT', 'IMAGE'],
        imageConfig: geminiImageConfigForSize(request.model, request.size)
      }
    }),
    signal: request.signal
  })

  if (!response.ok) {
    throw new Error(await extractGatewayError(response))
  }

  const payload = await response.json()
  const parts = Array.isArray(payload?.candidates)
    ? payload.candidates.flatMap((candidate: any) => Array.isArray(candidate?.content?.parts) ? candidate.content.parts : [])
    : []
  const images = parts
    .map((part: any): GeneratedImage | null => {
      const inlineData = part?.inlineData || part?.inline_data
      const data = typeof inlineData?.data === 'string' ? inlineData.data.trim() : ''
      if (!data) return null
      const mimeType = typeof inlineData?.mimeType === 'string'
        ? inlineData.mimeType
        : typeof inlineData?.mime_type === 'string'
          ? inlineData.mime_type
          : 'image/png'
      return { url: normalizeBase64ImageUrl(data, mimeType), mimeType }
    })
    .filter((item: GeneratedImage | null): item is GeneratedImage => item !== null)

  return {
    images
  }
}

export async function generateImage(request: GenerateImageRequest): Promise<GenerateImageResponse> {
  return isGeminiImageModel(request.model)
    ? generateGeminiImage(request)
    : generateOpenAIImage(request)
}

export async function listImageModels(apiKey: string, signal?: AbortSignal, platform: ImageGenerationPlatform | string = 'openai'): Promise<string[]> {
  const endpoint = platform === 'gemini'
    ? '/v1beta/models'
    : platform === 'antigravity'
      ? '/antigravity/v1beta/models'
      : '/v1/models'
  const response = await fetch(endpoint, {
    headers: {
      Authorization: `Bearer ${apiKey}`
    },
    signal
  })

  if (!response.ok) {
    throw new Error(await extractGatewayError(response))
  }

  const payload = await response.json()
  const openAIModels = Array.isArray(payload?.data)
    ? payload.data.map((item: any) => (typeof item?.id === 'string' ? item.id.trim() : ''))
    : []
  const geminiModels = Array.isArray(payload?.models)
    ? payload.models.map((item: any) => {
        const name = typeof item?.name === 'string' ? item.name.trim() : ''
        return name.replace(/^models\//, '')
      })
    : []
  return [...new Set([...openAIModels, ...geminiModels].filter((id: string) => id !== ''))]
}

export const imageGenerationAPI = {
  generate: generateImage,
  listModels: listImageModels
}

export default imageGenerationAPI
