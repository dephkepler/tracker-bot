package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tracker-bot/internal/llm"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
)

// RoadmapAIService is the AI half of the roadmap feature. Every method has a
// manual counterpart on RoadmapService, and every method degrades to
// ErrRoadmapAIDisabled when no provider is configured — the bot must stay
// fully usable without an API key.
type RoadmapAIService interface {
	// Enabled reports whether AI features should be offered in the UI at
	// all. False hides the buttons rather than showing ones that error.
	Enabled() bool

	// GeneratePlan writes a checklist for one technology from its name, its
	// goal and its mastery criteria, and appends the cards. Returns how many
	// were added and how many the model produced but we rejected.
	GeneratePlan(ctx context.Context, userID, roadmapID int64, lang string) (added, rejected int, err error)

	// AddCardsFromTextAI is AddCardsFromText with the kind and difficulty
	// decided by the model instead of by inline #kind / !difficulty tags.
	AddCardsFromTextAI(ctx context.Context, userID, roadmapID int64, text, lang string) (added, rejected int, err error)

	// DigestAdvice writes the "what to take next and why" note that goes
	// under the pending-card list in a push.
	DigestAdvice(ctx context.Context, userID int64, cards []models.RoadmapDigestCard, lang string) (string, error)

	// QuizCard asks one question about a card.
	QuizCard(ctx context.Context, userID, cardID int64, lang string) (models.RoadmapQuiz, error)

	// GradeQuizAnswer judges an answer to a question QuizCard produced.
	GradeQuizAnswer(ctx context.Context, quiz models.RoadmapQuiz, answer, lang string) (models.RoadmapQuizGrade, error)
}

// Per-call ceilings. These are wall-clock budgets, not token budgets: a
// Telegram handler is holding a "thinking..." message open while they run,
// and the schedulers push on a fixed interval.
const (
	planTimeout    = 120 * time.Second
	taggingTimeout = 60 * time.Second
	digestTimeout  = 45 * time.Second
	quizTimeout    = 45 * time.Second
)

const (
	// maxGeneratedCards bounds one generated plan. Twenty cards is already
	// more checklist than anyone works through in a sitting, and the cap
	// keeps a runaway reply from filling a technology in one go.
	maxGeneratedCards = 20
	// maxGeneratedCardTextLen is what the model is told to stay under, and it
	// is far below models.MaxRoadmapCardTextLen on purpose. The database limit
	// is 300 because a person pasting their own notes may write a long line;
	// but a generated card is read in a list beside a tappable number, and
	// telling the model the storage limit produced 120-character sentences that
	// were unreadable in the interface. The validator still accepts up to the
	// database limit — this is guidance for the generator, not a new rule.
	maxGeneratedCardTextLen = 70
	// maxTaggedLines bounds one AI-assisted paste. Past this the paste is
	// tagged by the deterministic parser instead — an unbounded paste would
	// be an unbounded prompt.
	maxTaggedLines = 40
	// maxDigestAdviceLen keeps the advice note to something that fits above
	// the card list in a Telegram message rather than pushing it offscreen.
	maxDigestAdviceLen = 600
	maxQuizQuestionLen = 400
	maxQuizFeedbackLen = 600
	// maxQuizAnswerLen bounds what we forward for grading.
	maxQuizAnswerLen = 2000
)

// llmRegistry is what the AI service needs from llm.Registry. Narrowed to an
// interface so a test can hand in canned replies instead of a live provider.
type llmRegistry interface {
	For(task llm.Task) (llm.Client, bool)
	Enabled() bool
}

type roadmapAIService struct {
	repo repo.RoadmapRepository
	llm  llmRegistry
}

// NewRoadmapAIService takes the same repository as RoadmapService: the AI
// paths read the plan for context and write cards through the same
// AddCards, so the caps and constraints there apply unchanged.
func NewRoadmapAIService(repo repo.RoadmapRepository, registry llmRegistry) RoadmapAIService {
	return &roadmapAIService{repo: repo, llm: registry}
}

func (srv *roadmapAIService) Enabled() bool { return srv.llm.Enabled() }

