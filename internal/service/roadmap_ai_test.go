package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"tracker-bot/internal/llm"
	"tracker-bot/internal/models"
)

// fakeLLM replies with canned JSON, so these tests cover our validation and
// storage rules rather than any provider's behavior. Reuses
// fakeRoadmapRepo from roadmap_test.go.

type fakeLLM struct {
	reply string
	err   error
	// captured is the last request, for asserting on prompt content.
	captured llm.Request
}

func (f *fakeLLM) Provider() string { return "fake" }
func (f *fakeLLM) Model() string    { return "fake-1" }

func (f *fakeLLM) Complete(_ context.Context, req llm.Request) (string, llm.Usage, error) {
	f.captured = req
	return f.reply, llm.Usage{}, f.err
}

func (f *fakeLLM) CompleteJSON(_ context.Context, req llm.Request, _ llm.Schema, out any) (llm.Usage, error) {
	f.captured = req
	if f.err != nil {
		return llm.Usage{}, f.err
	}
	return llm.Usage{}, json.Unmarshal([]byte(f.reply), out)
}

type fakeRegistry struct {
	client *fakeLLM
}

func (r fakeRegistry) For(llm.Task) (llm.Client, bool) {
	if r.client == nil {
		return nil, false
	}
	return r.client, true
}

func (r fakeRegistry) Enabled() bool { return r.client != nil }

// seedPlan creates a goal with one technology and returns their ids.
func seedPlan(t *testing.T, fake *fakeRoadmapRepo, userID int64) (goalID, techID int64) {
	t.Helper()
	ctx := context.Background()
	goalID, err := fake.CreateGoal(ctx, userID, "reach mid-level")
	if err != nil {
		t.Fatalf("CreateGoal: %v", err)
	}
	techID, err = fake.CreateRoadmap(ctx, userID, goalID, "Kafka")
	if err != nil {
		t.Fatalf("CreateRoadmap: %v", err)
	}
	return goalID, techID
}

func TestAIDisabledWithoutProvider(t *testing.T) {
	fake := newFakeRoadmapRepo()
	srv := NewRoadmapAIService(fake, fakeRegistry{})

	if srv.Enabled() {
		t.Fatal("Enabled must be false with no client")
	}
	if _, _, err := srv.GeneratePlan(context.Background(), 1, 1, "ru"); !errors.Is(err, models.ErrRoadmapAIDisabled) {
		t.Fatalf("GeneratePlan err = %v, want ErrRoadmapAIDisabled", err)
	}
	if _, err := srv.DigestAdvice(context.Background(), 1, []models.RoadmapDigestCard{{ID: 1}}, "ru"); !errors.Is(err, models.ErrRoadmapAIDisabled) {
		t.Fatalf("DigestAdvice err = %v, want ErrRoadmapAIDisabled", err)
	}
	if _, err := srv.QuizCard(context.Background(), 1, 1, "ru"); !errors.Is(err, models.ErrRoadmapAIDisabled) {
		t.Fatalf("QuizCard err = %v, want ErrRoadmapAIDisabled", err)
	}
}

func TestGeneratePlanStoresCards(t *testing.T) {
	fake := newFakeRoadmapRepo()
	_, techID := seedPlan(t, fake, 7)

	client := &fakeLLM{reply: `{"cards":[
		{"text":"What a broker is","kind":"topic","difficulty":1},
		{"text":"Kafka: The Definitive Guide","kind":"book","difficulty":3}
	]}`}
	srv := NewRoadmapAIService(fake, fakeRegistry{client: client})

	added, rejected, err := srv.GeneratePlan(context.Background(), 7, techID, "ru")
	if err != nil {
		t.Fatalf("GeneratePlan: %v", err)
	}
	if added != 2 || rejected != 0 {
		t.Fatalf("added=%d rejected=%d, want 2/0", added, rejected)
	}

	cards, err := fake.ListCards(context.Background(), 7, techID)
	if err != nil {
		t.Fatalf("ListCards: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("stored %d cards, want 2", len(cards))
	}
	if cards[0].Kind != models.RoadmapCardTopic || cards[0].Difficulty != models.RoadmapCardEasy {
		t.Fatalf("first card = %+v, want easy topic", cards[0])
	}
}

func TestGeneratePlanPromptCarriesGoalCriteriaAndExistingCards(t *testing.T) {
	fake := newFakeRoadmapRepo()
	ctx := context.Background()
	_, techID := seedPlan(t, fake, 7)
	if err := fake.SetMasteryCriteria(ctx, 7, techID, "can debug a lagging consumer"); err != nil {
		t.Fatalf("SetMasteryCriteria: %v", err)
	}
	if _, err := fake.AddCards(ctx, techID, []models.RoadmapCardItem{
		{Text: "Partitions and offsets", Kind: models.RoadmapCardTopic, Difficulty: 2},
	}); err != nil {
		t.Fatalf("AddCards: %v", err)
	}

	client := &fakeLLM{reply: `{"cards":[{"text":"Consumer groups","kind":"topic","difficulty":2}]}`}
	srv := NewRoadmapAIService(fake, fakeRegistry{client: client})

	if _, _, err := srv.GeneratePlan(ctx, 7, techID, "ru"); err != nil {
		t.Fatalf("GeneratePlan: %v", err)
	}

	prompt := client.captured.Prompt
	for _, want := range []string{"Kafka", "reach mid-level", "can debug a lagging consumer", "Partitions and offsets"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt is missing %q:\n%s", want, prompt)
		}
	}
	// A second run must extend rather than repeat, which only works if the
	// existing cards were sent.
	if !strings.Contains(client.captured.System, "Russian") {
		t.Fatalf("system prompt does not name the reply language:\n%s", client.captured.System)
	}
}

