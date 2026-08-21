package web

import (
	"context"
	"errors"
	"net/http"

	"tracker-bot/internal/models"
	"tracker-bot/internal/utils/webctx"
	"tracker-bot/internal/web/apidto"
	"tracker-bot/pkg/apptime"
)

// maxBodyBytes bounds a request body. Every write here is a handful of fields;
// anything larger is a mistake or an attempt.
const maxBodyBytes = 16 << 10

// aiErrorFor maps a roadmap-AI failure onto what the client should show. The
// underlying error is logged, never returned: it can carry a provider's
// response.
func aiErrorFor(err error) (code, message string) {
	switch {
	case errors.Is(err, models.ErrRoadmapAIDisabled):
		return codeAIDisabled, "ИИ не подключён"
	case errors.Is(err, models.ErrRoadmapAIEmptyResult):
		return codeAIEmpty, "ИИ не вернул ничего пригодного, попробуй ещё раз"
	case errors.Is(err, models.ErrRoadmapCardNotFound):
		return codeNotFound, "карточка не найдена"
	case errors.Is(err, models.ErrRoadmapNotFound):
		return codeNotFound, "технология не найдена"
	default:
		return codeAIFailed, "ИИ не ответил, попробуй позже"
	}
}

// handleRoadmap returns the whole plan: goals, the technologies under them and
// their cards.
//
// One request rather than a tree of them. The caps make the shape small — three
// goals, five technologies each — and a Mini App pays for every round trip on a
// cold mobile connection. The cards carry their own done_at, so a client can
// draw the completion history without another endpoint; nothing in the bot has
// ever shown that series.
func (s *Server) handleRoadmap(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	goals, err := s.roadmapsvc.ListGoals(r.Context(), user.DBUserID)
	if err != nil {
		s.fail(w, r, err, "list goals failed")
		return
	}

	out := apidto.RoadmapResponse{
		Goals: make([]apidto.RoadmapGoal, 0, len(goals)),
		Meta:  apidto.NewMeta(user.Location, apptime.NowIn(user.Location)),
	}

	for _, goal := range goals {
		techs, err := s.roadmapsvc.ListRoadmaps(r.Context(), user.DBUserID, goal.ID)
		if err != nil {
			s.fail(w, r, err, "list technologies failed")
			return
		}
		built, err := s.buildTechnologies(r, user, techs)
		if err != nil {
			s.fail(w, r, err, "list cards failed")
			return
		}
		out.Goals = append(out.Goals, apidto.FromRoadmapGoal(goal, built))
	}

	// Technologies attached to no goal: v1 leftovers, or ones whose goal was
	// deleted. They are still work in progress, so hiding them would lose them.
	orphans, err := s.roadmapsvc.ListOrphanRoadmaps(r.Context(), user.DBUserID)
	if err != nil {
		s.fail(w, r, err, "list orphan technologies failed")
		return
	}
	out.OrphanTechnologies, err = s.buildTechnologies(r, user, orphans)
	if err != nil {
		s.fail(w, r, err, "list cards failed")
		return
	}

	writeJSON(w, r, http.StatusOK, out)
}

func (s *Server) buildTechnologies(r *http.Request, user webctx.User, items []models.RoadmapItem) ([]apidto.RoadmapTechnology, error) {
	out := make([]apidto.RoadmapTechnology, 0, len(items))
	for _, item := range items {
		cards, err := s.roadmapsvc.ListCards(r.Context(), user.DBUserID, item.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, apidto.FromRoadmapTechnology(item, cards))
	}
	return out, nil
}

// handleRoadmapCardDone sets a card's state.
//
// A PUT carrying the state wanted, not a POST that flips it. The bot toggles
// because a tap is a toggle, but an HTTP request can be retried by the client,
// a proxy or an impatient thumb, and a flip would undo itself. Saying "done:
// true" twice lands in the same place.
func (s *Server) handleRoadmapCardDone(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	cardID, perr := pathID(r, "id")
	if perr != nil {
		s.failRequest(w, r, perr, "card id")
		return
	}

	var body struct {
		Done *bool `json:"done"`
	}
	if err := decodeBody(r, &body); err != nil {
		s.failRequest(w, r, err, "card body")
		return
	}
	if body.Done == nil {
		s.failRequest(w, r, badParam(`body must be {"done": true} or {"done": false}`), "card body")
		return
	}

	roadmapID, err := s.roadmapsvc.SetCardDone(r.Context(), user.DBUserID, cardID, *body.Done)
	if err != nil {
		if errors.Is(err, models.ErrRoadmapCardNotFound) {
			// Also the answer for someone else's card: whether it exists is
			// not something to confirm.
			writeErr(w, r, http.StatusNotFound, codeNotFound, "карточка не найдена")
			return
		}
		s.fail(w, r, err, "set card done failed")
		return
	}

	writeJSON(w, r, http.StatusOK, apidto.CardStateResponse{
		CardID:    cardID,
		RoadmapID: roadmapID,
		Done:      *body.Done,
	})
}

