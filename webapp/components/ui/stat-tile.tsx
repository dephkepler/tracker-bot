import { cx } from '@/lib/cx'

// A single current value with no plot — deliberately not a one-bar bar chart.
// Values are tabular so a row of tiles lines up and does not jitter on refresh.
export function StatTile({ label, value, hint, className }: { label: string; value: string; hint?: string; className?: string }) {
  return (
    <div className={cx('rounded-card border border-line bg-surface p-3', className)}>
      <div className='text-micro uppercase tracking-wide text-ink-3'>{label}</div>
      <div className='mt-1 text-h1 font-semibold tabular-nums text-ink'>{value}</div>
      {hint && <div className='mt-0.5 text-small text-ink-3'>{hint}</div>}
    </div>
  )
}
