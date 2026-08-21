package learning

const (
	LearningCBCreateCollection  = "learning:collection:create"
	LearningCBWordBase          = "learning:wordbase:open"
	LearningCBCollectionOpen    = "learning:collection:open:"
	LearningCBCollectionToggle  = "learning:collection:toggle:"
	LearningCBCollectionAddMore = "learning:collection:addwords:"
	LearningCBCollectionRename  = "learning:collection:rename:"
	LearningCBCollectionArchive = "learning:collection:archive:"
	LearningCBWordDelete        = "learning:word:delete:"
	LearningCBArchiveOpen       = "learning:archive:open"
	LearningCBArchiveRestore    = "learning:archive:restore:"
	LearningCBArchiveDelete     = "learning:archive:delete:"
	LearningCBStats             = "learning:stats:open"
	LearningCBReviewOpen        = "learning:review:open"
	LearningCBReviewStop        = "learning:review:stop"
	LearningCBBackMain          = "learning:back:main"

	// shown before the interval picker so the user explicitly picks collections, instead of silently defaulting to all active ones
	LearningCBReviewPickToggle = "learning:review:pick:toggle:"
	LearningCBReviewContinue   = "learning:review:pick:continue"

	// sent standalone (scheduler push or on-demand), not gated by the current screen; grading is Anki-style (4 levels, see models.LearningGrade)
	LearningCBReviewReveal = "learning:review:reveal:"
	LearningCBReviewAgain  = "learning:review:grade:again:"
	LearningCBReviewHard   = "learning:review:grade:hard:"
	LearningCBReviewGood   = "learning:review:grade:good:"
	LearningCBReviewEasy   = "learning:review:grade:easy:"
)

var BuiltInPushIntervals = []int{30, 60, 120}