// handleRoadmapPlan starts plan generation for one technology.
// startAIJob hands the work to the job store and answers immediately with the
// job to poll. Every AI endpoint goes through here so none of them can
// accidentally be synchronous — a minute-long request is one backgrounded app
// away from being lost.
func (s *Server) startAIJob(w http.ResponseWriter, r *http.Request, kind string, fn func(context.Context) (any, error)) {
	user := webctx.MustFrom(r.Context())

	entry, err := s.jobs.start(kind, user.DBUserID, fn)
	if err != nil {
		s.fail(w, r, err, "could not start an ai job")
		return
	}
	// 202: accepted and running, with the id to poll.
	writeJSON(w, r, http.StatusAccepted, apidto.JobResponse{
		ID: entry.id, Kind: entry.kind, Status: string(entry.status),
	})
}

func (s *Server) handleRoadmapPlan(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	techID, perr := pathID(r, "id")
	if perr != nil {
		s.failRequest(w, r, perr, "technology id")
		return
	}

	s.startAIJob(w, r, "plan", func(ctx context.Context) (any, error) {
		added, rejected, err := s.roadmapaisvc.GeneratePlan(ctx, user.DBUserID, techID, user.Language)
		if err != nil {
			return nil, err
		}
		return apidto.PlanResult{Added: added, Rejected: rejected}, nil
	})
}

// handleRoadmapQuiz starts generating a question about a card.
func (s *Server) handleRoadmapQuiz(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	cardID, perr := pathID(r, "id")
	if perr != nil {
		s.failRequest(w, r, perr, "card id")
		return
	}

	s.startAIJob(w, r, "quiz", func(ctx context.Context) (any, error) {
		quiz, err := s.roadmapaisvc.QuizCard(ctx, user.DBUserID, cardID, user.Language)
		if err != nil {
			return nil, err
		}
		return apidto.QuizResult{CardID: quiz.CardID, CardText: quiz.CardText, Question: quiz.Question}, nil
	})
}

// handleRoadmapQuizGrade starts grading an answer.
//
// The question comes back from the client rather than being held server-side.
// It is not a secret and it is not a credential — it is the user's own quiz —
// and keeping it client-side means no session state, so a graded answer works
// after a reload, from a second device, or when the process has restarted in
// between.
func (s *Server) handleRoadmapQuizGrade(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	cardID, perr := pathID(r, "id")
	if perr != nil {
		s.failRequest(w, r, perr, "card id")
		return
	}

	var body struct {
		CardText string `json:"card_text"`
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}
	if err := decodeBody(r, &body); err != nil {
		s.failRequest(w, r, err, "grade body")
		return
	}
	if body.Question == "" || body.Answer == "" {
		s.failRequest(w, r, badParam("question and answer are both required"), "grade body")
		return
	}

	quiz := models.RoadmapQuiz{CardID: cardID, CardText: body.CardText, Question: body.Question}
	s.startAIJob(w, r, "grade", func(ctx context.Context) (any, error) {
		grade, err := s.roadmapaisvc.GradeQuizAnswer(ctx, quiz, body.Answer, user.Language)
		if err != nil {
			return nil, err
		}
		return apidto.GradeResult{Verdict: string(grade.Verdict), Feedback: grade.Feedback}, nil
	})
}

// handleAIJob reports a job's state.
func (s *Server) handleAIJob(w http.ResponseWriter, r *http.Request) {
	user := webctx.MustFrom(r.Context())

	entry, ok := s.jobs.get(r.PathValue("id"), user.DBUserID)
	if !ok {
		// Unknown, expired, or someone else's — all the same answer, since
		// which one it is would itself be information.
		writeErr(w, r, http.StatusNotFound, codeNotFound, "задание не найдено")
		return
	}

	out := apidto.JobResponse{ID: entry.id, Kind: entry.kind, Status: string(entry.status)}
	switch entry.status {
	case jobDone:
		out.Result = entry.result
	case jobFailed:
		out.ErrorCode = entry.errCode
		out.ErrorMessage = entry.errMessage
	}
	writeJSON(w, r, http.StatusOK, out)
}
