import { cx } from '@/lib/cx'

// The one number a screen is about. Proportional figures, not tabular: this is
// display type, and tabular spacing at 34px reads as gappy. Tabular is for
// columns of numbers that have to line up.
export function HeroFigure({ value, unit, className }: { value: string; unit?: string; className?: string }) {
  return (
    <div className={cx('flex items-baseline gap-1.5', className)}>
      <span className='text-hero font-semibold tracking-tight text-ink'>{value}</span>
      {unit && <span className='text-body text-ink-3'>{unit}</span>}
    </div>
  )
}