func TestGeneratePlanRejectsUnusableCardsAndKeepsTheRest(t *testing.T) {
	fake := newFakeRoadmapRepo()
	_, techID := seedPlan(t, fake, 7)

	long := strings.Repeat("x", models.MaxRoadmapCardTextLen+1)
	client := &fakeLLM{reply: `{"cards":[
		{"text":"Good one","kind":"topic","difficulty":2},
		{"text":"","kind":"topic","difficulty":2},
		{"text":"` + long + `","kind":"topic","difficulty":2},
		{"text":"Odd kind and difficulty","kind":"podcast","difficulty":9}
	]}`}
	srv := NewRoadmapAIService(fake, fakeRegistry{client: client})

	added, rejected, err := srv.GeneratePlan(context.Background(), 7, techID, "en")
	if err != nil {
		t.Fatalf("GeneratePlan: %v", err)
	}
	// Empty and over-long text are dropped; an unknown kind and an
	// out-of-range difficulty fall back to topic/medium instead.
	if added != 2 || rejected != 2 {
		t.Fatalf("added=%d rejected=%d, want 2/2", added, rejected)
	}

	cards, _ := fake.ListCards(context.Background(), 7, techID)
	var fallback models.RoadmapCardItem
	for _, c := range cards {
		if c.Text == "Odd kind and difficulty" {
			fallback = c
		}
	}
	if fallback.Kind != models.RoadmapCardTopic || fallback.Difficulty != models.RoadmapCardMedium {
		t.Fatalf("fallback card = %+v, want topic/medium", fallback)
	}
}

func TestGeneratePlanEmptyResultIsAnError(t *testing.T) {
	fake := newFakeRoadmapRepo()
	_, techID := seedPlan(t, fake, 7)

	client := &fakeLLM{reply: `{"cards":[]}`}
	srv := NewRoadmapAIService(fake, fakeRegistry{client: client})

	if _, _, err := srv.GeneratePlan(context.Background(), 7, techID, "en"); !errors.Is(err, models.ErrRoadmapAIEmptyResult) {
		t.Fatalf("err = %v, want ErrRoadmapAIEmptyResult", err)
	}
}

func TestGeneratePlanCapsCardCount(t *testing.T) {
	fake := newFakeRoadmapRepo()
	_, techID := seedPlan(t, fake, 7)

	drafts := make([]string, 0, maxGeneratedCards+5)
	for i := 0; i < maxGeneratedCards+5; i++ {
		drafts = append(drafts, `{"text":"card","kind":"topic","difficulty":1}`)
	}
	client := &fakeLLM{reply: `{"cards":[` + strings.Join(drafts, ",") + `]}`}
	srv := NewRoadmapAIService(fake, fakeRegistry{client: client})

	added, _, err := srv.GeneratePlan(context.Background(), 7, techID, "en")
	if err != nil {
		t.Fatalf("GeneratePlan: %v", err)
	}
	if added != maxGeneratedCards {
		t.Fatalf("added = %d, want the cap %d", added, maxGeneratedCards)
	}
}

