// Mirrors internal/web/apidto one-to-one. Handwritten rather than generated,
// which is a real cost: change a Go struct and TypeScript will not notice. The
// discipline that keeps them honest is that both sides move in the same commit.
//
// Two wire conventions worth remembering when adding to this file:
//   - durations arrive as whole seconds and their field names say so;
//   - a calendar day arrives as "YYYY-MM-DD" and must never be fed to
//     new Date(), which reads it as UTC midnight and renders as the day before
//     in any negative-offset zone.

export interface Meta {
  timezone: string
  generated_at: string
  /** The window the server actually used, inclusive, as bare dates. */
  from?: string
  to?: string
  /** Resolved, never "auto". */
  granularity?: string
  /** The activity set actually queried — the expansion of an omitted filter. */
  activity_ids?: number[]
}

export interface Activity {
  id: number
  name: string
  emoji: string
  selected: boolean
  archived: boolean
  target_minutes: number | null
}

export interface ActivitiesResponse {
  activities: Activity[]
  meta: Meta
}

export interface ActivityTotal {
  activity_id: number
  name: string
  emoji: string
  seconds: number
  sessions: number
  share_percent: number
}

/** Every number here belongs to one activity, not to the day. */
export interface CurrentActivity {
  id: number
  name: string
  emoji: string
  today_seconds: number
  streak_days: number
  target_minutes: number | null
}

export interface OverviewToday {
  total_seconds: number
  sessions: number
  activities_count: number
  top_activities: ActivityTotal[]
}

export interface OverviewResponse {
  today: OverviewToday
  /** null on an account that has never tracked anything. */
  current_activity: CurrentActivity | null
  meta: Meta
}

export interface MeResponse {
  tg_user_id: number
  timezone: string
  utc_offset_minutes: number
  language: string
  now: string
}

export interface MonthTotal {
  month: string
  seconds: number
}

export interface BreakdownResponse {
  total_seconds: number
  total_sessions: number
  activities: ActivityTotal[]
  monthly: MonthTotal[]
  meta: Meta
}

export interface SeriesPart {
  name: string
  emoji: string
  seconds: number
}

/**
 * One point of a series. `start` is a bare "YYYY-MM-DD" for day and month
 * granularity and a naive local "YYYY-MM-DDTHH:mm" for hour — read it as a
 * string, never through new Date(), which would reinterpret it in the viewer's
 * zone rather than the one the server bucketed by.
 */
export interface SeriesBucket {
  start: string
  seconds: number
  parts?: SeriesPart[]
}

export interface SeriesResponse {
  by: 'total' | 'activity'
  buckets: SeriesBucket[]
  meta: Meta
}

export interface HeatDay {
  date: string
  seconds: number
}

export interface HeatmapResponse {
  days: HeatDay[]
  /** The busiest day in the window, for scaling the colour ramp. */
  max_seconds: number
  meta: Meta
}

export interface DayResponse {
  date: string
  total_seconds: number
  total_sessions: number
  activities: ActivityTotal[]
  hours: SeriesBucket[]
  meta: Meta
}
