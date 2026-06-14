import QRCode from 'qrcode'

interface AffiliatePosterOptions {
  inviteLink: string
  affiliateCode: string
  siteName?: string
}

const POSTER_WIDTH = 1080
const POSTER_HEIGHT = 1440
const FONT_FAMILY = '"Microsoft YaHei", "PingFang SC", "Noto Sans SC", Arial, sans-serif'

type PosterFeature = {
  title: string
  description: string
}

export async function generateAffiliatePoster(options: AffiliatePosterOptions): Promise<Blob> {
  const { inviteLink, affiliateCode, siteName = 'UseAiForMe' } = options
  const canvas = document.createElement('canvas')
  canvas.width = POSTER_WIDTH
  canvas.height = POSTER_HEIGHT

  const ctx = canvas.getContext('2d')
  if (!ctx) {
    throw new Error('Canvas is not supported')
  }

  const qrDataUrl = await QRCode.toDataURL(inviteLink, {
    width: 430,
    margin: 2,
    errorCorrectionLevel: 'H',
    color: {
      dark: '#0f172a',
      light: '#ffffff',
    },
  })

  const [qrImage, logoImage] = await Promise.all([
    loadImage(qrDataUrl),
    loadImage('/logo.png').catch(() => null),
  ])

  drawBackground(ctx)
  drawHeader(ctx, siteName, logoImage)
  drawHero(ctx)
  drawFeaturePanel(ctx)
  drawInvitePanel(ctx, {
    inviteLink,
    affiliateCode,
    qrImage,
    siteName,
  })

  return canvasToBlob(canvas)
}

export function buildAffiliatePosterFilename(affiliateCode: string): string {
  const safeCode = affiliateCode.replace(/[^\w-]/g, '').slice(0, 32) || 'invite'
  return `UseAiForMe-invite-${safeCode}.png`
}

export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 30_000)
}

function drawBackground(ctx: CanvasRenderingContext2D): void {
  const gradient = ctx.createLinearGradient(0, 0, POSTER_WIDTH, POSTER_HEIGHT)
  gradient.addColorStop(0, '#e7fbff')
  gradient.addColorStop(0.38, '#f7fbff')
  gradient.addColorStop(0.72, '#f4f0ff')
  gradient.addColorStop(1, '#ffffff')
  ctx.fillStyle = gradient
  ctx.fillRect(0, 0, POSTER_WIDTH, POSTER_HEIGHT)

  const glowA = ctx.createRadialGradient(80, 330, 0, 80, 330, 560)
  glowA.addColorStop(0, 'rgba(40, 210, 190, 0.22)')
  glowA.addColorStop(1, 'rgba(40, 210, 190, 0)')
  ctx.fillStyle = glowA
  ctx.fillRect(0, 0, POSTER_WIDTH, POSTER_HEIGHT)

  const glowB = ctx.createRadialGradient(1020, 110, 0, 1020, 110, 540)
  glowB.addColorStop(0, 'rgba(126, 96, 255, 0.18)')
  glowB.addColorStop(1, 'rgba(126, 96, 255, 0)')
  ctx.fillStyle = glowB
  ctx.fillRect(0, 0, POSTER_WIDTH, POSTER_HEIGHT)
}

function drawHeader(ctx: CanvasRenderingContext2D, siteName: string, logoImage: HTMLImageElement | null): void {
  withShadow(ctx, 'rgba(15, 23, 42, 0.10)', 34, 12, () => {
    fillRoundedRect(ctx, 72, 70, 936, 156, 36, 'rgba(255, 255, 255, 0.86)')
  })
  strokeRoundedRect(ctx, 72, 70, 936, 156, 36, 'rgba(148, 163, 184, 0.22)', 1)

  const logoX = 112
  const logoY = 108
  const logoSize = 80
  const logoBg = ctx.createLinearGradient(logoX, logoY, logoX + logoSize, logoY + logoSize)
  logoBg.addColorStop(0, '#0f172a')
  logoBg.addColorStop(1, '#155e75')
  fillRoundedRect(ctx, logoX, logoY, logoSize, logoSize, 22, logoBg)

  if (logoImage) {
    ctx.save()
    roundedRectPath(ctx, logoX + 8, logoY + 8, logoSize - 16, logoSize - 16, 16)
    ctx.clip()
    ctx.drawImage(logoImage, logoX + 8, logoY + 8, logoSize - 16, logoSize - 16)
    ctx.restore()
  } else {
    ctx.fillStyle = '#ffffff'
    ctx.font = posterFont(800, 38)
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText('U', logoX + logoSize / 2, logoY + logoSize / 2 + 1)
    ctx.textAlign = 'start'
    ctx.textBaseline = 'alphabetic'
  }

  ctx.fillStyle = '#0f172a'
  ctx.font = posterFont(800, 44)
  ctx.fillText(siteName, 218, 142)
  ctx.fillStyle = '#64748b'
  ctx.font = posterFont(500, 26)
  ctx.fillText('AI API 中转站 | 包月套餐 | 邀请返利', 220, 184)

  drawChip(ctx, 784, 114, '注册送好友', '#ecfdf5', '#047857', 176)
}

