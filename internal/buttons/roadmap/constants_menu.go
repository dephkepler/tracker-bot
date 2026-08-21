package roadmap

// Inline callbacks, all namespaced under "roadmap:" so the dispatcher routes
// the whole domain on one prefix check.
//
// Prefixes here are deliberately non-overlapping: the dispatcher matches with
// strings.HasPrefix, so a shorter prefix would shadow a longer one sharing
// its start (hence "roadmap:archive:this:" rather than a bare
// "roadmap:archive:" — same trap documented in
// internal/buttons/challenge/constants_menu.go).
const (
	RoadmapCBBackMain = "roadmap:back:main"
	RoadmapCBStats    = "roadmap:stats:open"
	RoadmapCBPushOpen = "roadmap:push:open"
	RoadmapCBPushStop = "roadmap:push:stop"

	// Goals.
	RoadmapCBGoalsOpen   = "roadmap:goals:open"
	RoadmapCBGoalCreate  = "roadmap:goal:create"
	RoadmapCBGoalOpen    = "roadmap:goal:open:"
	RoadmapCBGoalRename  = "roadmap:goal:rename:"
	RoadmapCBGoalArchive = "roadmap:goal:archive:"

	// Technologies.
	RoadmapCBTechCreate  = "roadmap:tech:create:"
	RoadmapCBOpen        = "roadmap:open:"
	RoadmapCBToggle      = "roadmap:toggle:"
	RoadmapCBAddCards    = "roadmap:addcards:"
	RoadmapCBRename      = "roadmap:rename:"
	RoadmapCBSetCriteria = "roadmap:criteria:"
	RoadmapCBArchiveThis = "roadmap:archive:this:"
	RoadmapCBOrphansOpen = "roadmap:orphans:open"
	RoadmapCBAssignOpen  = "roadmap:assign:open:"
	// Carries "<roadmapID>:<goalID>" — see ParseAssignPayload.
	RoadmapCBAssignPick = "roadmap:assign:pick:"

	// Archive.
	RoadmapCBArchiveOpen        = "roadmap:archive:open"
	RoadmapCBArchiveRestore     = "roadmap:archive:restore:"
	RoadmapCBArchiveDelete      = "roadmap:archive:delete:"
	RoadmapCBArchiveGoalRestore = "roadmap:archive:goalrestore:"
	RoadmapCBArchiveGoalDelete  = "roadmap:archive:goaldelete:"

	// Cards. RoadmapCBCardToggle ticks a card from its technology's
	// checklist; RoadmapCBDigestToggle does the same flip from a reminder
	// push, a separate prefix purely so the handler knows which message to
	// re-render. Both work without any particular screen: the card id
	// resolves its own technology (see repo.ToggleCardDone).
	RoadmapCBCardToggle   = "roadmap:card:toggle:"
	RoadmapCBCardDiff     = "roadmap:card:diff:"
	RoadmapCBCardDelete   = "roadmap:card:delete:"
	RoadmapCBDigestToggle = "roadmap:digest:toggle:"

	// AI-backed actions, all under "roadmap:ai:". None of these buttons is
	// drawn unless a provider is configured, so a handler reaching one with
	// AI off means a stale keyboard — hence the guard on each.
	//
	// "plan" and "paste" carry a technology id, "quiz" a card id.
	RoadmapCBAIPlan  = "roadmap:ai:plan:"
	RoadmapCBAIPaste = "roadmap:ai:paste:"
	RoadmapCBAIQuiz  = "roadmap:ai:quiz:"
)

// BuiltInPushIntervals are the reminder-interval choices (minutes) offered in
// the reply-keyboard picker. Deliberately longer than Learning's
// {30, 60, 120}: checking in on a technology roadmap is far less
// time-sensitive than a word review, so hourly is the shortest sensible step.
var BuiltInPushIntervals = []int{60, 180, 360}
