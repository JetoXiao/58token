const UPLOADED_ATTACHMENT_PREFIX = '/api/v1/help-center/attachments/'

export function isUploadedHelpCenterAttachment(url: string): boolean {
  if (!url) return false
  if (url.startsWith(UPLOADED_ATTACHMENT_PREFIX)) return true
  try {
    return new URL(url, window.location.origin).pathname.startsWith(UPLOADED_ATTACHMENT_PREFIX)
  } catch {
    return false
  }
}

export function getHelpCenterAttachmentAPIPath(url: string): string | null {
  if (!isUploadedHelpCenterAttachment(url)) return null
  try {
    const parsed = new URL(url, window.location.origin)
    return `${parsed.pathname.slice('/api/v1'.length)}${parsed.search}`
  } catch {
    if (url.startsWith(UPLOADED_ATTACHMENT_PREFIX)) {
      return url.slice('/api/v1'.length)
    }
    return null
  }
}
