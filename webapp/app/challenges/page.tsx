'use client'

import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Challenge, ChallengeDay, ChallengeDayStateResponse, ChallengesResponse } from '@/lib/api-types'
import { SectionNav } from '@/components/layout/section-nav'
import { Card } from '@/components/ui/card'
import { Meter } from '@/components/ui/meter'
import { Skeleton } from '@/components/ui/skeleton'
import { StatTile } from '@/components/ui/stat-tile'
import { formatStreak } from '@/lib/format'
import { DayGrid, DayGridLegend } from './day-grid'

export default function ChallengesPage() {
  const challenges = useQuery({
    queryKey: ['challenges'],
    queryFn: () => api<ChallengesResponse>('/v1/challenges'),
  })

  return (
    <main className='tg-shell mx-auto max-w-[680px] px-4 pt-3'>
      <h1 className='mb-3 text-h1 font-semibold text-ink'>Челленджи</h1>
      <SectionNav />

      {challenges.isPending && <Skeleton className='h-40' />}

      {challenges.isError && (
        <Card role='alert'>
          <p className='text-body text-ink-2'>Не удалось загрузить{challenges.error.message ? `: ${challenges.error.message}` : ''}.</p>
        </Card>
      )}

      {challenges.data && challenges.data.challenges.length === 0 && (
        <Card>
          <p className='text-body text-ink-2'>Активных челленджей нет. Создаются в боте.</p>
        </Card>
      )}

      {challenges.data && (
        <div className='flex flex-col gap-4'>
          {challenges.data.challenges.map((challenge) => (
            <ChallengeCard key={challenge.id} challenge={challenge} today={todayFrom(challenges.data.meta.generated_at)} />
          ))}
        </div>
      )}

      <div className='h-4' />
    </main>
  )
}

function ChallengeCard({ challenge, today }: { challenge: Challenge; today: string }) {
  const queryClient = useQueryClient()
  const [busyDate, setBusyDate] = useState<string | null>(null)

  const mark = useMutation({
    mutationFn: ({ date, done }: { date: string; done: boolean }) =>
      api<ChallengeDayStateResponse>(`/v1/challenges/${challenge.id}/days/${date}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ done }),
      }),
    onSettled: async () => {
      setBusyDate(null)
      await queryClient.invalidateQueries({ queryKey: ['challenges'] })
    },
  })

  // Tapping cycles the two states the schema keeps. A day cannot go back to
  // "not marked": forgetting a day and deciding to skip it are different
  // things, and only the first is what "pending" means.
  function pick(day: ChallengeDay) {
    setBusyDate(day.date)
    mark.mutate({ date: day.date, done: day.status !== 'done' })
  }

  return (
    <Card>
      <div className='flex items-baseline justify-between gap-3'>
        <h2 className='min-w-0 truncate text-h2 font-semibold text-ink'>{challenge.name}</h2>
        <span className='shrink-0 text-small tabular-nums text-ink-3'>
          {challenge.start_date} — {challenge.end_date}
        </span>
      </div>

      <Meter
        className='mt-3'
        value={challenge.done_days}
        target={Math.max(challenge.total_days, 1)}
        label={`${challenge.done_days} из ${challenge.total_days} дней`}
      />

      <div className='mt-3 grid grid-cols-3 gap-2'>
        <StatTile label='Серия' value={String(challenge.current_streak)} hint={formatStreak(challenge.current_streak)} />
        <StatTile label='Лучшая' value={String(challenge.best_streak)} />
        <StatTile label='Пропущено' value={String(challenge.skipped_days)} />
      </div>

      <div className='mt-4'>
        <DayGrid days={challenge.days} today={today} onPick={pick} busyDate={busyDate} />
        <DayGridLegend />
      </div>

      {mark.isError && <p className='mt-2 text-small text-bad'>Не удалось отметить день.</p>}

      <p className='mt-3 text-small text-ink-3'>Нажми на день, чтобы переключить между выполненным и пропущенным.</p>
    </Card>
  )
}

/** Today as a bare date, taken from the server's own clock in the user's zone. */
function todayFrom(generatedAt: string): string {
  return generatedAt.slice(0, 10)
}
