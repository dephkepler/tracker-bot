package learning

// Inline callbacks.
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

	// Collection picker shown before the interval picker, so the user
	// explicitly chooses which collections feed the review rotation
	// instead of it silently defaulting to "every active collection".
	LearningCBReviewPickToggle = "learning:review:pick:toggle:"
	LearningCBReviewContinue   = "learning:review:pick:continue"

	// Review-card callbacks, sent standalone (scheduler push or on-demand),
	// not gated by screen — carry the word id in the payload. Grading is
	// Anki-style: four levels instead of a binary knew-it/missed-it (see
	// models.LearningGrade).
	LearningCBReviewReveal = "learning:review:reveal:"
	LearningCBReviewAgain  = "learning:review:grade:again:"
	LearningCBReviewHard   = "learning:review:grade:hard:"
	LearningCBReviewGood   = "learning:review:grade:good:"
	LearningCBReviewEasy   = "learning:review:grade:easy:"
)

// BuiltInPushIntervals are the always-available review-push interval
// choices (minutes), shown in the reply-keyboard picker.
var BuiltInPushIntervals = []int{30, 60, 120}
