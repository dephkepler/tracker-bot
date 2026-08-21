'use client'

import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { LearningResponse } from '@/lib/api-types'
import { ChartFrame } from '@/components/charts/chart-frame'
import { LineChart } from '@/components/charts/line-chart'
import { MiniBars } from '@/components/charts/mini-bars'
import { SectionNav } from '@/components/layout/section-nav'
import { Card } from '@/components/ui/card'
import { Meter } from '@/components/ui/meter'
import { StatTile } from '@/components/ui/stat-tile'
import { formatPercent, formatStreak } from '@/lib/format'

export default function LearningPage() {
  const learning = useQuery({
    queryKey: ['learning'],
    queryFn: () => api<LearningResponse>('/v1/learning?days=30'),
  })

  const data = learning.data

  return (
    <main className='tg-shell mx-auto max-w-[680px] px-4 pt-3'>
      <h1 className='mb-3 text-h1 font-semibold text-ink'>Слова</h1>
      <SectionNav />

      {learning.isError && (
        <Card role='alert'>
          <p className='text-body text-ink-2'>Не удалось загрузить{learning.error.message ? `: ${learning.error.message}` : ''}.</p>
        </Card>
      )}

      <div className='mb-4 grid grid-cols-2 gap-2'>
        <StatTile label='К повтору' value={String(data?.due_words ?? 0)} hint='сегодня' />
        <StatTile label='Серия' value={String(data?.streak_days ?? 0)} hint={data ? formatStreak(data.streak_days) : undefined} />
        <StatTile label='Выучено' value={String(data?.learned_words ?? 0)} hint={data ? `из ${data.total_words}` : undefined} />
        <StatTile label='Точность' value={data ? formatPercent(data.accuracy_percent) : '—'} hint={data ? `${data.reviews_total} повторов` : undefined} />
      </div>

      <div className='flex flex-col gap-4'>
        <ChartFrame
          title='Повторов в день'
          hint='30 дней'
          isPending={learning.isPending}
          error={learning.error}
          isEmpty={(data?.reviews_by_day.length ?? 0) === 0}
        >
          {data && (
            <LineChart
              area
              series={[{ key: 'reviews', label: 'Повторы', color: 'var(--accent)', values: data.reviews_by_day.map((d) => d.total) }]}
              xLabels={data.reviews_by_day.map((d) => d.date.slice(8, 10))}
              formatValue={(v) => String(Math.round(v))}
            />
          )}
        </ChartFrame>

        <ChartFrame
          title='Коллекции'
          isPending={learning.isPending}
          error={learning.error}
          isEmpty={(data?.collections.length ?? 0) === 0}
          emptyText='Коллекций пока нет. Они создаются в боте.'
        >
          {data && (
            <div className='flex flex-col gap-3'>
              {data.collections.map((c) => (
                <div key={c.name}>
                  <div className='flex items-baseline justify-between gap-2 text-body'>
                    <span className='min-w-0 truncate text-ink'>{c.name}</span>
                    <span className='shrink-0 tabular-nums text-ink-2'>
                      {c.learned_words}/{c.total_words}
                      {c.due_words > 0 && <span className='ml-1.5 text-ink-3'>· {c.due_words} к повтору</span>}
                    </span>
                  </div>
                  <Meter className='mt-1' value={c.learned_words} target={Math.max(c.total_words, 1)} />
                </div>
              ))}
            </div>
          )}
        </ChartFrame>

        {data && data.reviews_by_day.some((d) => d.total > 0) && (
          <ChartFrame title='Верных ответов' hint='по дням, из повторов того дня'>
            <MiniBars
              items={data.reviews_by_day
                .filter((d) => d.total > 0)
                .slice(-8)
                .map((d) => ({
                  label: d.date.slice(5).replace('-', '.'),
                  value: Math.round((d.correct / d.total) * 100),
                  hint: `${Math.round((d.correct / d.total) * 100)}%`,
                }))}
            />
          </ChartFrame>
        )}

        {data && (
          <Card>
            <p className='text-body text-ink-2'>
              {data.reminder_active
                ? `Напоминания включены, каждые ${data.reminder_interval_minutes} мин.`
                : 'Напоминания выключены.'}
            </p>
            <p className='mt-1 text-small text-ink-3'>Повторять слова и менять расписание — в боте.</p>
          </Card>
        )}
      </div>

      <div className='h-4' />
    </main>
  )
}
