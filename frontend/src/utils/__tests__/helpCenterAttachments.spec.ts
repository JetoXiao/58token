import { describe, expect, it } from 'vitest'
import { getHelpCenterAttachmentAPIPath } from '../helpCenterAttachments'

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
})
