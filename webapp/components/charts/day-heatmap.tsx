'use client'

import { useState } from 'react'
import { formatDuration } from '@/lib/format'
import type { HeatDay } from '@/lib/api-types'

// The calendar: one cell per day, colour by how much was tracked.
//
// Magnitude by colour is the one place a sequential ramp belongs, and its four
// steps were validated against both surfaces. Days with nothing get the plain
// inset colour rather than the lightest step, so "nothing" never reads as "a
// little".
//
// Dates are handled as strings throughout. new Date("2026-08-21") is UTC
// midnight and renders as the 20th in any negative-offset zone, which in a
// calendar is a visibly wrong cell.

const RAMP = ['var(--seq-1)', 'var(--seq-2)', 'var(--seq-3)', 'var(--seq-4)']
const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

/** Days in a "YYYY-MM-DD" range, walked as UTC to keep it pure arithmetic. */
function daysBetween(from: string, to: string): string[] {
  const out: string[] = []
  const start = new Date(`${from}T00:00:00Z`)
  const end = new Date(`${to}T00:00:00Z`)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return out
  for (let d = start; d <= end; d = new Date(d.getTime() + 86400000)) {
    out.push(d.toISOString().slice(0, 10))
  }
  return out
}

/** Monday-first weekday index of a date string. */
function weekdayIndex(date: string): number {
  const day = new Date(`${date}T00:00:00Z`).getUTCDay()
  return (day + 6) % 7
}

export function DayHeatmap({
  days,
  maxSeconds,
  from,
  to,
  onPickDay,
}: {
  days: HeatDay[]
  maxSeconds: number
  from: string
  to: string
  onPickDay?: (date: string) => void
}) {
  const [picked, setPicked] = useState<string | null>(null)

  const byDate = new Map(days.map((d) => [d.date, d.seconds]))
  const all = daysBetween(from, to)
  if (all.length === 0) return <p className='text-body text-ink-3'>Нет данных за период.</p>

  // Pad the first week so every column is one weekday.
  const leading = weekdayIndex(all[0])
  const cells: Array<string | null> = [...Array<null>(leading).fill(null), ...all]

  function stepFor(seconds: number): string {
    if (seconds <= 0) return 'var(--surface-2)'
    if (maxSeconds <= 0) return RAMP[0]
    const idx = Math.min(RAMP.length - 1, Math.floor((seconds / maxSeconds) * RAMP.length))
    return RAMP[idx]
  }

  const pickedSeconds = picked ? (byDate.get(picked) ?? 0) : 0

  return (
    <div>
      <div className='overflow-x-auto'>
        <div className='flex gap-1'>
          <div className='flex flex-col gap-1 pr-1'>
            {WEEKDAYS.map((w, i) => (
              <div key={w} className='flex h-4 items-center text-micro text-ink-3'>
                {i % 2 === 0 ? w : ''}
              </div>
            ))}
          </div>

          <div className='grid grid-flow-col grid-rows-7 gap-1'>
            {cells.map((date, i) =>
              date === null ? (
                <div key={`pad-${i}`} className='h-4 w-4' />
              ) : (
                <button
                  key={date}
                  type='button'
                  className='h-4 w-4 rounded-mark'
                  style={{
                    backgroundColor: stepFor(byDate.get(date) ?? 0),
                    outline: picked === date ? '2px solid var(--ink)' : undefined,
                    outlineOffset: picked === date ? '1px' : undefined,
                  }}
                  title={`${date} — ${formatDuration(byDate.get(date) ?? 0)}`}
                  aria-label={`${date} — ${formatDuration(byDate.get(date) ?? 0)}`}
                  onClick={() => {
                    setPicked(date)
                    onPickDay?.(date)
                  }}
                />
              )
            )}
          </div>
        </div>
      </div>

      <div className='mt-3 flex items-center justify-between gap-3'>
        <div className='flex items-center gap-1.5 text-micro text-ink-3'>
          <span>меньше</span>
          {RAMP.map((colour) => (
            <span key={colour} className='h-3 w-3 rounded-mark' style={{ backgroundColor: colour }} />
          ))}
          <span>больше</span>
        </div>
        {picked && (
          <span className='text-small tabular-nums text-ink-2'>
            {picked}: {formatDuration(pickedSeconds)}
          </span>
        )}
      </div>
    </div>
  )
}
