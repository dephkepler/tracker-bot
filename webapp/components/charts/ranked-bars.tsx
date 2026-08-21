'use client'

import { useState } from 'react'
import type { ActivityTotal } from '@/lib/api-types'
import { formatDuration, formatPercent } from '@/lib/format'

// Part-to-whole across activities, as ranked horizontal bars.
//
// Not a pie or donut: with more than a handful of categories a ring cannot be
// read, and comparing angles is harder than comparing lengths even with three.
//
// One hue for every bar, not a gradient. Length already encodes magnitude, so
// stepping the colour by rank would be decoration restating the same variable —
// and colour that follows rank repaints the survivors whenever a filter changes
// the set. Identity comes from each activity's own emoji, which means this chart
// needs no categorical palette at all and so has no cap on how many activities
// it can show.
//
// Every bar carries its value and share as a direct label. That is not a style
// choice: the accent sits at 2.82:1 against the light surface, below the 3:1
// floor, and the validator's verdict is that visible labels or a table view are
// then mandatory. Both are here.

const BAR_MIN_PERCENT = 1.5

export function RankedBars({ items, total }: { items: ActivityTotal[]; total: number }) {
  const [asTable, setAsTable] = useState(false)

  if (items.length === 0) {
    return <p className='text-body text-ink-3'>Сегодня пока ничего не записано.</p>
  }

  const max = Math.max(...items.map((i) => i.seconds))

  return (
    <div>
      <div className='mb-2 flex justify-end'>
        <button
          type='button'
          onClick={() => setAsTable((v) => !v)}
          className='rounded-control px-2 py-1 text-small text-ink-3 hover:bg-surface-2 hover:text-ink-2'
          aria-pressed={asTable}
        >
          {asTable ? 'Показать полосами' : 'Показать таблицей'}
        </button>
      </div>

      {asTable ? <BreakdownTable items={items} total={total} /> : <Bars items={items} max={max} />}
    </div>
  )
}

function Bars({ items, max }: { items: ActivityTotal[]; max: number }) {
  return (
    // 2px between bars: adjacent fills need a surface gap or two rows read as
    // one shape.
    <ul className='flex flex-col gap-2'>
      {items.map((item) => {
        const width = max > 0 ? Math.max((item.seconds / max) * 100, BAR_MIN_PERCENT) : 0
        return (
          <li key={item.activity_id}>
            <div className='flex items-baseline justify-between gap-2 text-body'>
              <span className='min-w-0 truncate text-ink'>
                {item.emoji && <span aria-hidden='true'>{item.emoji} </span>}
                {item.name}
              </span>
              {/* Values wear text tokens, never the mark's colour. */}
              <span className='shrink-0 tabular-nums text-ink-2'>
                {formatDuration(item.seconds)}
                <span className='ml-1.5 text-ink-3'>{formatPercent(item.share_percent)}</span>
              </span>
            </div>
            <div className='mt-1 h-2 rounded-full bg-surface-2'>
              {/* Anchored to the left baseline with a 4px rounded data-end. */}
              <div className='h-full rounded-mark bg-accent' style={{ width: `${width}%` }} />
            </div>
          </li>
        )
      })}
    </ul>
  )
}

function BreakdownTable({ items, total }: { items: ActivityTotal[]; total: number }) {
  return (
    <div className='-mx-1 overflow-x-auto'>
      <table className='w-full text-body'>
        <thead>
          <tr className='text-left text-micro uppercase tracking-wide text-ink-3'>
            <th scope='col' className='px-1 pb-1 font-medium'>
              Активность
            </th>
            <th scope='col' className='px-1 pb-1 text-right font-medium'>
              Время
            </th>
            <th scope='col' className='px-1 pb-1 text-right font-medium'>
              Сессии
            </th>
            <th scope='col' className='px-1 pb-1 text-right font-medium'>
              Доля
            </th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.activity_id} className='border-t border-line'>
              <td className='px-1 py-1.5 text-ink'>
                {item.emoji && <span aria-hidden='true'>{item.emoji} </span>}
                {item.name}
              </td>
              <td className='px-1 py-1.5 text-right tabular-nums text-ink-2'>{formatDuration(item.seconds)}</td>
              <td className='px-1 py-1.5 text-right tabular-nums text-ink-2'>{item.sessions}</td>
              <td className='px-1 py-1.5 text-right tabular-nums text-ink-2'>{formatPercent(item.share_percent)}</td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr className='border-t border-line'>
            <td className='px-1 pt-1.5 text-ink-3'>Всего</td>
            <td className='px-1 pt-1.5 text-right tabular-nums text-ink'>{formatDuration(total)}</td>
            <td />
            <td />
          </tr>
        </tfoot>
      </table>
    </div>
  )
}
