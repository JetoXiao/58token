import { describe, expect, it } from 'vitest'
import {
  createUniqueDownloadResourceSlug,
  detectDownloadPlatform,
  extractVersionFromFileName,
  inspectDownloadResourceFile,
  parseWindowsVersionResource,
} from '../downloadResourceMetadata'

describe('download resource metadata', () => {
  it('extracts versions and platforms from installer file names', () => {
    expect(extractVersionFromFileName('CCSwitch-3.8.2-win64.exe')).toBe('3.8.2')
    expect(extractVersionFromFileName('Codex_Setup_v1.12.0.exe')).toBe('1.12.0')
    expect(detectDownloadPlatform('CCSwitch-3.8.2-win64.exe')).toBe('Windows')
    expect(detectDownloadPlatform('client-2.0.1-arm64.dmg')).toBe('macOS')
    expect(detectDownloadPlatform('setup.bat')).toBe('Windows')
    expect(detectDownloadPlatform('guide.pdf')).toBe('Universal')
  })

  it('reads product metadata from a Windows version string block', () => {
    const productName = makeVersionStringBlock('ProductName', 'CCSwitch')
    const productVersion = makeVersionStringBlock('ProductVersion', '3.8.2')
    const joined = new Uint8Array(productName.length + productVersion.length)
    joined.set(productName)
    joined.set(productVersion, productName.length)

    expect(parseWindowsVersionResource(joined.buffer)).toEqual({ productName: 'CCSwitch', version: '3.8.2' })
  })

  it('adds a suffix when an automatically generated slug already exists', () => {
    expect(createUniqueDownloadResourceSlug('ccswitch-windows-3-8-2', ['ccswitch-windows-3-8-2']))
      .toBe('ccswitch-windows-3-8-2-2')
  })

  it('generates useful metadata for non-installer files', async () => {
    const metadata = await inspectDownloadResourceFile(new File(['{}'], 'release-notes-2.4.1.json', {
      type: 'application/json',
    }))

    expect(metadata).toEqual({
      name: 'release notes',
      platform: 'Universal',
      slug: 'release-notes-universal-2-4-1',
      version: '2.4.1',
    })
  })
})

function makeVersionStringBlock(key: string, value: string): Uint8Array {
  const keyBytes = utf16(`${key}\0`)
  const valueBytes = utf16(`${value}\0`)
  const valueOffset = align4(6 + keyBytes.length)
  const block = new Uint8Array(valueOffset + valueBytes.length)
  const view = new DataView(block.buffer)
  view.setUint16(0, block.length, true)
  view.setUint16(2, value.length + 1, true)
  view.setUint16(4, 1, true)
  block.set(keyBytes, 6)
  block.set(valueBytes, valueOffset)
  return block
}

function utf16(value: string): Uint8Array {
  const bytes = new Uint8Array(value.length * 2)
  for (let index = 0; index < value.length; index += 1) {
    const code = value.charCodeAt(index)
    bytes[index * 2] = code & 0xff
    bytes[index * 2 + 1] = code >>> 8
  }
  return bytes
}

function align4(value: number): number {
  return (value + 3) & ~3
}
