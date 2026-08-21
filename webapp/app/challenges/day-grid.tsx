'use client'

import { cx } from '@/lib/cx'
import type { ChallengeDay, ChallengeDayStatus } from '@/lib/api-types'

// A challenge as a row of squares, one per day.
//
// Status is state, not magnitude, so it wears the status palette rather than the
// sequential ramp — and each square carries its date and status in a label, so
// the meaning never rests on colour alone.
const STATUS_STYLE: Record<ChallengeDayStatus, string> = {
  done: 'bg-good',
  skipped: 'bg-bad',
  pending: 'bg-surface-2',
}

const STATUS_LABEL: Record<ChallengeDayStatus, string> = {
  done: 'выполнен',
  skipped: 'пропущен',
  pending: 'не отмечен',
}

export function DayGrid({
  days,
  today,
  onPick,
  busyDate,
}: {
  days: ChallengeDay[]
  /** Today as a bare date, for outlining the current square. */
  today: string
  onPick: (day: ChallengeDay) => void
  busyDate: string | null
}) {
  return (
    <div className='flex flex-wrap gap-1'>
      {days.map((day) => (
        <button
          key={day.date}
          type='button'
          onClick={() => onPick(day)}
          disabled={busyDate === day.date}
          title={`${day.date} — ${STATUS_LABEL[day.status]}`}
          aria-label={`${day.date} — ${STATUS_LABEL[day.status]}`}
          className={cx(
            'h-5 w-5 rounded-mark transition-opacity',
            STATUS_STYLE[day.status],
            busyDate === day.date && 'opacity-50',
            day.date === today && 'ring-2 ring-ink ring-offset-1 ring-offset-surface'
          )}
        />
      ))}
    </div>
  )
}

export function DayGridLegend() {
  return (
    <div className='mt-2 flex flex-wrap items-center gap-3 text-micro text-ink-3'>
      {(['done', 'skipped', 'pending'] as ChallengeDayStatus[]).map((status) => (
        <span key={status} className='flex items-center gap-1.5'>
          <span className={cx('h-3 w-3 rounded-mark', STATUS_STYLE[status])} />
          {STATUS_LABEL[status]}
        </span>
      ))}
    </div>
  )
}
