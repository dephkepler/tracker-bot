'use client'

import { useState } from 'react'
import { formatDuration } from '@/lib/format'
import type { SeriesBucket } from '@/lib/api-types'

// The shape of a day: twenty-four columns, one per hour.
//
// HTML divs rather than SVG. The bars are rectangles with rounded data-ends
// anchored to a baseline, which CSS gives for free, and staying in the DOM
// means the labels wrap and a screen reader can read the thing.
//
// One hue for every column: height already encodes the magnitude, and stepping
// the colour by size would restate it. Which activities made up an hour goes in
// that column's tooltip rather than into a stack — tens of activities against
// any palette is a legend nobody can read.

const HOURS = Array.from({ length: 24 }, (_, h) => h)

export function ColumnChart({ buckets }: { buckets: SeriesBucket[] }) {
  const [openHour, setOpenHour] = useState<number | null>(null)

  // Buckets arrive as naive local wall clocks ("2026-08-21T09:00"), so the hour
  // is read off the string. Parsing it into a Date would convert it into the
  // viewer's zone, which is not necessarily the zone it was bucketed in.
  const byHour = new Map<number, SeriesBucket>()
  for (const bucket of buckets) {
    const hour = Number(bucket.start.slice(11, 13))
    if (!Number.isNaN(hour)) byHour.set(hour, bucket)
  }

  const max = Math.max(1, ...buckets.map((b) => b.seconds))
  const open = openHour !== null ? byHour.get(openHour) : undefined

  return (
    <div>
      <div className='flex h-32 items-end gap-[2px]'>
        {HOURS.map((hour) => {
          const bucket = byHour.get(hour)
          const seconds = bucket?.seconds ?? 0
          const height = seconds > 0 ? Math.max((seconds / max) * 100, 4) : 0
          const selected = openHour === hour
          return (
            <button
              key={hour}
              type='button'
              // The whole column is the hit target, not just the painted part:
              // a two-minute hour would otherwise be a 4px sliver to tap.
              className='group relative flex h-full flex-1 cursor-pointer flex-col justify-end'
              onClick={() => setOpenHour(selected ? null : hour)}
              aria-label={`${hour}:00 — ${formatDuration(seconds)}`}
              aria-pressed={selected}
            >
              <span
                className='w-full rounded-t-mark transition-[height]'
                style={{
                  height: `${height}%`,
                  backgroundColor: seconds > 0 ? 'var(--accent)' : 'var(--surface-2)',
                  opacity: selected ? 1 : undefined,
                }}
              />
            </button>
          )
        })}
      </div>

      {/* Every third hour, or the labels collide at phone width. */}
      <div className='mt-1 flex gap-[2px]'>
        {HOURS.map((hour) => (
          <div key={hour} className='min-w-0 flex-1 text-center text-micro text-ink-3'>
            {hour % 3 === 0 ? hour : ''}
          </div>
        ))}
      </div>

      {open ? (
        <div className='mt-3 rounded-control bg-surface-2 p-3'>
          <div className='flex items-baseline justify-between gap-2'>
            <span className='text-body text-ink'>{openHour}:00</span>
            <span className='text-body font-medium tabular-nums text-ink'>{formatDuration(open.seconds)}</span>
          </div>
          {open.parts && open.parts.length > 0 && (
            <ul className='mt-2 flex flex-col gap-1'>
              {open.parts.map((part, i) => (
                <li key={`${part.name}-${i}`} className='flex items-baseline justify-between gap-2 text-small'>
                  <span className='min-w-0 truncate text-ink-2'>
                    {part.emoji && <span aria-hidden='true'>{part.emoji} </span>}
                    {part.name}
                  </span>
                  <span className='shrink-0 tabular-nums text-ink-3'>{formatDuration(part.seconds)}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      ) : (
        <p className='mt-3 text-small text-ink-3'>Нажми на столбец, чтобы увидеть, из чего он состоит.</p>
      )}
    </div>
  )
}
