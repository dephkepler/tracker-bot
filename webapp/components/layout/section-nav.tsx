'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cx } from '@/lib/cx'

// Lateral movement between sections.
//
// Two axes, deliberately: Telegram's own back arrow goes up to the overview,
// and this row goes sideways. That keeps the vertical gesture where a Mini App
// user already expects it while making section-to-section one tap instead of
// three.
//
// Chips rather than a bottom tab bar. A tab bar would cost 56px plus the safe
// area on every screen forever, on a viewport that is already about 620px; this
// row is 36px and only appears where it is useful.
const SECTIONS = [
  { href: '/', label: 'Обзор' },
  { href: '/track', label: 'Время' },
  { href: '/learning', label: 'Слова' },
  { href: '/challenges', label: 'Челленджи' },
  { href: '/roadmap', label: 'Роадмапы' },
]

export function SectionNav() {
  const pathname = usePathname()
  // Trailing slashes: the export is built with trailingSlash, so the path can
  // arrive either way depending on how the page was reached.
  const current = pathname.replace(/\/+$/, '') || '/'

  return (
    <nav
      aria-label='Разделы'
      className='-mx-4 mb-3 overflow-x-auto px-4'
      // The row scrolls sideways on a narrow phone rather than wrapping into
      // two lines and pushing the content down.
      style={{ scrollbarWidth: 'none' }}
    >
      <div className='flex w-max gap-1.5'>
        {SECTIONS.map((section) => {
          const active = current === section.href
          return (
            <Link
              key={section.href}
              href={section.href}
              aria-current={active ? 'page' : undefined}
              className={cx(
                'rounded-control px-3 py-1.5 text-small whitespace-nowrap transition-colors',
                active ? 'bg-accent-soft font-medium text-ink' : 'bg-surface-2 text-ink-3'
              )}
            >
              {section.label}
            </Link>
          )
        })}
      </div>
    </nav>
  )
}
