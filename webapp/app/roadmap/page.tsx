'use client'

import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { RoadmapResponse } from '@/lib/api-types'
import { ChartFrame } from '@/components/charts/chart-frame'
import { MiniBars } from '@/components/charts/mini-bars'
import { Card } from '@/components/ui/card'
import { Meter } from '@/components/ui/meter'
import { Skeleton } from '@/components/ui/skeleton'
import { allTechnologies, closedByMonth, goalPercent } from './progress'
import { TechnologyCard } from './technology-card'

export default function RoadmapPage() {
  const roadmap = useQuery({
    queryKey: ['roadmap'],
    queryFn: () => api<RoadmapResponse>('/v1/roadmap'),
  })

  const technologies = roadmap.data ? allTechnologies(roadmap.data.goals, roadmap.data.orphan_technologies) : []
  const closed = closedByMonth(technologies)

  return (
    <main className='tg-shell mx-auto max-w-[680px] px-4 pt-3'>
      <h1 className='mb-3 text-h1 font-semibold text-ink'>Роадмапы</h1>

      {roadmap.isPending && (
        <div className='flex flex-col gap-4' role='status' aria-live='polite'>
          <Skeleton className='h-24' />
          <Skeleton className='h-56' />
        </div>
      )}

      {roadmap.isError && (
        <Card role='alert'>
          <p className='text-body text-ink-2'>
            Не удалось загрузить план{roadmap.error.message ? `: ${roadmap.error.message}` : ''}.
          </p>
        </Card>
      )}

      {roadmap.data && (
        <div className='flex flex-col gap-4'>
          {roadmap.data.goals.length === 0 && roadmap.data.orphan_technologies.length === 0 && (
            <Card>
              <p className='text-body text-ink-2'>Планов пока нет. Цели и технологии создаются в боте.</p>
            </Card>
          )}

          {closed.length > 0 && (
            <ChartFrame title='Закрыто карточек' hint='по месяцам'>
              <MiniBars items={closed} />
            </ChartFrame>
          )}

          {roadmap.data.goals.map((goal) => (
            <section key={goal.id} className='flex flex-col gap-3'>
              <Card>
                <div className='flex items-baseline justify-between gap-3'>
                  <h2 className='min-w-0 truncate text-h1 font-semibold text-ink'>🎯 {goal.name}</h2>
                  <span className='shrink-0 text-body tabular-nums text-ink-2'>{goalPercent(goal)}%</span>
                </div>
                <Meter
                  className='mt-3'
                  value={goal.done_cards}
                  target={Math.max(goal.total_cards, 1)}
                  label={`${goal.done_cards} из ${goal.total_cards} карточек`}
                />
              </Card>

              {goal.technologies.map((tech) => (
                <TechnologyCard key={tech.id} tech={tech} />
              ))}
            </section>
          ))}

          {roadmap.data.orphan_technologies.length > 0 && (
            <section className='flex flex-col gap-3'>
              <h2 className='text-h2 font-semibold text-ink'>Без цели</h2>
              {roadmap.data.orphan_technologies.map((tech) => (
                <TechnologyCard key={tech.id} tech={tech} />
              ))}
            </section>
          )}
        </div>
      )}

      <div className='h-4' />
    </main>
  )
}
