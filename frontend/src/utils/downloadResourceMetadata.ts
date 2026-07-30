const MAX_PE_HEADER_BYTES = 4 * 1024 * 1024
const MAX_RESOURCE_SECTION_BYTES = 64 * 1024 * 1024

export interface DownloadResourceFileMetadata {
  name: string
  platform: string
  slug: string
  version: string
}

interface WindowsExecutableMetadata {
  productName: string
  version: string
}

export async function inspectDownloadResourceFile(file: File): Promise<DownloadResourceFileMetadata> {
  const fallbackVersion = extractVersionFromFileName(file.name)
  const platform = detectDownloadPlatform(file.name)
  let executableMetadata: WindowsExecutableMetadata = { productName: '', version: '' }

  if (/\.exe$/i.test(file.name)) {
    try {
      executableMetadata = await inspectWindowsExecutable(file)
    } catch {
      // Installers without a standard PE resource table still use filename fallbacks.
    }
  }

  const version = executableMetadata.version || fallbackVersion
  const name = executableMetadata.productName || displayNameFromFileName(file.name, version)
  return {
    name,
    platform,
    version,
    slug: createDownloadResourceSlug(name, platform, version, file.name),
  }
}

export function createUniqueDownloadResourceSlug(candidate: string, existingSlugs: Iterable<string>): string {
  const occupied = new Set(Array.from(existingSlugs, (value) => value.toLowerCase()))
  const base = candidate || `download-${Date.now().toString(36)}`
  if (!occupied.has(base)) return base

  for (let suffix = 2; suffix < 10_000; suffix += 1) {
    const suffixText = `-${suffix}`
    const next = `${base.slice(0, 63 - suffixText.length).replace(/-+$/g, '')}${suffixText}`
    if (!occupied.has(next)) return next
  }
  return `${base.slice(0, 54).replace(/-+$/g, '')}-${Date.now().toString(36)}`
}

export function extractVersionFromFileName(fileName: string): string {
  const baseName = stripFileExtension(fileName)
  const match = baseName.match(/(?:^|[-_\s])v?(\d+\.\d+(?:\.\d+){0,2})(?=$|[-_\s])/i)
  return match?.[1] || ''
}

export function detectDownloadPlatform(fileName: string): string {
  const lowerName = fileName.toLowerCase()
  if (/\.(exe|msi|msix|bat|cmd|ps1)$/.test(lowerName) || /(?:^|[-_.\s])(win|windows|win32|win64)(?:$|[-_.\s])/.test(lowerName)) return 'Windows'
  if (/\.(dmg|pkg)$/.test(lowerName) || /(?:^|[-_.\s])(mac|macos|darwin)(?:$|[-_.\s])/.test(lowerName)) return 'macOS'
  if (/\.(deb|rpm|appimage)$/.test(lowerName) || /(?:^|[-_.\s])linux(?:$|[-_.\s])/.test(lowerName)) return 'Linux'
  if (/\.(sh|bash|zsh)$/.test(lowerName)) return 'Linux / macOS'
  return 'Universal'
}

export function parseWindowsVersionResource(buffer: ArrayBuffer): WindowsExecutableMetadata {
  const bytes = new Uint8Array(buffer)
  const view = new DataView(buffer)
  const productVersion = readVersionString(bytes, view, 'ProductVersion')
  const fileVersion = readVersionString(bytes, view, 'FileVersion')
  const productName = readVersionString(bytes, view, 'ProductName')
  return {
    productName,
    version: normalizeVersion(productVersion || fileVersion || readFixedFileVersion(view)),
  }
}

async function inspectWindowsExecutable(file: File): Promise<WindowsExecutableMetadata> {
  let headerBuffer = await file.slice(0, Math.min(file.size, 1024 * 1024)).arrayBuffer()
  let headerView = new DataView(headerBuffer)
  if (headerView.byteLength < 64 || headerView.getUint16(0, true) !== 0x5a4d) return { productName: '', version: '' }

  const peOffset = headerView.getUint32(0x3c, true)
  if (peOffset + 24 > headerView.byteLength) {
    const requiredBytes = Math.min(file.size, Math.min(MAX_PE_HEADER_BYTES, peOffset + 24 + 40 * 96))
    headerBuffer = await file.slice(0, requiredBytes).arrayBuffer()
    headerView = new DataView(headerBuffer)
  }
  if (peOffset + 24 > headerView.byteLength || headerView.getUint32(peOffset, true) !== 0x00004550) return { productName: '', version: '' }

  const sectionCount = headerView.getUint16(peOffset + 6, true)
  const optionalHeaderSize = headerView.getUint16(peOffset + 20, true)
  const sectionTableOffset = peOffset + 24 + optionalHeaderSize
  const sectionTableEnd = sectionTableOffset + sectionCount * 40
  if (sectionCount <= 0 || sectionCount > 96 || sectionTableEnd > headerView.byteLength) return { productName: '', version: '' }

  for (let index = 0; index < sectionCount; index += 1) {
    const sectionOffset = sectionTableOffset + index * 40
    const sectionName = readASCII(headerView, sectionOffset, 8)
    if (!sectionName.startsWith('.rsrc')) continue

    const rawSize = headerView.getUint32(sectionOffset + 16, true)
    const rawOffset = headerView.getUint32(sectionOffset + 20, true)
    if (!rawSize || rawOffset >= file.size) return { productName: '', version: '' }
    const end = Math.min(file.size, rawOffset + Math.min(rawSize, MAX_RESOURCE_SECTION_BYTES))
    return parseWindowsVersionResource(await file.slice(rawOffset, end).arrayBuffer())
  }
  return { productName: '', version: '' }
}