// cardDraft is what the model returns for one card. Kind and difficulty are
// validated on the way in — a model naming a kind the DB CHECK does not
// allow would otherwise fail the insert for the whole batch.
type cardDraft struct {
	Text       string `json:"text"`
	Kind       string `json:"kind"`
	Difficulty int    `json:"difficulty"`
}

type cardDraftList struct {
	Cards []cardDraft `json:"cards"`
}

// cardsSchema is shared by plan generation and tagging: both produce the
// same shape, and keeping one schema keeps the two prompts honest about it.
var cardsSchema = llm.Schema{
	Name:        "roadmap_cards",
	Description: "The checklist cards.",
	Properties: map[string]any{
		"cards": map[string]any{
			"type":        "array",
			"description": "One entry per card, ordered easiest first.",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"text", "kind", "difficulty"},
				"properties": map[string]any{
					"text": map[string]any{
						"type":        "string",
						"description": fmt.Sprintf("What to study — a short phrase under %d characters, no list marker and no tags.", maxGeneratedCardTextLen),
					},
					"kind": map[string]any{
						"type":        "string",
						"enum":        []string{"topic", "article", "book", "lecture"},
						"description": "topic for something to learn, article/book/lecture for a specific named source.",
					},
					"difficulty": map[string]any{
						"type":        "integer",
						"enum":        []int{1, 2, 3},
						"description": "1 easy, 2 medium, 3 hard.",
					},
				},
			},
		},
	},
	Required: []string{"cards"},
}

func (srv *roadmapAIService) GeneratePlan(ctx context.Context, userID, roadmapID int64, lang string) (int, int, error) {
	client, ok := srv.llm.For(llm.TaskPlan)
	if !ok {
		return 0, 0, models.ErrRoadmapAIDisabled
	}

	tech, err := srv.repo.GetRoadmap(ctx, userID, roadmapID)
	if err != nil {
		return 0, 0, err
	}

	goalName := ""
	if tech.GoalID != nil {
		// A technology can hang off no goal at all (a v1 leftover, or its
		// goal was deleted), in which case the plan is written from the
		// technology alone.
		if goal, err := srv.repo.GetGoal(ctx, userID, *tech.GoalID); err == nil {
			goalName = goal.Name
		}
	}

	// Existing cards go into the prompt so a second run extends the plan
	// instead of repeating it.
	existing, err := srv.repo.ListCards(ctx, userID, roadmapID)
	if err != nil {
		return 0, 0, err
	}

	ctx, cancel := context.WithTimeout(ctx, planTimeout)
	defer cancel()

	var out cardDraftList
	if _, err := client.CompleteJSON(ctx, llm.Request{
		System: planSystemPrompt(lang),
		Prompt: planUserPrompt(goalName, tech, existing),
		Effort: llm.EffortHigh,
	}, cardsSchema, &out); err != nil {
		return 0, 0, fmt.Errorf("generate plan: %w", err)
	}

	return srv.storeDrafts(ctx, roadmapID, out.Cards)
}

func (srv *roadmapAIService) AddCardsFromTextAI(ctx context.Context, userID, roadmapID int64, text, lang string) (int, int, error) {
	client, ok := srv.llm.For(llm.TaskTagging)
	if !ok {
		return 0, 0, models.ErrRoadmapAIDisabled
	}
	if _, err := srv.repo.GetRoadmap(ctx, userID, roadmapID); err != nil {
		return 0, 0, err
	}

	lines := make([]string, 0, maxTaggedLines)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) == maxTaggedLines {
			break
		}
	}
	if len(lines) == 0 {
		return 0, 0, models.ErrRoadmapNoCardsParsed
	}

	ctx, cancel := context.WithTimeout(ctx, taggingTimeout)
	defer cancel()

	var out cardDraftList
	if _, err := client.CompleteJSON(ctx, llm.Request{
		System: taggingSystemPrompt(lang),
		Prompt: "Lines:\n" + strings.Join(lines, "\n"),
		Effort: llm.EffortLow,
	}, cardsSchema, &out); err != nil {
		return 0, 0, fmt.Errorf("tag cards: %w", err)
	}

	return srv.storeDrafts(ctx, roadmapID, out.Cards)
}

