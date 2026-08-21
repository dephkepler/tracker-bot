import type { OverviewToday } from '@/lib/api-types'
import { Card } from '@/components/ui/card'
import { HeroFigure } from '@/components/ui/hero-figure'
import { StatTile } from '@/components/ui/stat-tile'
import { formatClock, formatDuration } from '@/lib/format'

// The day, across every activity.
//
// There is deliberately no meter here. The bot used to compare the day against
// a hardcoded 120 minutes, but a target now belongs to an individual activity
// (activities.target_minutes), so there is no day-level limit left to measure
// against — inventing one would be a number the user never set.
export function TodayCard({ today }: { today: OverviewToday }) {
  return (
    <Card>
      <div className='text-micro uppercase tracking-wide text-ink-3'>Сегодня</div>
      <HeroFigure className='mt-1' value={formatClock(today.total_seconds)} unit={formatDuration(today.total_seconds)} />

      <div className='mt-4 grid grid-cols-2 gap-2'>
        <StatTile label='Сессий' value={String(today.sessions)} />
        <StatTile label='Активностей' value={String(today.activities_count)} />
      </div>
    </Card>
  )
}
