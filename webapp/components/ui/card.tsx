import type { ReactNode } from 'react'
import { cx } from '@/lib/cx'

// A hairline ring on a surface one step off the plane. No shadow anywhere in
// this app: on a dark surface an elevation shadow is either invisible or reads
// as a smudge, and one pixel of line already does the separating.
export function Card({
  className,
  children,
  role,
}: {
  className?: string
  children: ReactNode
  /** Set to "alert" or "status" when the card *is* the message, not a container. */
  role?: 'alert' | 'status'
}) {
  return (
    <div role={role} className={cx('rounded-card border border-line bg-surface p-4', className)}>
      {children}
    </div>
  )
}
