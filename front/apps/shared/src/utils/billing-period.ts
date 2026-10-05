/** 业务时区：北京时间 */
export const BEIJING_TZ = 'Asia/Shanghai'

const DATE_ONLY = /^\d{4}-\d{2}-\d{2}$/

/** 北京时间当日 12:00 对应的 instant（避免跨日边界） */
function beijingNoon(y: number, m: number, d: number): Date {
  return new Date(Date.UTC(y, m - 1, d, 4, 0, 0))
}

export function parseSubscriptionDate(s: string): Date | null {
  const raw = s?.trim()
  if (!raw) return null
  const datePart = raw.slice(0, 10)
  if (!DATE_ONLY.test(datePart)) {
    const normalized =
      raw.includes('Z') || /[+-]\d{2}:\d{2}$/.test(raw)
        ? raw
        : `${raw.replace(' ', 'T')}+08:00`
    const d = new Date(normalized)
    return Number.isNaN(d.getTime()) ? null : d
  }
  const [y, m, day] = datePart.split('-').map(Number)
  return beijingNoon(y, m, day)
}

export function resolvePeriodDays(periodDays: number | undefined, billingPeriod: string): number {
  if (periodDays != null && periodDays > 0) return periodDays
  if (billingPeriod === 'year') return 365
  if (billingPeriod === 'once') return 30
  return 30
}

export function addPeriodDays(from: Date, periodDays: number): Date {
  const ymd = formatSubscriptionDate(from)
  const [y, m, day] = ymd.split('-').map(Number)
  return beijingNoon(y, m, day + periodDays)
}

/** 升档计价周期：当前周期 1；expires_at 晚于 period_end 时自 period_end 按 period_days 累加。 */
export function countUpgradeBillingCycles(
  periodEndStr: string,
  expiresAtStr: string,
  periodDays: number,
): number {
  const maxQty = 99
  let qty = 1
  const periodEnd = parseSubscriptionDate(periodEndStr)
  const expiresAt = parseSubscriptionDate(expiresAtStr)
  if (!periodEnd || !expiresAt || expiresAt.getTime() <= periodEnd.getTime()) {
    return qty
  }
  let cursor = periodEnd
  while (cursor.getTime() < expiresAt.getTime() && qty < maxQty) {
    cursor = addPeriodDays(cursor, periodDays)
    qty++
  }
  return qty
}

/** 格式化为北京时间日历日期 YYYY-MM-DD */
export function formatSubscriptionDate(d: Date): string {
  return new Intl.DateTimeFormat('en-CA', { timeZone: BEIJING_TZ }).format(d)
}

/** 当前北京时间，用于展示 */
export function formatNowBeijing(dateStyle: 'date' | 'datetime' = 'datetime'): string {
  const opts: Intl.DateTimeFormatOptions =
    dateStyle === 'date'
      ? { timeZone: BEIJING_TZ, year: 'numeric', month: '2-digit', day: '2-digit' }
      : {
          timeZone: BEIJING_TZ,
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          hour12: false,
        }
  const parts = new Intl.DateTimeFormat('en-CA', opts).formatToParts(new Date())
  const get = (type: Intl.DateTimeFormatPartTypes) =>
    parts.find((p) => p.type === type)?.value ?? ''
  if (dateStyle === 'date') {
    return `${get('year')}-${get('month')}-${get('day')}`
  }
  return `${get('year')}-${get('month')}-${get('day')} ${get('hour')}:${get('minute')}`
}

/** @deprecated 使用 addPeriodDays / resolvePeriodDays */
export function addBillingPeriod(from: Date, billingPeriod: string): Date {
  return addPeriodDays(from, resolvePeriodDays(undefined, billingPeriod))
}
