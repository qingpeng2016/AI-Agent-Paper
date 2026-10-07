export function parseMoney(v: number | string | undefined | null): number {
  if (v == null || v === '') return 0
  if (typeof v === 'number') return Number.isFinite(v) ? v : 0
  const n = parseFloat(String(v).replace(/,/g, ''))
  return Number.isFinite(n) ? n : 0
}

/** 格式化为人民币展示（元，最多两位小数） */
export function formatCny(amount: number | string): string {
  const n = parseMoney(amount)
  const cents = Math.round(n * 100)
  if (cents % 100 === 0) {
    return `¥${Math.trunc(n)}`
  }
  return `¥${n.toFixed(2)}`
}

/** 资金流水等：正数带 +，负数带 -，零不带符号 */
export function formatSignedCny(amount: number | string): string {
  const n = parseMoney(amount)
  const absText = formatCny(Math.abs(n))
  if (n > 0) return `+ ${absText}`
  if (n < 0) return `- ${absText}`
  return absText
}

export function signedMoneyClass(amount: number | string): 'amount-plus' | 'amount-minus' | 'amount-zero' {
  const n = parseMoney(amount)
  if (n > 0) return 'amount-plus'
  if (n < 0) return 'amount-minus'
  return 'amount-zero'
}

/** @deprecated 使用 formatCny（金额已为元） */
export function formatCnyFromCents(cents: number): string {
  return formatCny(cents / 100)
}

export function formatTokenCount(n: number): string {
  if (n >= 1_000_000) {
    return `${(n / 1_000_000).toFixed(n % 1_000_000 === 0 ? 0 : 1)}M`
  }
  if (n >= 1_000) {
    return `${(n / 1_000).toFixed(n % 1_000 === 0 ? 0 : 1)}K`
  }
  return String(n)
}

export function billingPeriodLabel(period: 'month' | 'year' | 'once'): string {
  switch (period) {
    case 'month':
      return '/ 月'
    case 'year':
      return '/ 年'
    case 'once':
      return '一次性'
    default:
      return ''
  }
}