func TestAddCardsFromTextAISendsOnlyNonBlankLinesUpToTheCap(t *testing.T) {
	fake := newFakeRoadmapRepo()
	_, techID := seedPlan(t, fake, 7)

	var lines []string
	for i := 0; i < maxTaggedLines+10; i++ {
		lines = append(lines, "line", "")
	}
	client := &fakeLLM{reply: `{"cards":[{"text":"line","kind":"topic","difficulty":2}]}`}
	srv := NewRoadmapAIService(fake, fakeRegistry{client: client})

	if _, _, err := srv.AddCardsFromTextAI(context.Background(), 7, techID, strings.Join(lines, "\n"), "en"); err != nil {
		t.Fatalf("AddCardsFromTextAI: %v", err)
	}
	if got := strings.Count(client.captured.Prompt, "line"); got != maxTaggedLines {
		t.Fatalf("prompt carried %d lines, want the cap %d", got, maxTaggedLines)
	}
}

func TestAddCardsFromTextAIRejectsBlankInput(t *testing.T) {
	fake := newFakeRoadmapRepo()
	_, techID := seedPlan(t, fake, 7)
	srv := NewRoadmapAIService(fake, fakeRegistry{client: &fakeLLM{}})

	if _, _, err := srv.AddCardsFromTextAI(context.Background(), 7, techID, "\n  \n", "en"); !errors.Is(err, models.ErrRoadmapNoCardsParsed) {
		t.Fatalf("err = %v, want ErrRoadmapNoCardsParsed", err)
	}
}

func TestDigestAdviceClampsLength(t *testing.T) {
	fake := newFakeRoadmapRepo()
	_, techID := seedPlan(t, fake, 7)
	if _, err := fake.AddCards(context.Background(), techID, []models.RoadmapCardItem{
		{Text: "Partitions", Kind: models.RoadmapCardTopic, Difficulty: 1},
	}); err != nil {
		t.Fatalf("AddCards: %v", err)
	}

	long := strings.Repeat("а", maxDigestAdviceLen+50)
	client := &fakeLLM{reply: `{"advice":"` + long + `"}`}
	srv := NewRoadmapAIService(fake, fakeRegistry{client: client})

	advice, err := srv.DigestAdvice(context.Background(), 7, []models.RoadmapDigestCard{
		{ID: 1, RoadmapName: "Kafka", Text: "Partitions", Kind: models.RoadmapCardTopic, Difficulty: 1},
	}, "ru")
	if err != nil {
		t.Fatalf("DigestAdvice: %v", err)
	}
	if len([]rune(advice)) != maxDigestAdviceLen {
		t.Fatalf("advice is %d runes, want it clamped to %d", len([]rune(advice)), maxDigestAdviceLen)
	}
	if !strings.Contains(client.captured.Prompt, "Kafka") {
		t.Fatalf("digest prompt is missing the technology name:\n%s", client.captured.Prompt)
	}
}

func TestDigestAdviceWithNoCards(t *testing.T) {
	fake := newFakeRoadmapRepo()
	srv := NewRoadmapAIService(fake, fakeRegistry{client: &fakeLLM{}})

	if _, err := srv.DigestAdvice(context.Background(), 7, nil, "ru"); !errors.Is(err, models.ErrRoadmapAIEmptyResult) {
		t.Fatalf("err = %v, want ErrRoadmapAIEmptyResult", err)
	}
}

func TestQuizCardFindsTheCardAndCarriesItsText(t *testing.T) {
	fake := newFakeRoadmapRepo()
	ctx := context.Background()
	_, techID := seedPlan(t, fake, 7)
	if _, err := fake.AddCards(ctx, techID, []models.RoadmapCardItem{
		{Text: "Consumer groups", Kind: models.RoadmapCardTopic, Difficulty: 2},
	}); err != nil {
		t.Fatalf("AddCards: %v", err)
	}
	cards, _ := fake.ListCards(ctx, 7, techID)

	client := &fakeLLM{reply: `{"question":"Как ребалансировка влияет на порядок обработки?"}`}
	srv := NewRoadmapAIService(fake, fakeRegistry{client: client})

	quiz, err := srv.QuizCard(ctx, 7, cards[0].ID, "ru")
	if err != nil {
		t.Fatalf("QuizCard: %v", err)
	}
	if quiz.CardID != cards[0].ID || quiz.CardText != "Consumer groups" {
		t.Fatalf("quiz = %+v, want it tied to the card", quiz)
	}
	if quiz.Question == "" {
		t.Fatal("question is empty")
	}
	if !strings.Contains(client.captured.Prompt, "Consumer groups") {
		t.Fatalf("quiz prompt is missing the card text:\n%s", client.captured.Prompt)
	}
}

