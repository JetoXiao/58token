import type { HelpCenterTutorial } from '@/types'

export interface ParsedCodeBlockDraft {
  language: string
  content: string
}

export function parseCodeBlockDraft(raw: string, fallbackLanguage = 'bash'): ParsedCodeBlockDraft {
  const text = raw.trim()
  const fenced = text.match(/^```([A-Za-z0-9_+.-]*)[ \t]*\r?\n([\s\S]*?)\r?\n```$/)
  if (!fenced) {
    return {
      language: fallbackLanguage || 'bash',
      content: raw,
    }
  }
  return {
    language: fenced[1] || fallbackLanguage || 'bash',
    content: fenced[2],
  }
}

export function extractImageSourcesFromHTML(html: string): string[] {
  if (!html) return []
  const doc = new DOMParser().parseFromString(html, 'text/html')
  return Array.from(doc.querySelectorAll('img'))
    .map((image) => image.getAttribute('src') || '')
    .filter(Boolean)
}

export async function dataURLToFile(dataURL: string, filename: string): Promise<File> {
  const response = await fetch(dataURL)
  const blob = await response.blob()
  const extension = blob.type.split('/')[1]?.split('+')[0] || 'png'
  return new File([blob], filename.includes('.') ? filename : `${filename}.${extension}`, { type: blob.type || 'image/png' })
}

export function cloneHelpCenterTutorial(
  tutorial: HelpCenterTutorial,
  existingTutorials: HelpCenterTutorial[],
  sortOrder: number,
  copySuffix: string,
): HelpCenterTutorial {
  const clone = JSON.parse(JSON.stringify(tutorial)) as HelpCenterTutorial
  clone.id = uniqueTutorialCopyID(tutorial.id || 'tutorial', existingTutorials)
  clone.title = copyTitle(tutorial.title, copySuffix)
  clone.sort_order = sortOrder
  return clone
}

function uniqueTutorialCopyID(sourceID: string, existingTutorials: HelpCenterTutorial[]): string {
  const normalizedSourceID = slugifyID(sourceID) || 'tutorial'
  const existingIDs = new Set(existingTutorials.map((tutorial) => tutorial.id).filter(Boolean))
  const baseID = normalizedSourceID.endsWith('-copy') ? normalizedSourceID : `${normalizedSourceID}-copy`
  if (!existingIDs.has(baseID)) return baseID
  for (let index = 2; index < 10000; index += 1) {
    const candidate = `${baseID}-${index}`
    if (!existingIDs.has(candidate)) return candidate
  }
  return `${baseID}-${Date.now().toString(36)}`
}

function slugifyID(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function copyTitle(title: string, suffix: string): string {
  const trimmedTitle = title.trim() || 'Tutorial'
  const trimmedSuffix = suffix.trim() || 'Copy'
  return `${trimmedTitle} ${trimmedSuffix}`
}
