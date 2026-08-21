'use client'

import { cx } from '@/lib/cx'

// The period control. A tablist rather than a <select>: five short options are
// worth one tap, and a native picker on a phone costs two plus a modal.
export function Segmented<T extends string>({
  options,
  value,
  onChange,
  label,
}: {
  options: Array<{ value: T; label: string }>
  value: T
  onChange: (value: T) => void
  label: string
}) {
  return (
    <div role='tablist' aria-label={label} className='flex gap-1 rounded-control bg-surface-2 p-1'>
      {options.map((option) => {
        const selected = option.value === value
        return (
          <button
            key={option.value}
            role='tab'
            aria-selected={selected}
            type='button'
            onClick={() => onChange(option.value)}
            className={cx(
              'flex-1 rounded-control px-2 py-1.5 text-small transition-colors',
              selected ? 'bg-surface font-medium text-ink shadow-none' : 'text-ink-3'
            )}
          >
            {option.label}
          </button>
        )
      })}
    </div>
  )
}