function readVersionString(bytes: Uint8Array, view: DataView, key: string): string {
  const encodedKey = encodeUTF16LE(`${key}\0`)
  for (let offset = 6; offset <= bytes.length - encodedKey.length; offset += 2) {
    if (!bytesEqualAt(bytes, encodedKey, offset)) continue

    const blockOffset = offset - 6
    const blockLength = view.getUint16(blockOffset, true)
    const valueLength = view.getUint16(blockOffset + 2, true)
    const type = view.getUint16(blockOffset + 4, true)
    const valueOffset = blockOffset + align4(6 + encodedKey.length)
    const blockEnd = Math.min(bytes.length, blockOffset + blockLength)
    if (type !== 1 || valueLength <= 0 || blockLength <= 6 || valueOffset >= blockEnd) continue

    const maxCharacters = Math.min(valueLength, Math.floor((blockEnd - valueOffset) / 2))
    let value = ''
    for (let character = 0; character < maxCharacters; character += 1) {
      const code = view.getUint16(valueOffset + character * 2, true)
      if (code === 0) break
      value += String.fromCharCode(code)
    }
    value = value.trim()
    if (value) return value
  }
  return ''
}

function readFixedFileVersion(view: DataView): string {
  for (let offset = 0; offset + 24 <= view.byteLength; offset += 4) {
    if (view.getUint32(offset, true) !== 0xfeef04bd) continue
    const product = versionParts(view.getUint32(offset + 16, true), view.getUint32(offset + 20, true))
    if (product.some((part) => part !== 0)) return product.join('.')
    return versionParts(view.getUint32(offset + 8, true), view.getUint32(offset + 12, true)).join('.')
  }
  return ''
}

function versionParts(ms: number, ls: number): number[] {
  return [(ms >>> 16) & 0xffff, ms & 0xffff, (ls >>> 16) & 0xffff, ls & 0xffff]
}

function displayNameFromFileName(fileName: string, version: string): string {
  let name = stripFileExtension(fileName)
  if (version) name = name.replace(new RegExp(`(?:^|[-_\\s])v?${escapeRegExp(version)}(?=$|[-_\\s])`, 'i'), ' ')
  name = name
    .replace(/(?:^|[-_.\s])(setup|installer|install|portable|windows|win32|win64|x86|x64|amd64|arm64)(?=$|[-_.\s])/gi, ' ')
    .replace(/[-_]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
  return name || stripFileExtension(fileName)
}

function createDownloadResourceSlug(name: string, platform: string, version: string, fileName: string): string {
  const source = [name || stripFileExtension(fileName), platform, version].filter(Boolean).join('-')
  const slug = source
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 63)
    .replace(/-+$/g, '')
  return slug || `download-${Date.now().toString(36)}`
}

function stripFileExtension(fileName: string): string {
  return fileName.replace(/\.(?:tar\.(?:gz|bz2|xz)|[^.]+)$/i, '')
}

function readASCII(view: DataView, offset: number, length: number): string {
  let value = ''
  for (let index = 0; index < length; index += 1) {
    const code = view.getUint8(offset + index)
    if (code === 0) break
    value += String.fromCharCode(code)
  }
  return value
}

function encodeUTF16LE(value: string): Uint8Array {
  const bytes = new Uint8Array(value.length * 2)
  for (let index = 0; index < value.length; index += 1) {
    const code = value.charCodeAt(index)
    bytes[index * 2] = code & 0xff
    bytes[index * 2 + 1] = code >>> 8
  }
  return bytes
}

function bytesEqualAt(bytes: Uint8Array, expected: Uint8Array, offset: number): boolean {
  for (let index = 0; index < expected.length; index += 1) {
    if (bytes[offset + index] !== expected[index]) return false
  }
  return true
}

function align4(value: number): number {
  return (value + 3) & ~3
}

function normalizeVersion(value: string): string {
  return value.trim().replace(/^[vV](?=\d)/, '').replace(/\s+/g, '')
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}
