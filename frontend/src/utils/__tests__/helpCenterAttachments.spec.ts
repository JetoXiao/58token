import { describe, expect, it } from 'vitest'
import {
  getHelpCenterAttachmentAPIPath,
  getHelpCenterAttachmentBrowserURL,
} from '../helpCenterAttachments'

describe('help center attachment url helpers', () => {
  it('converts relative uploaded attachment urls to api client paths', () => {
    expect(getHelpCenterAttachmentAPIPath('/api/v1/help-center/attachments/guide.png'))
      .toBe('/help-center/attachments/guide.png')
  })

  it('converts absolute uploaded attachment urls to api client paths', () => {
    expect(getHelpCenterAttachmentAPIPath('https://example.com/api/v1/help-center/attachments/guide.png?download=1'))
      .toBe('/help-center/attachments/guide.png?download=1')
  })

  it('ignores external non-uploaded image urls', () => {
    expect(getHelpCenterAttachmentAPIPath('https://cdn.example.com/guide.png')).toBeNull()
  })

  it('normalizes absolute uploaded attachment urls to same-origin browser paths', () => {
    expect(getHelpCenterAttachmentBrowserURL('http://127.0.0.1:8080/api/v1/help-center/attachments/guide.png?download=1'))
      .toBe('/api/v1/help-center/attachments/guide.png?download=1')
  })

  it('keeps external non-uploaded browser urls unchanged', () => {
    expect(getHelpCenterAttachmentBrowserURL('https://cdn.example.com/guide.png'))
      .toBe('https://cdn.example.com/guide.png')
  })
})
