import type { ReactNode } from 'react'
import { cx } from '@/lib/cx'

export function SectionHeader({
  title,
  hint,
  as: Tag = 'h2',
  className,
}: {
  title: ReactNode
  hint?: ReactNode
  as?: 'h1' | 'h2'
  className?: string
}) {
  return (
    <div className={cx('mb-3 flex items-baseline justify-between gap-3', className)}>
      <Tag className={cx('font-semibold text-ink', Tag === 'h1' ? 'text-h1' : 'text-h2')}>{title}</Tag>
      {hint && <span className='text-small text-ink-3'>{hint}</span>}
    </div>
  )
}
