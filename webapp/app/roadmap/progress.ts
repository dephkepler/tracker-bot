import type { RoadmapGoal, RoadmapTechnology } from '@/lib/api-types'

const MONTHS = ['янв', 'фев', 'мар', 'апр', 'май', 'июн', 'июл', 'авг', 'сен', 'окт', 'ноя', 'дек']

/**
 * Cards closed per month, newest last.
 *
 * Aggregated by month rather than by day on purpose: with three goals and a
 * handful of technologies, a daily series over several months is almost all
 * zeros — a sparse line that reads as a broken chart rather than as progress.
 *
 * Built from each card's done_at, which is a real instant, so slicing the ISO
 * string is safe and avoids the timezone question entirely: the month a card
 * was closed in is the month its own timestamp says.
 */
export function closedByMonth(technologies: RoadmapTechnology[], months = 6): Array<{ label: string; value: number }> {
  const counts = new Map<string, number>()
  for (const tech of technologies) {
    for (const card of tech.cards) {
      if (!card.done_at) continue
      const key = card.done_at.slice(0, 7) // YYYY-MM
      counts.set(key, (counts.get(key) ?? 0) + 1)
    }
  }
  if (counts.size === 0) return []

  // A continuous run of months ending at the latest one that has anything, so
  // a gap shows as a gap instead of collapsing.
  const keys = [...counts.keys()].sort()
  const last = keys[keys.length - 1]
  const [lastYear, lastMonth] = last.split('-').map(Number)

  const out: Array<{ label: string; value: number }> = []
  for (let i = months - 1; i >= 0; i--) {
    const d = new Date(Date.UTC(lastYear, lastMonth - 1 - i, 1))
    const key = `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, '0')}`
    out.push({ label: MONTHS[d.getUTCMonth()], value: counts.get(key) ?? 0 })
  }
  return out
}

export function goalPercent(goal: RoadmapGoal): number {
  return goal.total_cards > 0 ? Math.round((goal.done_cards / goal.total_cards) * 100) : 0
}

export function allTechnologies(goals: RoadmapGoal[], orphans: RoadmapTechnology[]): RoadmapTechnology[] {
  return [...goals.flatMap((g) => g.technologies), ...orphans]
}

const KIND_ICON: Record<string, string> = { topic: '📌', article: '📄', book: '📚', lecture: '🎧' }
const DIFFICULTY_ICON: Record<number, string> = { 1: '🟢', 2: '🟡', 3: '🔴' }

/** The same icons the bot uses, so the two surfaces read the same. */
export function kindIcon(kind: string): string {
  return KIND_ICON[kind] ?? KIND_ICON.topic
}

export function difficultyIcon(difficulty: number): string {
  return DIFFICULTY_ICON[difficulty] ?? DIFFICULTY_ICON[2]
}
