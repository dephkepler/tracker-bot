'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { BreakdownResponse, DayResponse, HeatmapResponse, SeriesResponse } from '@/lib/api-types'
import { ChartFrame } from '@/components/charts/chart-frame'
import { ColumnChart } from '@/components/charts/column-chart'
import { DayHeatmap } from '@/components/charts/day-heatmap'
import { LineChart } from '@/components/charts/line-chart'
import { RankedBars } from '@/components/charts/ranked-bars'
import { Segmented } from '@/components/ui/segmented'
import { StatTile } from '@/components/ui/stat-tile'
import { formatClock, formatDuration } from '@/lib/format'
import { DEFAULT_PERIOD, PERIODS, periodLabel, seriesGranularity, type PeriodKey } from './period'

export default function TrackPage() {
  const [period, setPeriod] = useState<PeriodKey>(DEFAULT_PERIOD)
  const [pickedDay, setPickedDay] = useState<string | null>(null)

  const breakdown = useQuery({
    queryKey: ['breakdown', period],
    queryFn: () => api<BreakdownResponse>(`/v1/track/breakdown?period=${period}`),
    // Without this the whole section blanks on every period change; the stale
    // numbers stay put for the moment the next ones take.
    placeholderData: (previous) => previous,
  })

  const series = useQuery({
    queryKey: ['series', period],
    queryFn: () => api<SeriesResponse>(`/v1/track/series?period=${period}&granularity=${seriesGranularity(period)}${period === 'today' ? '&by=activity' : ''}`),
    placeholderData: (previous) => previous,
  })

  // The heatmap deliberately ignores the period control: its job is the long
  // view, and re-cropping it to seven days would defeat the point.
  const heatmap = useQuery({
    queryKey: ['heatmap'],
    queryFn: () => api<HeatmapResponse>('/v1/track/heatmap?weeks=26'),
  })

  const day = useQuery({
    queryKey: ['day', pickedDay],
    queryFn: () => api<DayResponse>(`/v1/track/day?date=${pickedDay}`),
    enabled: pickedDay !== null,
  })

  const total = breakdown.data?.total_seconds ?? 0
  const windowDays = countDays(breakdown.data?.meta.from, breakdown.data?.meta.to)

  return (
    <main className='tg-shell mx-auto max-w-[680px] px-4 pt-3'>
      <h1 className='mb-3 text-h1 font-semibold text-ink'>Время</h1>

      <div className='sticky top-[var(--tg-safe-top,0px)] z-20 -mx-4 mb-4 bg-plane px-4 pb-2 pt-1'>
        <Segmented options={PERIODS} value={period} onChange={setPeriod} label='Период' />
      </div>

      <div className='mb-4 grid grid-cols-2 gap-2'>
        <StatTile label='Всего' value={formatClock(total)} hint={formatDuration(total)} />
        <StatTile label='Сессий' value={String(breakdown.data?.total_sessions ?? 0)} />
        <StatTile label='Активностей' value={String(breakdown.data?.activities.length ?? 0)} />
        <StatTile
          label='В среднем в день'
          value={windowDays > 0 ? formatClock(Math.round(total / windowDays)) : '—'}
          hint={windowDays > 0 ? `за ${windowDays} дн.` : undefined}
        />
      </div>

      <div className='flex flex-col gap-4'>
        <ChartFrame
          title={period === 'today' ? 'По часам' : 'Динамика'}
          hint={periodLabel(period)}
          isPending={series.isPending}
          error={series.error}
          isEmpty={(series.data?.buckets.length ?? 0) === 0}
        >
          {series.data &&
            (period === 'today' ? (
              <ColumnChart buckets={series.data.buckets} />
            ) : (
              <LineChart
                area
                series={[{ key: 'total', label: 'Время', color: 'var(--accent)', values: series.data.buckets.map((b) => b.seconds / 60) }]}
                xLabels={series.data.buckets.map((b) => b.start.slice(8, 10))}
                formatValue={(minutes) => `${Math.round(minutes)}м`}
              />
            ))}
        </ChartFrame>

        <ChartFrame
          title='На что ушло время'
          hint={periodLabel(period)}
          isPending={breakdown.isPending}
          error={breakdown.error}
          isEmpty={(breakdown.data?.activities.length ?? 0) === 0}
        >
          {breakdown.data && <RankedBars items={breakdown.data.activities} total={breakdown.data.total_seconds} />}
        </ChartFrame>

        <ChartFrame
          title='Календарь'
          hint='26 недель'
          isPending={heatmap.isPending}
          error={heatmap.error}
          isEmpty={(heatmap.data?.days.length ?? 0) === 0}
        >
          {heatmap.data?.meta.from && heatmap.data.meta.to && (
            <DayHeatmap
              days={heatmap.data.days}
              maxSeconds={heatmap.data.max_seconds}
              from={heatmap.data.meta.from}
              to={heatmap.data.meta.to}
              onPickDay={setPickedDay}
            />
          )}
        </ChartFrame>

        {pickedDay && (
          <ChartFrame
            title={pickedDay}
            hint='выбранный день'
            isPending={day.isPending}
            error={day.error}
            isEmpty={(day.data?.total_seconds ?? 0) === 0}
            emptyText='В этот день ничего не записано.'
          >
            {day.data && (
              <div className='flex flex-col gap-4'>
                <RankedBars items={day.data.activities} total={day.data.total_seconds} />
                {day.data.hours.length > 0 && <ColumnChart buckets={day.data.hours} />}
              </div>
            )}
          </ChartFrame>
        )}
      </div>

      <div className='h-4' />
    </main>
  )
}

/** Inclusive day count of a bare-date window, as pure arithmetic on UTC. */
function countDays(from?: string, to?: string): number {
  if (!from || !to) return 0
  const start = Date.parse(`${from}T00:00:00Z`)
  const end = Date.parse(`${to}T00:00:00Z`)
  if (Number.isNaN(start) || Number.isNaN(end)) return 0
  return Math.round((end - start) / 86400000) + 1
}