function drawHero(ctx: CanvasRenderingContext2D): void {
  ctx.fillStyle = '#0f172a'
  ctx.font = posterFont(900, 70)
  ctx.fillText('把 AI 能力分享给朋友', 72, 330)

  ctx.fillStyle = '#475569'
  ctx.font = posterFont(600, 34)
  drawWrappedText(
    ctx,
    '通过你的专属邀请链接注册，好友可快速开通，你也能获得返利额度。',
    76,
    392,
    900,
    48,
    2,
  )

  drawChip(ctx, 76, 480, 'Codex / Claude / OpenAI', '#dcfce7', '#047857', 292)
  drawChip(ctx, 390, 480, 'USDT / 支付宝', '#eef2ff', '#4338ca', 196)
  drawChip(ctx, 608, 480, '自动绑定邀请码', '#fff1f2', '#be123c', 236)
}

function drawFeaturePanel(ctx: CanvasRenderingContext2D): void {
  withShadow(ctx, 'rgba(15, 23, 42, 0.13)', 38, 18, () => {
    fillRoundedRect(ctx, 72, 568, 936, 462, 42, 'rgba(255, 255, 255, 0.92)')
  })
  strokeRoundedRect(ctx, 72, 568, 936, 462, 42, 'rgba(148, 163, 184, 0.18)', 1)

  ctx.fillStyle = '#0f172a'
  ctx.font = posterFont(800, 38)
  ctx.fillText('为什么选择 UseAiForMe', 118, 640)

  const features: PosterFeature[] = [
    { title: '模型覆盖', description: 'Codex / Claude / OpenAI 常用模型' },
    { title: '灵活充值', description: '支持 USDT / 支付宝，按需使用' },
    { title: '包月套餐', description: '日 / 周 / 月额度，团队使用更稳' },
    { title: '返利绑定', description: '扫码注册后自动绑定你的邀请码' },
  ]

  const positions = [
    [118, 690],
    [560, 690],
    [118, 832],
    [560, 832],
  ] as const

  features.forEach((feature, index) => {
    const [x, y] = positions[index]
    drawFeatureCard(ctx, x, y, feature, index)
  })
}

function drawFeatureCard(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  feature: PosterFeature,
  index: number,
): void {
  fillRoundedRect(ctx, x, y, 402, 112, 24, '#f8fafc')
  strokeRoundedRect(ctx, x, y, 402, 112, 24, 'rgba(226, 232, 240, 0.96)', 1)

  const iconX = x + 28
  const iconY = y + 32
  const colors = ['#10b981', '#3b82f6', '#8b5cf6', '#f97316']
  fillRoundedRect(ctx, iconX, iconY, 48, 48, 14, `${colors[index]}22`)
  ctx.strokeStyle = colors[index]
  ctx.lineWidth = 4
  ctx.lineCap = 'round'
  ctx.beginPath()
  ctx.moveTo(iconX + 15, iconY + 25)
  ctx.lineTo(iconX + 24, iconY + 34)
  ctx.lineTo(iconX + 35, iconY + 16)
  ctx.stroke()

  ctx.fillStyle = '#0f172a'
  ctx.font = posterFont(800, 28)
  ctx.fillText(feature.title, x + 96, y + 46)
  ctx.fillStyle = '#64748b'
  ctx.font = posterFont(500, 21)
  drawWrappedText(ctx, feature.description, x + 96, y + 80, 270, 30, 1)
}

function drawInvitePanel(
  ctx: CanvasRenderingContext2D,
  options: {
    inviteLink: string
    affiliateCode: string
    qrImage: HTMLImageElement
    siteName: string
  },
): void {
  const { inviteLink, affiliateCode, qrImage, siteName } = options
  const highlightGradient = ctx.createLinearGradient(72, 1056, 1008, 1056)
  highlightGradient.addColorStop(0, '#ecfdf5')
  highlightGradient.addColorStop(1, '#fff1f2')
  fillRoundedRect(ctx, 72, 1056, 936, 104, 28, highlightGradient)
  strokeRoundedRect(ctx, 72, 1056, 936, 104, 28, 'rgba(16, 185, 129, 0.28)', 1.5)

  ctx.fillStyle = '#047857'
  ctx.font = posterFont(850, 36)
  ctx.fillText('扫码注册，邀请码自动绑定', 116, 1122)

  ctx.fillStyle = '#475569'
  ctx.font = posterFont(500, 24)
  ctx.fillText('邀请关系会通过专属链接自动记录，好友注册后无需手动填写。', 116, 1156)

  withShadow(ctx, 'rgba(15, 23, 42, 0.15)', 40, 18, () => {
    fillRoundedRect(ctx, 72, 1190, 936, 218, 36, '#ffffff')
  })
  strokeRoundedRect(ctx, 72, 1190, 936, 218, 36, 'rgba(148, 163, 184, 0.20)', 1)

  const qrX = 104
  const qrY = 1208
  const qrSize = 184
  fillRoundedRect(ctx, qrX - 10, qrY - 10, qrSize + 20, qrSize + 20, 24, '#f8fafc')
  ctx.drawImage(qrImage, qrX, qrY, qrSize, qrSize)

  ctx.fillStyle = '#0f172a'
  ctx.font = posterFont(850, 32)
  ctx.fillText(`扫码加入 ${siteName}`, 336, 1256)

  ctx.fillStyle = '#64748b'
  ctx.font = posterFont(500, 22)
  drawWrappedText(ctx, getPosterLinkText(inviteLink, affiliateCode), 336, 1304, 570, 31, 2)

  fillRoundedRect(ctx, 336, 1352, 332, 40, 20, '#ecfdf5')
  ctx.fillStyle = '#047857'
  ctx.font = posterFont(800, 22)
  ctx.fillText(`邀请码：${affiliateCode}`, 358, 1379)

  ctx.fillStyle = '#94a3b8'
  ctx.font = posterFont(500, 18)
  ctx.fillText('AI 能力稳定接入，适合个人与团队持续使用', 76, 1430)
}