// storeDrafts validates drafts and inserts what survives. Invalid drafts are
// counted and dropped rather than failing the batch: one bad line out of
// fifteen should not cost the user the other fourteen.
func (srv *roadmapAIService) storeDrafts(ctx context.Context, roadmapID int64, drafts []cardDraft) (added, rejected int, err error) {
	cards := make([]models.RoadmapCardItem, 0, len(drafts))
	for _, d := range drafts {
		card, ok := validateDraft(d)
		if !ok {
			rejected++
			continue
		}
		cards = append(cards, card)
		if len(cards) == maxGeneratedCards {
			break
		}
	}
	if len(cards) == 0 {
		return 0, rejected, models.ErrRoadmapAIEmptyResult
	}

	added, err = srv.repo.AddCards(ctx, roadmapID, cards)
	if err != nil {
		return 0, rejected, err
	}
	return added, rejected, nil
}

func validateDraft(d cardDraft) (models.RoadmapCardItem, bool) {
	text := strings.TrimSpace(d.Text)
	if text == "" || len([]rune(text)) > models.MaxRoadmapCardTextLen {
		return models.RoadmapCardItem{}, false
	}

	kind := models.RoadmapCardKind(strings.ToLower(strings.TrimSpace(d.Kind)))
	switch kind {
	case models.RoadmapCardTopic, models.RoadmapCardArticle, models.RoadmapCardBook, models.RoadmapCardLecture:
	default:
		// Same default as a plain pasted line: an unrecognised kind is a
		// topic rather than a rejected card.
		kind = models.RoadmapCardTopic
	}

	difficulty := d.Difficulty
	if difficulty < models.RoadmapCardEasy || difficulty > models.RoadmapCardHard {
		difficulty = models.RoadmapCardMedium
	}

	return models.RoadmapCardItem{Text: text, Kind: kind, Difficulty: difficulty}, true
}

var digestSchema = llm.Schema{
	Name:        "digest_advice",
	Description: "The note shown under the pending cards.",
	Properties: map[string]any{
		"advice": map[string]any{
			"type":        "string",
			"description": fmt.Sprintf("At most %d characters. Which card to take next and why, in two or three sentences.", maxDigestAdviceLen),
		},
	},
	Required: []string{"advice"},
}

func (srv *roadmapAIService) DigestAdvice(ctx context.Context, userID int64, cards []models.RoadmapDigestCard, lang string) (string, error) {
	client, ok := srv.llm.For(llm.TaskDigest)
	if !ok {
		return "", models.ErrRoadmapAIDisabled
	}
	if len(cards) == 0 {
		return "", models.ErrRoadmapAIEmptyResult
	}

	total, done, err := srv.repo.CountCards(ctx, userID)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, digestTimeout)
	defer cancel()

	var out struct {
		Advice string `json:"advice"`
	}
	if _, err := client.CompleteJSON(ctx, llm.Request{
		System: digestSystemPrompt(lang),
		Prompt: digestUserPrompt(cards, total, done),
		Effort: llm.EffortLow,
	}, digestSchema, &out); err != nil {
		return "", fmt.Errorf("digest advice: %w", err)
	}

	advice := clampRunes(strings.TrimSpace(out.Advice), maxDigestAdviceLen)
	if advice == "" {
		return "", models.ErrRoadmapAIEmptyResult
	}
	return advice, nil
}

var quizSchema = llm.Schema{
	Name:        "card_question",
	Description: "One question about the card.",
	Properties: map[string]any{
		"question": map[string]any{
			"type":        "string",
			"description": fmt.Sprintf("At most %d characters. One open question, answerable in a few sentences.", maxQuizQuestionLen),
		},
	},
	Required: []string{"question"},
}

