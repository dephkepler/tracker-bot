// The period presets, modelled on the reference app's finance/period.ts but
// collapsed to what a phone-sized control can hold.
//
// The server owns the arithmetic: it knows the user's timezone and its own
// definition of "week", and it echoes the window it used back in meta.from and
// meta.to. So this file carries labels and a query string, never dates — two
// implementations of "the last 30 days" would eventually disagree.

export type PeriodKey = 'today' | 'week' | 'month' | 'year'

export const PERIODS: Array<{ value: PeriodKey; label: string }> = [
  { value: 'today', label: 'Сегодня' },
  { value: 'week', label: '7 дней' },
  { value: 'month', label: '30 дней' },
  { value: 'year', label: 'Год' },
]

export const DEFAULT_PERIOD: PeriodKey = 'month'

/**
 * Granularity for the series. Only "today" is stated explicitly, and only to be
 * unambiguous that a single day is an hourly profile rather than one lonely
 * daily bar; the rest is left to the server's auto rule.
 */
export function seriesGranularity(period: PeriodKey): string {
  return period === 'today' ? 'hour' : 'day'
}

export function periodLabel(period: PeriodKey): string {
  return PERIODS.find((p) => p.value === period)?.label ?? period
}
