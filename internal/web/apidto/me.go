package apidto

import "time"

// MeResponse is who the dashboard is talking to, and the timezone every other
// response is bucketed by.
//
// No users.id, phone number or email: the client has no use for them, and the
// database id in particular is not something to hand out.
//
// No daily target either. Targets are per-activity since migration 0013
// (activities.target_minutes), so a single global number would be a second,
// disagreeing copy of something that now belongs to each activity.
type MeResponse struct {
	TgUserID int64 `json:"tg_user_id"`
	// Timezone is the IANA name the server bucketed by, which is what makes
	// LocalTime values in other responses interpretable.
	Timezone string `json:"timezone"`
	// UTCOffsetMinutes is that zone's offset right now — a convenience for the
	// client, which should not be parsing zone databases.
	UTCOffsetMinutes int     `json:"utc_offset_minutes"`
	Language         string  `json:"language"`
	Now              Instant `json:"now"`
}

// NewMeResponse builds the response from the resolved request user.
func NewMeResponse(tgUserID int64, loc *time.Location, language string, now time.Time) MeResponse {
	local := now.In(loc)
	_, offsetSeconds := local.Zone()
	return MeResponse{
		TgUserID:         tgUserID,
		Timezone:         loc.String(),
		UTCOffsetMinutes: offsetSeconds / 60,
		Language:         language,
		Now:              FromInstant(local),
	}
}