function getPosterLinkText(inviteLink: string, affiliateCode: string): string {
  try {
    const url = new URL(inviteLink)
    return `${url.host}${url.pathname}?aff=${affiliateCode}`
  } catch {
    return inviteLink
  }
}

function drawChip(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  label: string,
  background: string,
  color: string,
  width: number,
): void {
  fillRoundedRect(ctx, x, y, width, 48, 24, background)
  ctx.fillStyle = color
  ctx.font = posterFont(800, 23)
  ctx.textAlign = 'center'
  ctx.fillText(label, x + width / 2, y + 32)
  ctx.textAlign = 'start'
}

function drawWrappedText(
  ctx: CanvasRenderingContext2D,
  text: string,
  x: number,
  y: number,
  maxWidth: number,
  lineHeight: number,
  maxLines: number,
): void {
  const chars = Array.from(text)
  let line = ''
  let lineCount = 0

  for (let index = 0; index < chars.length; index += 1) {
    const char = chars[index]
    const testLine = `${line}${char}`
    const isLastLine = lineCount === maxLines - 1

    if (ctx.measureText(testLine).width > maxWidth && line) {
      if (isLastLine) {
        ctx.fillText(trimLineToWidth(ctx, `${line}...`, maxWidth), x, y + lineCount * lineHeight)
        return
      }
      ctx.fillText(line, x, y + lineCount * lineHeight)
      line = char
      lineCount += 1
      continue
    }

    line = testLine
  }

  if (line && lineCount < maxLines) {
    ctx.fillText(line, x, y + lineCount * lineHeight)
  }
}

function trimLineToWidth(ctx: CanvasRenderingContext2D, line: string, maxWidth: number): string {
  let output = line
  while (output.length > 1 && ctx.measureText(output).width > maxWidth) {
    output = `${output.slice(0, -4)}...`
  }
  return output
}

function withShadow(
  ctx: CanvasRenderingContext2D,
  color: string,
  blur: number,
  offsetY: number,
  draw: () => void,
): void {
  ctx.save()
  ctx.shadowColor = color
  ctx.shadowBlur = blur
  ctx.shadowOffsetY = offsetY
  draw()
  ctx.restore()
}

function fillRoundedRect(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  radius: number,
  fillStyle: string | CanvasGradient,
): void {
  ctx.save()
  roundedRectPath(ctx, x, y, width, height, radius)
  ctx.fillStyle = fillStyle
  ctx.fill()
  ctx.restore()
}

function strokeRoundedRect(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  radius: number,
  strokeStyle: string,
  lineWidth: number,
): void {
  ctx.save()
  roundedRectPath(ctx, x, y, width, height, radius)
  ctx.strokeStyle = strokeStyle
  ctx.lineWidth = lineWidth
  ctx.stroke()
  ctx.restore()
}

function roundedRectPath(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  radius: number,
): void {
  const safeRadius = Math.min(radius, width / 2, height / 2)
  ctx.beginPath()
  ctx.moveTo(x + safeRadius, y)
  ctx.lineTo(x + width - safeRadius, y)
  ctx.quadraticCurveTo(x + width, y, x + width, y + safeRadius)
  ctx.lineTo(x + width, y + height - safeRadius)
  ctx.quadraticCurveTo(x + width, y + height, x + width - safeRadius, y + height)
  ctx.lineTo(x + safeRadius, y + height)
  ctx.quadraticCurveTo(x, y + height, x, y + height - safeRadius)
  ctx.lineTo(x, y + safeRadius)
  ctx.quadraticCurveTo(x, y, x + safeRadius, y)
  ctx.closePath()
}

function posterFont(weight: number, size: number): string {
  return `${weight} ${size}px ${FONT_FAMILY}`
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error(`Failed to load image: ${src}`))
    image.src = src
  })
}

function canvasToBlob(canvas: HTMLCanvasElement): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (blob) {
        resolve(blob)
      } else {
        reject(new Error('Failed to export poster image'))
      }
    }, 'image/png')
  })
}
