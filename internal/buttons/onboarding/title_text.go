package onboarding

// stepTexts holds the tour's content, one entry per step (see StepCount).
var stepTexts = [StepCount]string{
	"👋 *Here's what you can do here:*\n\n" +
		"📈 *Track* — log time on real-life activities, get reminded, see reports\n" +
		"🧠 *Learning* — build word collections, review them on a smart schedule\n" +
		"🎯 *Challenges* — set a goal for N days, check off each day, watch your progress\n\n" +
		"Let's go through them one by one.",

	"📈 *Track*\n\n" +
		"• Create activities (e.g. \"Work\", \"Reading\")\n" +
		"• Start a timer (15/30 min, or your own) — the bot pings you to log what you did\n" +
		"• Reports: today, any date range, hour-by-hour breakdown\n" +
		"• 🔥 Heatmap — a calendar of which days you tracked anything; tap a day to see what you did",

	"🧠 *Learning*\n\n" +
		"• Create a word collection, paste pairs as \"word - translation\"\n" +
		"• Pick which collections are in rotation and how often you want reviews\n" +
		"• The bot pushes you one word at a time — mark whether you knew it\n" +
		"• Real spaced repetition: words you know well show up less, tricky ones more often",

	"🎯 *Challenges*\n\n" +
		"• Set a goal for N days (e.g. \"100 days of reading\") — pick a start/end date\n" +
		"• Every day in the range gets its own square\n" +
		"• Each evening the bot asks if you did it — tap ✅ or ❌\n" +
		"• Missed a day? Go back and mark it any time\n" +
		"• Watch your progress fill in as a %",

	"✅ *That's the tour!*\n\nJump straight in, or explore anytime from 🏠 Home.",
}

// StepText returns one tour step's message text, clamped to a valid index.
func StepText(step int) string {
	if step < 0 {
		step = 0
	}
	if step >= StepCount {
		step = StepCount - 1
	}
	return stepTexts[step]
}
