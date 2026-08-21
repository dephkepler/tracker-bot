'use client'

import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { ChallengesResponse, LearningResponse, RoadmapResponse } from '@/lib/api-types'
import { Skeleton } from '@/components/ui/skeleton'

// One live number per remaining domain, each a way into its section.
//
// Three separate requests rather than an aggregate endpoint. They run in
// parallel over one HTTP/2 connection, and an aggregate would be a fourth
// serving of numbers the sections already return — a place for the two to
// disagree. The cost is that each response carries its full payload to show one
// figure, which at one user's scale is kilobytes.
export function DomainTiles() {
  const learning = useQuery({
    queryKey: ['learning'],
    queryFn: () => api<LearningResponse>('/v1/learning?days=30'),
  })
  const challenges = useQuery({
    queryKey: ['challenges'],
    queryFn: () => api<ChallengesResponse>('/v1/challenges'),
  })
  const roadmap = useQuery({
    queryKey: ['roadmap'],
    queryFn: () => api<RoadmapResponse>('/v1/roadmap'),
  })

  const activeChallenge = challenges.data?.challenges[0]
  const roadmapTotals = roadmap.data
    ? roadmap.data.goals.reduce(
        (acc, goal) => ({ done: acc.done + goal.done_cards, total: acc.total + goal.total_cards }),
        { done: 0, total: 0 }
      )
    : null

  return (
    <div className='grid grid-cols-3 gap-2'>
      <DomainTile
        href='/learning'
        label='Слова'
        value={learning.data ? String(learning.data.due_words) : undefined}
        hint='к повтору'
        isPending={learning.isPending}
      />
      <DomainTile
        href='/challenges'
        label='Челленджи'
        value={activeChallenge ? `${activeChallenge.done_days}/${activeChallenge.total_days}` : challenges.data ? '—' : undefined}
        hint={activeChallenge ? `серия ${activeChallenge.current_streak}` : 'нет активных'}
        isPending={challenges.isPending}
      />
      <DomainTile
        href='/roadmap'
        label='Роадмапы'
        value={roadmapTotals ? `${roadmapTotals.done}/${roadmapTotals.total}` : undefined}
        hint='карточек'
        isPending={roadmap.isPending}
      />
    </div>
  )
}

function DomainTile({
  href,
  label,
  value,
  hint,
  isPending,
}: {
  href: string
  label: string
  value?: string
  hint: string
  isPending: boolean
}) {
  return (
    <Link href={href} className='rounded-card border border-line bg-surface p-3'>
      <div className='text-micro uppercase tracking-wide text-ink-3'>{label}</div>
      {isPending ? (
        <Skeleton className='mt-1 h-6' />
      ) : (
        <div className='mt-1 text-h1 font-semibold tabular-nums text-ink'>{value ?? '—'}</div>
      )}
      <div className='mt-0.5 text-micro text-ink-3'>{hint}</div>
    </Link>
  )
}
