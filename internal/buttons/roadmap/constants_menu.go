package roadmap

// Inline callbacks. All colon-namespaced under "roadmap:" so the dispatcher
// can route the whole domain on one prefix check.
//
// Note the deliberately longer "roadmap:archive:this:" — a bare
// "roadmap:archive:" would shadow the archive-screen callbacks below in the
// dispatcher's strings.HasPrefix switch (same trap documented in
// internal/buttons/challenge/constants_menu.go).
const (
	RoadmapCBCreate      = "roadmap:create"
	RoadmapCBList        = "roadmap:list:open"
	RoadmapCBOpen        = "roadmap:open:"
	RoadmapCBToggle      = "roadmap:toggle:"
	RoadmapCBAddCards    = "roadmap:addcards:"
	RoadmapCBRename      = "roadmap:rename:"
	RoadmapCBSetGoal     = "roadmap:goal:"
	RoadmapCBArchiveThis = "roadmap:archive:this:"

	RoadmapCBArchiveOpen    = "roadmap:archive:open"
	RoadmapCBArchiveRestore = "roadmap:archive:restore:"
	RoadmapCBArchiveDelete  = "roadmap:archive:delete:"

	RoadmapCBCardDelete = "roadmap:card:delete:"

	RoadmapCBStats    = "roadmap:stats:open"
	RoadmapCBPushOpen = "roadmap:push:open"
	RoadmapCBPushStop = "roadmap:push:stop"
	RoadmapCBBackMain = "roadmap:back:main"

	// RoadmapCBCardToggle ticks/unticks a card from its roadmap's checklist
	// screen. RoadmapCBDigestToggle does the same flip from a reminder push
	// message — a separate prefix purely so the handler knows which message
	// to re-render (the checklist, or the digest itself). Both are handled
	// without requiring any particular screen: the card id in the payload is
	// enough to resolve its roadmap (see repo.ToggleCardDone).
	RoadmapCBCardToggle   = "roadmap:card:toggle:"
	RoadmapCBDigestToggle = "roadmap:digest:toggle:"
)

// BuiltInPushIntervals are the always-available reminder-interval choices
// (minutes), shown in the reply-keyboard picker. Deliberately longer than
// Learning's {30, 60, 120}: checking in on a technology roadmap is far less
// time-sensitive than a word review, so hourly is the shortest sensible
// step here.
var BuiltInPushIntervals = []int{60, 180, 360}
