package challenge

import (
	"strings"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
)

func ListTitle(lang i18n.Lang, count int) string {
	if count == 0 {
		return i18n.T(lang, i18n.KeyChallengeListEmpty)
	}
	return i18n.T(lang, i18n.KeyChallengeListTitleFmt, count)
}

func GridTitle(lang i18n.Lang, item models.ChallengeItem) string {
	pct := 0
	if item.TotalDays > 0 {
		pct = item.DoneDays * 100 / item.TotalDays
	}
	pending := item.TotalDays - item.DoneDays - item.SkippedDays
	return i18n.T(lang, i18n.KeyChallengeGridTitleFmt,
		item.Name, item.StartDate.Format("2006-01-02"), item.EndDate.Format("2006-01-02"), item.TotalDays,
		item.DoneDays, item.SkippedDays, pending, pct,
	)
}

// DayConfirmTitle renders the day-square tap screen: header + status, then
// the progress donut, trend strip, and streak — all above the existing
// Done/Skip buttons (see DayConfirmInlineMenu), which this text doesn't
// replace.
func DayConfirmTitle(lang i18n.Lang, item models.ChallengeItem, detail models.ChallengeDayDetail) string {
	statusKey := i18n.KeyChallengeDayStatusUnmarked
	switch detail.Status {
	case models.ChallengeDayDone:
		statusKey = i18n.KeyChallengeDayStatusDoneText
	case models.ChallengeDaySkipped:
		statusKey = i18n.KeyChallengeDayStatusSkippedText
	}

	pending := item.TotalDays - item.DoneDays - item.SkippedDays
	donePct, skipPct, pendingPct := 0, 0, 0
	if item.TotalDays > 0 {
		// allocateSegments (largest-remainder) rather than three independent
		// truncating divisions, so the printed percentages always sum to
		// exactly 100 instead of drifting under it (e.g. 50/7/42 = 99).
		pcts := allocateSegments(100, item.TotalDays, item.DoneDays, item.SkippedDays, pending)
		donePct, skipPct, pendingPct = pcts[0], pcts[1], pcts[2]
	}

	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyChallengeDayHeaderFmt, item.Name, detail.Day.Format("Mon, 2 Jan 2006"), i18n.T(lang, statusKey)))
	b.WriteString("\n\n")
	b.WriteString(segmentStrip(item.DoneDays, item.SkippedDays, pending))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyChallengeDayProportionsFmt, donePct, skipPct, pendingPct))
	b.WriteString("\n\n")
	b.WriteString(i18n.T(lang, i18n.KeyChallengeDayTrendLabelFmt, len(detail.Trend)))
	b.WriteString("  ")
	b.WriteString(trendStrip(detail.Trend))
	b.WriteString("\n\n")
	b.WriteString(i18n.T(lang, i18n.KeyChallengeDayStreakFmt, detail.CurrentStreak, detail.BestStreak))
	b.WriteString("\n\n")
	b.WriteString(i18n.T(lang, i18n.KeyChallengeDayMarkPrompt))
	return b.String()
}

// segmentWidth is the fixed number of emoji cells the proportions donut is
// allocated across, regardless of the challenge's actual day count.
const segmentWidth = 14

// segmentStrip renders done/skipped/pending as a fixed-width row of colored
// emoji, e.g. "🟢🟢🟢🟢🟢🟢🟢🟡🟡🔴🔴⚪⚪⚪". Uses the largest-remainder
// method so the three counts always sum to exactly segmentWidth cells
// regardless of rounding — simple truncation could under- or over-allocate
// by a cell when percentages don't divide evenly.
func segmentStrip(done, skipped, pending int) string {
	total := done + skipped + pending
	if total <= 0 {
		return strings.Repeat("⚪", segmentWidth)
	}

	counts := allocateSegments(segmentWidth, total, done, skipped, pending)
	var b strings.Builder
	b.WriteString(strings.Repeat("🟢", counts[0]))
	b.WriteString(strings.Repeat("🟡", counts[1]))
	b.WriteString(strings.Repeat("🔴", counts[2]))
	return b.String()
}

// allocateSegments distributes width cells across values proportionally to
// their share of total, using the largest-remainder method: each value gets
// its floor(share) first, then the leftover cells go one each to the values
// with the largest fractional remainder, until the allocation sums to
// exactly width. Guarantees sum(result) == width whenever total > 0.
func allocateSegments(width, total int, values ...int) []int {
	out := make([]int, len(values))
	remainders := make([]float64, len(values))
	assigned := 0
	for i, v := range values {
		exact := float64(v) * float64(width) / float64(total)
		out[i] = int(exact)
		remainders[i] = exact - float64(out[i])
		assigned += out[i]
	}
	for assigned < width {
		best := -1
		for i, r := range remainders {
			if best == -1 || r > remainders[best] {
				best = i
			}
		}
		if best == -1 {
			break
		}
		out[best]++
		remainders[best] = -1 // spent, don't pick it again this round
		assigned++
	}
	return out
}

// trendCharDone/Pending/Skipped are the 3 fixed levels a day's status maps
// to — categorical, not a true intensity sparkline (Challenge days are
// done/skipped/pending, not a continuous quantity), but still reads as a
// shape: a run of trendCharDone then a dip to trendCharSkipped shows up at
// a glance.
const (
	trendCharDone    = "█"
	trendCharPending = "▄"
	trendCharSkipped = "▁"
)

func trendStrip(trend []models.ChallengeDayStatus) string {
	var b strings.Builder
	for _, s := range trend {
		switch s {
		case models.ChallengeDayDone:
			b.WriteString(trendCharDone)
		case models.ChallengeDaySkipped:
			b.WriteString(trendCharSkipped)
		default:
			b.WriteString(trendCharPending)
		}
	}
	return b.String()
}

func ArchiveTitle(lang i18n.Lang, count int) string {
	return i18n.T(lang, i18n.KeyChallengeArchiveTitleFmt, count)
}

func PushText(lang i18n.Lang, challengeName string, dayNum, totalDays int) string {
	return i18n.T(lang, i18n.KeyChallengePushTextFmt, challengeName, dayNum, totalDays)
}

func CreatePromptNameText(lang i18n.Lang) string {
	return i18n.T(lang, i18n.KeyChallengeCreateNamePrompt)
}

func CreatePromptRangeText(lang i18n.Lang) string {
	return i18n.T(lang, i18n.KeyChallengeCreateRangeIntro)
}

func CreateCalendarHeaderText(lang i18n.Lang) string {
	return i18n.T(lang, i18n.KeyChallengeCreateCalendarHeader)
}

func CreatedText(lang i18n.Lang, name string, totalDays int) string {
	return i18n.T(lang, i18n.KeyChallengeCreatedFmt, name, totalDays)
}
