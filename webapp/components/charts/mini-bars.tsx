// A short labelled bar list, in the order given rather than ranked. For a
// handful of buckets this is the whole chart: length carries the magnitude, the
// label carries the identity, and there is nothing to hover for.
export function MiniBars({ items }: { items: Array<{ label: string; value: number; hint?: string }> }) {
  if (items.length === 0) return null
  const max = Math.max(1, ...items.map((i) => i.value))

  return (
    <ul className='flex flex-col gap-1.5'>
      {items.map((item) => (
        <li key={item.label} className='flex items-center gap-2'>
          <span className='w-14 shrink-0 text-small tabular-nums text-ink-3'>{item.label}</span>
          <span className='h-2 flex-1 rounded-full bg-surface-2'>
            <span
              className='block h-full rounded-mark bg-accent'
              style={{ width: `${Math.max((item.value / max) * 100, item.value > 0 ? 3 : 0)}%` }}
            />
          </span>
          <span className='w-8 shrink-0 text-right text-small tabular-nums text-ink-2'>{item.hint ?? item.value}</span>
        </li>
      ))}
    </ul>
  )
}
