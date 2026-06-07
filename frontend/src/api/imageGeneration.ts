export type ImageGenerationQuality = 'low' | 'medium' | 'high' | 'auto'
export type ImageGenerationSize = string

export interface GenerateImageRequest {
  apiKey: string
  model: string
  prompt: string
  size: ImageGenerationSize
  quality: ImageGenerationQuality
  n: number
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

export async function generateImage(request: GenerateImageRequest): Promise<GenerateImageResponse> {
  const response = await fetch('/v1/images/generations', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${request.apiKey}`,
      'Content-Type': 'application/json',
      ...(request.clientRequestId ? { 'X-Client-Request-ID': request.clientRequestId } : {})
    },
    body: JSON.stringify({
      model: request.model.trim(),
      prompt: request.prompt.trim(),
      n: request.n,
      size: request.size,
      quality: request.quality,
      response_format: 'b64_json'
    }),
    signal: request.signal
  })

  if (!response.ok) {
    throw new Error(await extractGatewayError(response))
  }

  const payload = await response.json()
  const images: GeneratedImage[] = Array.isArray(payload?.data)
    ? payload.data
        .map((item: any): GeneratedImage | null => {
          if (typeof item?.b64_json === 'string' && item.b64_json.trim()) {
            return {
              url: normalizeBase64ImageUrl(item.b64_json),
              mimeType: 'image/png',
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

export async function listImageModels(apiKey: string, signal?: AbortSignal): Promise<string[]> {
  const response = await fetch('/v1/models', {
    headers: {
      Authorization: `Bearer ${apiKey}`
    },
    signal
  })

  if (!response.ok) {
    throw new Error(await extractGatewayError(response))
  }

  const payload = await response.json()
  return Array.isArray(payload?.data)
    ? payload.data
        .map((item: any) => (typeof item?.id === 'string' ? item.id.trim() : ''))
        .filter((id: string) => id !== '')
    : []
}

export const imageGenerationAPI = {
  generate: generateImage,
  listModels: listImageModels
}

export default imageGenerationAPI
