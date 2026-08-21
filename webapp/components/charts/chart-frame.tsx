import type { ReactNode } from 'react'
import { Skeleton } from '@/components/ui/skeleton'

// One place for the four states every chart has, so five charts do not grow
// five spellings of "нет данных".
export function ChartFrame({
  title,
  hint,
  isPending,
  error,
  isEmpty,
  emptyText = 'Нет данных за период.',
  children,
}: {
  title: string
  hint?: string
  isPending?: boolean
  error?: Error | null
  isEmpty?: boolean
  emptyText?: string
  children: ReactNode
}) {
  return (
    <section className='rounded-card border border-line bg-surface p-4'>
      <div className='mb-3 flex items-baseline justify-between gap-3'>
        <h2 className='text-h2 font-semibold text-ink'>{title}</h2>
        {hint && <span className='text-small text-ink-3'>{hint}</span>}
      </div>

      {isPending ? (
        <Skeleton className='h-40' />
      ) : error ? (
        <p role='alert' className='text-body text-ink-2'>
          Не удалось загрузить{error.message ? `: ${error.message}` : ''}.
        </p>
      ) : isEmpty ? (
        // A fresh account has no sessions at all, and every chart has to say so
        // rather than render a collapsed axis that looks broken.
        <p className='text-body text-ink-3'>{emptyText}</p>
      ) : (
        children
      )}
    </section>
  )
}
