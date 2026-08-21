package service

import (
	"fmt"
	"strings"

	"tracker-bot/internal/models"
)

// Prompts live apart from the service so the wording can be edited without
// touching the call plumbing. They are English regardless of the user's
// language: the instruction names the reply language, and mixing the two
// makes the constraint easy to miss.

// languageNames maps the language codes the bot supports (i18n.Lang) to what
// to call them in a prompt. Anything unknown falls back to English, matching
// i18n.Default.
var languageNames = map[string]string{
	"ru": "Russian",
	"en": "English",
	"de": "German",
	"uk": "Ukrainian",
	"ar": "Arabic",
}

func languageName(lang string) string {
	if name, ok := languageNames[strings.ToLower(strings.TrimSpace(lang))]; ok {
		return name
	}
	return "English"
}

func replyIn(lang string) string {
	return fmt.Sprintf("Write every piece of user-facing text in %s.", languageName(lang))
}

func planSystemPrompt(lang string) string {
	return fmt.Sprintf(`You design learning checklists for a self-study tracker.

Given a technology, the outcome the learner is working toward, and what they
consider "I know this", produce the cards that get them there.

Rules:
- Between 8 and %d cards, ordered easiest first.
- Each card is one concrete thing to study or do, not a whole subject area.
- Use kind "topic" for something to learn, and "article", "book" or
  "lecture" only for a specific source you can name. Do not invent URLs.
- Difficulty is relative to this technology: 1 easy, 2 medium, 3 hard.
- Do not repeat cards the learner already has; extend the plan instead.
- %s`, maxGeneratedCards, replyIn(lang))
}

func planUserPrompt(goalName string, tech models.RoadmapItem, existing []models.RoadmapCardItem) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Technology: %s\n", tech.Name)
	if goalName != "" {
		fmt.Fprintf(&b, "Goal it feeds into: %s\n", goalName)
	}
	if tech.MasteryCriteria != "" {
		fmt.Fprintf(&b, "What the learner means by knowing it: %s\n", tech.MasteryCriteria)
	}
	if len(existing) > 0 {
		b.WriteString("\nCards they already have (do not repeat):\n")
		for _, card := range existing {
			status := "pending"
			if card.IsDone {
				status = "done"
			}
			fmt.Fprintf(&b, "- [%s] %s\n", status, card.Text)
		}
	}
	return b.String()
}

func taggingSystemPrompt(lang string) string {
	return fmt.Sprintf(`The learner pasted lines from their notes into a study checklist.

Return one card per line, keeping the lines in the order given and keeping
each line's own wording — you are classifying, not rewriting. Drop a line
only if there is nothing studiable in it.

For each: kind is "book", "article" or "lecture" when the line names a
specific source (a URL is almost always an article), otherwise "topic".
Difficulty is 1 easy, 2 medium, 3 hard, judged against the other lines.

Keep each line's original language; %s`, strings.ToLower(replyIn(lang)))
}

func digestSystemPrompt(lang string) string {
	return fmt.Sprintf(`You write the nudge under a study reminder.

You get the learner's pending cards, easiest first, and their overall
progress. Say which one to take next and why, in two or three sentences.
Concrete and plain — no motivational filler, no restating the list.

%s`, replyIn(lang))
}

func digestUserPrompt(cards []models.RoadmapDigestCard, total, done int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Progress: %d of %d cards done.\n\nPending, easiest first:\n", done, total)
	for _, card := range cards {
		fmt.Fprintf(&b, "- [%s, difficulty %d] %s — %s\n",
			card.Kind, card.Difficulty, card.RoadmapName, card.Text)
	}
	return b.String()
}

func quizSystemPrompt(lang string) string {
	return fmt.Sprintf(`Ask one question that checks whether the learner
actually understands the card's subject.

Open question, answerable in a few sentences from understanding rather than
recall of a definition. No multiple choice, no yes/no. Ask only the
question — no preamble, no answer.

%s`, replyIn(lang))
}

func gradeSystemPrompt(lang string) string {
	return fmt.Sprintf(`Judge the learner's answer to a question about a
study card.

"correct" — the answer covers the point, even if worded loosely.
"partial" — the idea is there but something important is missing or muddled.
"wrong" — the answer misses or contradicts the point.

Feedback names what was missing or wrong, or what to look at next. Be
straight about a wrong answer; do not inflate the verdict to be kind.

%s`, replyIn(lang))
}
