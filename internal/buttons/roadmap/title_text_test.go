package roadmap

import (
	"strings"
	"testing"

	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
)

func TestRoadmapDetailTextListsCardsInFull(t *testing.T) {
	// The bug this guards: a generated card runs to a hundred characters or
	// more, an inline button is about thirty-five wide and four share a row, so
	// the card used to arrive as "Re…tables,…". It belongs in the message text,
	// where there is room to read it.
	const long = "Relational database basics: tables, rows, columns, primary keys, and how a database differs from a spreadsheet"

	text := RoadmapDetailText(i18n.EN, models.RoadmapItem{Name: "SQL", TotalCards: 2, DoneCards: 1}, []models.RoadmapCardItem{
		{ID: 1, Text: long, Kind: models.RoadmapCardTopic, Difficulty: 1, IsDone: true},
		{ID: 2, Text: "Window functions", Kind: models.RoadmapCardBook, Difficulty: 3},
	})

	if !strings.Contains(text, long) {
		t.Fatalf("the card text was not written out in full:\n%s", text)
	}
	// Numbered, because the buttons below carry those numbers.
	if !strings.Contains(text, "1. ") || !strings.Contains(text, "2. ") {
		t.Fatalf("cards are not numbered:\n%s", text)
	}
	// A finished card shows a tick instead of its difficulty.
	if !strings.Contains(text, "✅") {
		t.Fatalf("no tick for the finished card:\n%s", text)
	}
}

func TestRoadmapDetailTextWithNoCards(t *testing.T) {
	text := RoadmapDetailText(i18n.EN, models.RoadmapItem{Name: "SQL"}, nil)

	if strings.Contains(text, "1. ") {
		t.Fatalf("an empty technology got a card list:\n%s", text)
	}
	if text == "" {
		t.Fatal("no text at all")
	}
}

// A long paste times many cards must not push the message past Telegram's
// 4096-character limit, which would make the whole screen fail to send.
func TestRoadmapDetailTextStaysWithinTelegramsLimit(t *testing.T) {
	cards := make([]models.RoadmapCardItem, 0, 100)
	for i := range 100 {
		cards = append(cards, models.RoadmapCardItem{
			ID:         int64(i + 1),
			Text:       strings.Repeat("длинная карточка ", 20), // ~340 runes
			Kind:       models.RoadmapCardTopic,
			Difficulty: 2,
		})
	}

	text := RoadmapDetailText(i18n.RU, models.RoadmapItem{Name: "SQL", TotalCards: 100}, cards)

	if runes := len([]rune(text)); runes > 4000 {
		t.Fatalf("message is %d runes — Telegram would reject it", runes)
	}
	// The ones not spelled out are still accounted for.
	if !strings.Contains(text, "ещё") {
		t.Fatalf("the remaining cards are not mentioned:\n%s", text[:200])
	}
}

func TestCardButtonLabelIsJustTheNumberAndState(t *testing.T) {
	label := CardButtonLabel(3, models.RoadmapCardItem{
		Text:       "a card whose text is far too long for any button",
		Kind:       models.RoadmapCardBook,
		Difficulty: 3,
	})

	if strings.Contains(label, "card whose text") {
		t.Fatalf("the card text is back on the button: %q", label)
	}
	if !strings.HasPrefix(label, "3 ") {
		t.Fatalf("label = %q, want it to start with its list number", label)
	}
	// Short enough that four buttons share a row without eliding.
	if runes := len([]rune(label)); runes > 6 {
		t.Fatalf("label %q is %d runes — too wide to share a row", label, runes)
	}
}
