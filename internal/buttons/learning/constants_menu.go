package learning

// Inline callbacks. Text is not yet localized (see internal/i18n) — this
// whole module is planned for i18n Phase 4, once its UI has settled.
const (
	LearningCBCreateCollection  = "learning:collection:create"
	LearningCBWordBase          = "learning:wordbase:open"
	LearningCBCollectionOpen    = "learning:collection:open:"
	LearningCBCollectionToggle  = "learning:collection:toggle:"
	LearningCBCollectionAddMore = "learning:collection:addwords:"
	LearningCBCollectionArchive = "learning:collection:archive:"
	LearningCBWordDelete        = "learning:word:delete:"
	LearningCBArchiveOpen       = "learning:archive:open"
	LearningCBArchiveRestore    = "learning:archive:restore:"
	LearningCBArchiveDelete     = "learning:archive:delete:"
	LearningCBStats             = "learning:stats:open"
	LearningCBReviewOpen        = "learning:review:open"
	LearningCBReviewStop        = "learning:review:stop"
	LearningCBBackMain          = "learning:back:main"

	// Collection picker shown before the interval picker, so the user
	// explicitly chooses which collections feed the review rotation
	// instead of it silently defaulting to "every active collection".
	LearningCBReviewPickToggle = "learning:review:pick:toggle:"
	LearningCBReviewContinue   = "learning:review:pick:continue"

	// Review-card callbacks, sent standalone (scheduler push or on-demand),
	// not gated by screen — carry the word id in the payload.
	LearningCBReviewReveal = "learning:review:reveal:"
	LearningCBReviewKnew   = "learning:review:grade:knew:"
	LearningCBReviewMissed = "learning:review:grade:missed:"
)

// Inline menu buttons.
const (
	LearningButtonCreateCollection = "➕ Create a collection"
	LearningButtonStartReviews     = "🎲 Start reviews"
	LearningButtonStopReviews      = "⏹ Stop reviews"
	LearningButtonArchive          = "🔁 Archive of collections"
	LearningButtonStatistics       = "📈 Statistics"
	LearningButtonWordBase         = "🗂 Word base"
	LearningButtonHome             = "🏠 Home"
	LearningButtonBack             = "◀ Back"
	LearningButtonContinue         = "▶️ Continue"
)

// Word-base / collection-detail buttons.
const (
	LearningButtonAddWords     = "➕ Add words"
	LearningButtonArchiveThis  = "📦 Archive collection"
	LearningButtonToggleOnFmt  = "🟢 %s — %d words"
	LearningButtonToggleOffFmt = "⚪ %s — %d words"
)

// Archive-screen buttons.
const (
	LearningButtonRestore       = "♻️ Restore"
	LearningButtonDeleteForever = "🗑 Delete forever"
)

// "Waiting for input" reply-keyboard buttons.
const (
	LearningButtonDone   = "✅ Done"
	LearningButtonCancel = "❌ Cancel"
)

// Review-card buttons (push message / on-demand review).
const (
	LearningButtonShowAnswer = "👁 Show answer"
	LearningButtonKnewIt     = "✅ I knew it"
	LearningButtonMissedIt   = "🔁 Didn't know"
)

// BuiltInPushIntervals are the always-available review-push interval
// choices (minutes), shown in the reply-keyboard picker.
var BuiltInPushIntervals = []int{30, 60, 120}
