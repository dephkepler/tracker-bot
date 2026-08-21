'use client'

import Link from 'next/link'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api, ApiError, StaleLaunchError } from '@/lib/api'
import type { OverviewResponse } from '@/lib/api-types'
import { Card } from '@/components/ui/card'
import { SectionHeader } from '@/components/ui/section-header'
import { Skeleton } from '@/components/ui/skeleton'
import { RankedBars } from '@/components/charts/ranked-bars'
import { formatTimeOfDay } from '@/lib/format'
import { closeApp } from '@/lib/telegram'
import { TodayCard } from './overview/today-card'
import { CurrentCard } from './overview/current-card'

export default function OverviewPage() {
  const queryClient = useQueryClient()
  const overview = useQuery({
    queryKey: ['overview'],
    queryFn: () => api<OverviewResponse>('/v1/track/overview'),
  })

  return (
    <main className='tg-shell mx-auto max-w-[680px] px-4 pt-3'>
      <SectionHeader as='h1' title='Обзор' hint={overview.data ? `обновлено ${formatTimeOfDay(overview.data.meta.generated_at)}` : undefined} />

      {overview.isPending && <LoadingState />}
      {overview.isError && <ErrorState error={overview.error} />}

      {overview.data && (
        <div className='flex flex-col gap-4'>
          <TodayCard today={overview.data.today} />

          {overview.data.current_activity ? (
            <CurrentCard current={overview.data.current_activity} />
          ) : (
            <Card>
              <p className='text-body text-ink-2'>Пока нет ни одной записанной сессии. Начни трекинг в боте — здесь появится сводка.</p>
            </Card>
          )}

          <Card>
            <SectionHeader title='На что ушло время' hint='сегодня' />
            <RankedBars items={overview.data.today.top_activities} total={overview.data.today.total_seconds} />
          </Card>

          {/* The ways into the sections. The overview stays a hub rather than
              growing a tab bar, which would cost 56px of a phone viewport
              permanently. */}
          <Link
            href='/track'
            className='flex items-center justify-between rounded-card border border-line bg-surface p-4 text-body text-ink'
          >
            <span>Время: периоды, часы, календарь</span>
            <span aria-hidden='true' className='text-ink-3'>
              →
            </span>
          </Link>

          <Link
            href='/roadmap'
            className='flex items-center justify-between rounded-card border border-line bg-surface p-4 text-body text-ink'
          >
            <span>Роадмапы: цели, карточки, ИИ</span>
            <span aria-hidden='true' className='text-ink-3'>
              →
            </span>
          </Link>

          <div className='flex items-center justify-center gap-4 pb-2'>
            <button
              type='button'
              onClick={() => queryClient.invalidateQueries({ queryKey: ['overview'] })}
              className='rounded-control px-3 py-2 text-small text-ink-3 hover:bg-surface-2 hover:text-ink-2'
            >
              Обновить
            </button>
            {/* Inside a Mini App there is no address bar, so the only way to
                reach the init-data page is a link. Deliberately understated:
                it is a maintenance tool for keeping tgauth's golden vector
                current, not a feature. */}
            <Link href='/debug/initdata' className='rounded-control px-3 py-2 text-small text-ink-3'>
              initData
            </Link>
          </div>
        </div>
      )}
    </main>
  )
}

// Vertical swipes are disabled so the WebView does not read a scroll as
// swipe-to-close, which also means there is no pull-to-refresh. The button
// above is the only way to refetch, so it is not optional.

function LoadingState() {
  return (
    <div className='flex flex-col gap-4' role='status' aria-live='polite'>
      <Skeleton className='h-32' />
      <Skeleton className='h-28' />
      <Skeleton className='h-40' />
    </div>
  )
}

function ErrorState({ error }: { error: Error }) {
  // A dead launch is not an error the user can retry into — only reopening the
  // app mints new init data, so say that instead of "try again".
  if (error instanceof StaleLaunchError) {
    return (
      <Card role='alert'>
        <p className='text-body text-ink-2'>Этот запуск больше не действителен. Закрой и открой приложение заново.</p>
        <button type='button' onClick={closeApp} className='mt-3 rounded-control bg-accent px-3 py-2 text-body font-medium text-white'>
          Закрыть
        </button>
      </Card>
    )
  }

  if (error instanceof ApiError && error.code === 'user_not_found') {
    return (
      <Card role='alert'>
        <p className='text-body text-ink-2'>Бот тебя ещё не знает. Открой его и нажми Start — после этого здесь появятся данные.</p>
      </Card>
    )
  }

  return (
    <Card role='alert'>
      <p className='text-body text-ink-2'>Не удалось загрузить сводку{error.message ? `: ${error.message}` : ''}.</p>
    </Card>
  )
}
