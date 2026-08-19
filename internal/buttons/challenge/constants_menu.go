package challenge

// Inline callbacks. Text is not yet localized (see internal/i18n) — this
// whole module is planned for i18n Phase 4, once its UI has settled.
const (
	CBCreate = "challenge:create"
	CBOpen   = "challenge:open:" // + challengeID
	// CBArchiveThis archives one challenge — deliberately not a prefix of
	// CBArchiveOpen/CBArchiveRestore/CBArchiveDelete below (a bare
	// "challenge:archive:" would shadow all three in the dispatcher's
	// prefix-matching switch).
	CBArchiveThis = "challenge:archive:this:" // + challengeID

	CBArchiveOpen    = "challenge:archive:open"
	CBArchiveRestore = "challenge:archive:restore:" // + challengeID
	CBArchiveDelete  = "challenge:archive:delete:"  // + challengeID

	// CBDayOpen carries "<challengeID>:<2006-01-02>".
	CBDayOpen = "challenge:day:open:"
	CBDayDone = "challenge:day:done:"
	CBDaySkip = "challenge:day:skip:"

	CBBackList = "challenge:back:list"

	// Calendar range-picker shown while creating a challenge — same shape
	// as Track's period-report calendar, own prefix so the two don't clash.
	CBCalPrev     = "challenge:cal:prev"
	CBCalNext     = "challenge:cal:next"
	CBCalPrevYear = "challenge:cal:prev_year"
	CBCalNextYear = "challenge:cal:next_year"
	CBCalPick     = "challenge:cal:pick:" // + 2006-01-02
	CBCalDone     = "challenge:cal:done"
	CBCalCancel   = "challenge:cal:cancel"

	// Evening push callbacks, sent standalone by the scheduler — carry
	// "<challengeID>", not gated by screen.
	CBPushDone = "challenge:push:done:"
	CBPushSkip = "challenge:push:skip:"
)

// Inline menu buttons.
const (
	ButtonCreate  = "➕ New challenge"
	ButtonArchive = "🔁 Archive"
	ButtonHome    = "🏠 Home"
	ButtonBack    = "◀ Back"

	ButtonRestore       = "♻️ Restore"
	ButtonDeleteForever = "🗑 Delete forever"

	ButtonMarkDone = "✅ Done"
	ButtonMarkSkip = "❌ Skipped"

	ButtonSetRange  = "📅 Pick dates"
	ButtonSelectEnd = "Now pick the end date"
	ButtonConfirm   = "✅ Confirm range"
	ButtonCancel    = "❌ Cancel"
)

// "Waiting for input" reply-keyboard buttons.
const (
	ReplyCancel = "❌ Cancel"
)
