// Formatting helpers. Two rules worth keeping in mind:
//   - durations arrive as whole seconds, never as a pre-formatted string;
//   - a calendar day arrives as "YYYY-MM-DD" and must not go through
//     new Date(), which reads it as UTC midnight and so renders as the previous
//     day in any negative-offset zone.

/** "2 ч 15 м", "45 м", "—" for nothing. */
export function formatDuration(seconds: number): string {
  if (seconds <= 0) return '—'
  const totalMinutes = Math.round(seconds / 60)
  const hours = Math.floor(totalMinutes / 60)
  const minutes = totalMinutes % 60
  if (hours === 0) return `${minutes} м`
  if (minutes === 0) return `${hours} ч`
  return `${hours} ч ${minutes} м`
}

/** Compact form for a tile: "2:15". */
export function formatClock(seconds: number): string {
  const totalMinutes = Math.max(0, Math.round(seconds / 60))
  const hours = Math.floor(totalMinutes / 60)
  const minutes = totalMinutes % 60
  return `${hours}:${String(minutes).padStart(2, '0')}`
}

export function formatPercent(value: number): string {
  return `${Number.isInteger(value) ? value : value.toFixed(1)}%`
}

/** Russian plural agreement — "1 день", "2 дня", "5 дней". */
export function pluralRu(n: number, one: string, few: string, many: string): string {
  const mod100 = Math.abs(n) % 100
  const mod10 = mod100 % 10
  if (mod100 >= 11 && mod100 <= 14) return many
  if (mod10 === 1) return one
  if (mod10 >= 2 && mod10 <= 4) return few
  return many
}

export function formatStreak(days: number): string {
  return `${days} ${pluralRu(days, 'день', 'дня', 'дней')}`
}

/** Renders an ISO instant as the user's own wall clock, hours and minutes. */
export function formatTimeOfDay(iso: string): string {
  // The string already carries the user's offset, so slicing beats parsing:
  // new Date() would convert it into the viewer's local zone, which is not
  // necessarily the zone the numbers were bucketed in.
  const match = /T(\d{2}:\d{2})/.exec(iso)
  return match ? match[1] : ''
}
