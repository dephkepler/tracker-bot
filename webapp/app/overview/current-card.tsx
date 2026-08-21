import type { CurrentActivity } from '@/lib/api-types'
import { Card } from '@/components/ui/card'
import { Meter } from '@/components/ui/meter'
import { formatDuration, formatStreak } from '@/lib/format'

// The activity the user touched most recently.
//
// Every figure in here belongs to that one activity, which is why they live in
// their own card rather than beside the day totals: its time today is not the
// day's total and its streak is not an overall streak. The meter compares the
// activity against its own target, so both sides of the ratio describe the same
// thing.
export function CurrentCard({ current }: { current: CurrentActivity }) {
  const targetSeconds = current.target_minutes != null ? current.target_minutes * 60 : 0

  return (
    <Card>
      <div className='flex items-baseline justify-between gap-3'>
        <div className='min-w-0'>
          <div className='text-micro uppercase tracking-wide text-ink-3'>Последняя активность</div>
          <div className='mt-1 truncate text-h1 font-semibold text-ink'>
            {current.emoji && <span aria-hidden='true'>{current.emoji} </span>}
            {current.name}
          </div>
        </div>
        <div className='shrink-0 text-right'>
          <div className='text-h1 font-semibold tabular-nums text-ink'>{formatDuration(current.today_seconds)}</div>
          <div className='text-small text-ink-3'>сегодня</div>
        </div>
      </div>

      {targetSeconds > 0 ? (
        <Meter
          className='mt-4'
          value={current.today_seconds}
          target={targetSeconds}
          label={`Цель ${formatDuration(targetSeconds)}`}
        />
      ) : (
        <p className='mt-4 text-small text-ink-3'>Дневная цель для этой активности не задана — её ставят в боте.</p>
      )}

      <div className='mt-3 flex items-baseline gap-1.5 text-body'>
        <span className='text-ink-3'>Серия</span>
        <span className='font-medium tabular-nums text-ink'>{formatStreak(current.streak_days)}</span>
        <span className='text-small text-ink-3'>подряд по этой активности</span>
      </div>
    </Card>
  )
}
