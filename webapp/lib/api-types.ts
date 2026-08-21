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
