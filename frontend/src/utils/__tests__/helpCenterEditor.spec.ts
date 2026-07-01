import { describe, expect, it } from 'vitest'
import { cloneHelpCenterTutorial, extractImageSourcesFromHTML, parseCodeBlockDraft } from '../helpCenterEditor'
import type { HelpCenterTutorial } from '@/types'

describe('help center editor helpers', () => {
  it('parses fenced markdown code blocks', () => {
    expect(parseCodeBlockDraft('```bash\nexport API_KEY=xxx\ncodex\n```', 'json')).toEqual({
      language: 'bash',
      content: 'export API_KEY=xxx\ncodex',
    })
  })

  it('keeps plain pasted code as content with the fallback language', () => {
    expect(parseCodeBlockDraft('console.log("hello")', 'javascript')).toEqual({
      language: 'javascript',
      content: 'console.log("hello")',
    })
  })

  it('extracts image sources from pasted html', () => {
    expect(extractImageSourcesFromHTML('<p><img src="data:image/png;base64,abc"><img src="/img/a.png"></p>'))
      .toEqual(['data:image/png;base64,abc', '/img/a.png'])
  })

  it('clones a whole tutorial block with a unique id while reusing uploaded file urls', () => {
    const source: HelpCenterTutorial = {
      id: 'codex',
      enabled: true,
      sort_order: 20,
      title: 'Codex',
      badge: 'CLI',
      summary: 'Setup guide',
      content_md: 'Intro',
      steps: [
        {
          title: 'Install',
          description: 'Download the package',
          images: [
            {
              label: 'Screenshot',
              url: '/api/v1/help-center/attachments/screenshot.png',
              file_name: 'screenshot.png',
            },
          ],
          attachments: [
            {
              label: 'Installer',
              url: '/api/v1/help-center/attachments/installer.zip',
              file_name: 'installer.zip',
            },
          ],
          code_blocks: [
            {
              title: 'Command',
              language: 'bash',
              content: 'codex --version',
            },
          ],
        },
      ],
      code_blocks: [
        {
          title: 'Config',
          language: 'json',
          content: '{}',
        },
      ],
      links: [
        {
          label: 'API Keys',
          url: '/keys',
        },
      ],
      attachments: [
        {
          label: 'Full package',
          url: '/api/v1/help-center/attachments/full.zip',
          file_name: 'full.zip',
        },
      ],
    }

    const clone = cloneHelpCenterTutorial(source, [source, { ...source, id: 'codex-copy' }], 30, '副本')

    expect(clone).toEqual({
      ...source,
      id: 'codex-copy-2',
      title: 'Codex 副本',
      sort_order: 30,
    })
    expect(clone).not.toBe(source)
    expect(clone.steps[0]).not.toBe(source.steps[0])
    expect(clone.steps[0].attachments[0].url).toBe(source.steps[0].attachments[0].url)
    expect(clone.attachments[0].url).toBe(source.attachments[0].url)

    clone.steps[0].title = 'Changed'
    clone.attachments[0].label = 'Changed package'

    expect(source.steps[0].title).toBe('Install')
    expect(source.attachments[0].label).toBe('Full package')
  })
})