func (srv *roadmapAIService) QuizCard(ctx context.Context, userID, cardID int64, lang string) (models.RoadmapQuiz, error) {
	client, ok := srv.llm.For(llm.TaskQuiz)
	if !ok {
		return models.RoadmapQuiz{}, models.ErrRoadmapAIDisabled
	}

	card, err := srv.findCard(ctx, userID, cardID)
	if err != nil {
		return models.RoadmapQuiz{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, quizTimeout)
	defer cancel()

	var out struct {
		Question string `json:"question"`
	}
	if _, err := client.CompleteJSON(ctx, llm.Request{
		System: quizSystemPrompt(lang),
		Prompt: fmt.Sprintf("Card (%s): %s", card.Kind, card.Text),
		Effort: llm.EffortLow,
	}, quizSchema, &out); err != nil {
		return models.RoadmapQuiz{}, fmt.Errorf("quiz card: %w", err)
	}

	question := clampRunes(strings.TrimSpace(out.Question), maxQuizQuestionLen)
	if question == "" {
		return models.RoadmapQuiz{}, models.ErrRoadmapAIEmptyResult
	}
	return models.RoadmapQuiz{CardID: card.ID, CardText: card.Text, Question: question}, nil
}

var gradeSchema = llm.Schema{
	Name:        "answer_grade",
	Description: "The judgement of the answer.",
	Properties: map[string]any{
		"verdict": map[string]any{
			"type":        "string",
			"enum":        []string{"correct", "partial", "wrong"},
			"description": "correct if the answer covers the point, partial if it gets the idea but misses something, wrong otherwise.",
		},
		"feedback": map[string]any{
			"type":        "string",
			"description": fmt.Sprintf("At most %d characters. What was missing or wrong, or what to read next.", maxQuizFeedbackLen),
		},
	},
	Required: []string{"verdict", "feedback"},
}

func (srv *roadmapAIService) GradeQuizAnswer(ctx context.Context, quiz models.RoadmapQuiz, answer, lang string) (models.RoadmapQuizGrade, error) {
	client, ok := srv.llm.For(llm.TaskQuiz)
	if !ok {
		return models.RoadmapQuizGrade{}, models.ErrRoadmapAIDisabled
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return models.RoadmapQuizGrade{}, models.ErrRoadmapAIEmptyResult
	}

	ctx, cancel := context.WithTimeout(ctx, quizTimeout)
	defer cancel()

	var out struct {
		Verdict  string `json:"verdict"`
		Feedback string `json:"feedback"`
	}
	if _, err := client.CompleteJSON(ctx, llm.Request{
		System: gradeSystemPrompt(lang),
		Prompt: fmt.Sprintf("Card: %s\nQuestion: %s\nAnswer: %s",
			quiz.CardText, quiz.Question, clampRunes(answer, maxQuizAnswerLen)),
		Effort: llm.EffortMedium,
	}, gradeSchema, &out); err != nil {
		return models.RoadmapQuizGrade{}, fmt.Errorf("grade answer: %w", err)
	}

	verdict := models.RoadmapQuizVerdict(strings.ToLower(strings.TrimSpace(out.Verdict)))
	switch verdict {
	case models.RoadmapQuizCorrect, models.RoadmapQuizPartial, models.RoadmapQuizWrong:
	default:
		// Treated as the middle verdict rather than dropped: the feedback is
		// still worth showing, and calling an unknown verdict "wrong" would
		// misjudge the user.
		verdict = models.RoadmapQuizPartial
	}

	return models.RoadmapQuizGrade{
		Verdict:  verdict,
		Feedback: clampRunes(strings.TrimSpace(out.Feedback), maxQuizFeedbackLen),
	}, nil
}

// findCard resolves a card by id within the user's own plan. The repository
// has no "get one card" query — ListCards is per technology — so this walks
// the user's technologies. Bounded by MaxRoadmapGoalsPerUser ×
// MaxRoadmapsPerGoal, and only runs on an explicit quiz tap.
func (srv *roadmapAIService) findCard(ctx context.Context, userID, cardID int64) (models.RoadmapCardItem, error) {
	techs, err := srv.repo.ListRoadmapsAnyGoal(ctx, userID, false)
	if err != nil {
		return models.RoadmapCardItem{}, err
	}
	for _, tech := range techs {
		cards, err := srv.repo.ListCards(ctx, userID, tech.ID)
		if err != nil {
			return models.RoadmapCardItem{}, err
		}
		for _, card := range cards {
			if card.ID == cardID {
				return card, nil
			}
		}
	}
	return models.RoadmapCardItem{}, models.ErrRoadmapCardNotFound
}

func clampRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return strings.TrimSpace(string(r[:max]))
}
