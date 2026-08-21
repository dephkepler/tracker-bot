package challenge

// Callback constants. Button text moved to i18n (see keys.go's Challenge
// block) — this file now only holds the routing data.
const (
	CBCreate = "challenge:create"
	CBOpen   = "challenge:open:" // + challengeID
	// Deliberately not a shared "challenge:archive:" prefix — that would shadow
	// CBArchiveOpen/CBArchiveRestore/CBArchiveDelete in the dispatcher's prefix switch.
	CBArchiveThis = "challenge:archive:this:" // + challengeID

	CBArchiveOpen    = "challenge:archive:open"
	CBArchiveRestore = "challenge:archive:restore:" // + challengeID
	CBArchiveDelete  = "challenge:archive:delete:"  // + challengeID

	// CBDayOpen carries "<challengeID>:<2006-01-02>".
	CBDayOpen = "challenge:day:open:"
	CBDayDone = "challenge:day:done:"
	CBDaySkip = "challenge:day:skip:"

	CBBackList = "challenge:back:list"

	// Own prefix so this calendar doesn't clash with Track's period-report calendar.
	CBCalPrev     = "challenge:cal:prev"
	CBCalNext     = "challenge:cal:next"
	CBCalPrevYear = "challenge:cal:prev_year"
	CBCalNextYear = "challenge:cal:next_year"
	CBCalPick     = "challenge:cal:pick:" // + 2006-01-02
	CBCalDone     = "challenge:cal:done"
	CBCalCancel   = "challenge:cal:cancel"

	// Sent standalone by the scheduler; not gated by screen.
	CBPushDone = "challenge:push:done:"
	CBPushSkip = "challenge:push:skip:"
)