func TestQuizCardUnknownCard(t *testing.T) {
	fake := newFakeRoadmapRepo()
	seedPlan(t, fake, 7)
	srv := NewRoadmapAIService(fake, fakeRegistry{client: &fakeLLM{}})

	if _, err := srv.QuizCard(context.Background(), 7, 9999, "ru"); !errors.Is(err, models.ErrRoadmapCardNotFound) {
		t.Fatalf("err = %v, want ErrRoadmapCardNotFound", err)
	}
}

func TestQuizCardOfAnotherUsersCard(t *testing.T) {
	fake := newFakeRoadmapRepo()
	ctx := context.Background()
	_, techID := seedPlan(t, fake, 7)
	if _, err := fake.AddCards(ctx, techID, []models.RoadmapCardItem{
		{Text: "Consumer groups", Kind: models.RoadmapCardTopic, Difficulty: 2},
	}); err != nil {
		t.Fatalf("AddCards: %v", err)
	}
	cards, _ := fake.ListCards(ctx, 7, techID)

	srv := NewRoadmapAIService(fake, fakeRegistry{client: &fakeLLM{}})
	if _, err := srv.QuizCard(ctx, 8, cards[0].ID, "ru"); !errors.Is(err, models.ErrRoadmapCardNotFound) {
		t.Fatalf("err = %v, want ErrRoadmapCardNotFound for another user's card", err)
	}
}

func TestGradeQuizAnswerVerdicts(t *testing.T) {
	fake := newFakeRoadmapRepo()
	quiz := models.RoadmapQuiz{CardID: 1, CardText: "Consumer groups", Question: "Q?"}

	cases := []struct {
		reply string
		want  models.RoadmapQuizVerdict
	}{
		{`{"verdict":"correct","feedback":"ok"}`, models.RoadmapQuizCorrect},
		{`{"verdict":"WRONG","feedback":"no"}`, models.RoadmapQuizWrong},
		{`{"verdict":"partial","feedback":"almost"}`, models.RoadmapQuizPartial},
		// An unrecognised verdict must not be read as "wrong" — that would
		// misjudge the user on a model slip.
		{`{"verdict":"excellent","feedback":"?"}`, models.RoadmapQuizPartial},
	}
	for _, tc := range cases {
		srv := NewRoadmapAIService(fake, fakeRegistry{client: &fakeLLM{reply: tc.reply}})
		grade, err := srv.GradeQuizAnswer(context.Background(), quiz, "some answer", "ru")
		if err != nil {
			t.Fatalf("GradeQuizAnswer(%s): %v", tc.reply, err)
		}
		if grade.Verdict != tc.want {
			t.Fatalf("reply %s -> verdict %q, want %q", tc.reply, grade.Verdict, tc.want)
		}
	}
}

func TestGradeQuizAnswerRejectsBlankAnswer(t *testing.T) {
	fake := newFakeRoadmapRepo()
	srv := NewRoadmapAIService(fake, fakeRegistry{client: &fakeLLM{}})

	_, err := srv.GradeQuizAnswer(context.Background(), models.RoadmapQuiz{CardID: 1}, "   ", "ru")
	if !errors.Is(err, models.ErrRoadmapAIEmptyResult) {
		t.Fatalf("err = %v, want ErrRoadmapAIEmptyResult", err)
	}
}

func TestGradeQuizAnswerClampsWhatItForwards(t *testing.T) {
	fake := newFakeRoadmapRepo()
	client := &fakeLLM{reply: `{"verdict":"partial","feedback":"ok"}`}
	srv := NewRoadmapAIService(fake, fakeRegistry{client: client})

	answer := strings.Repeat("z", maxQuizAnswerLen+500)
	if _, err := srv.GradeQuizAnswer(context.Background(), models.RoadmapQuiz{CardID: 1, Question: "Q?"}, answer, "ru"); err != nil {
		t.Fatalf("GradeQuizAnswer: %v", err)
	}
	if got := strings.Count(client.captured.Prompt, "z"); got != maxQuizAnswerLen {
		t.Fatalf("forwarded %d chars of the answer, want the cap %d", got, maxQuizAnswerLen)
	}
}

func TestLanguageNameFallsBackToEnglish(t *testing.T) {
	if got := languageName("uk"); got != "Ukrainian" {
		t.Fatalf("languageName(uk) = %q", got)
	}
	if got := languageName("zz"); got != "English" {
		t.Fatalf("languageName(zz) = %q, want English", got)
	}
}
