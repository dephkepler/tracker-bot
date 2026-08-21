import { cx } from '@/lib/cx'
import { formatPercent } from '@/lib/format'

// Progress toward one limit. A meter, not a chart: a single ratio against a
// target has no second dimension to plot, and a one-bar bar chart would imply
// there were more bars coming.
export function Meter({
  value,
  target,
  label,
  className,
}: {
  value: number
  target: number
  /** Shown at the right. Omitted when there is no target to compare against. */
  label?: string
  className?: string
}) {
  const ratio = target > 0 ? Math.min(value / target, 1) : 0
  const over = target > 0 && value >= target

  return (
    <div className={className}>
      <div className='h-2 overflow-hidden rounded-full bg-surface-2' role='presentation'>
        <div
          className={cx('h-full rounded-full transition-[width] duration-300', over ? 'bg-good' : 'bg-accent')}
          style={{ width: `${Math.max(ratio * 100, value > 0 ? 2 : 0)}%` }}
        />
      </div>
      {label && (
        <div className='mt-1.5 flex justify-between text-small text-ink-3'>
          <span>{label}</span>
          <span className='tabular-nums'>{formatPercent(Math.round(ratio * 100))}</span>
        </div>
      )}
    </div>
  )
}
