export function formatMarketplaceMultiplier(label: string, multiplier: number): string {
  const normalizedLabel = label.trim() || '倍率'
  const numericMultiplier = Number(multiplier)

  if (!Number.isFinite(numericMultiplier) || numericMultiplier <= 0) {
    return `${normalizedLabel} -`
  }

  const formattedMultiplier = numericMultiplier
    .toFixed(6)
    .replace(/\.?0+$/, '')

  return `${normalizedLabel} ${formattedMultiplier}x`
}
