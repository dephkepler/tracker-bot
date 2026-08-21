package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tracker-bot/internal/models"
)

func newRoadmapServer(t *testing.T) (*Server, *fakeRoadmapSvc, *fakeRoadmapAISvc) {
	t.Helper()
	deps := testDeps()
	srv, err := NewServer(t.Context(), testWebConfig(), deps)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return srv, deps.Roadmap.(*fakeRoadmapSvc), deps.RoadmapAI.(*fakeRoadmapAISvc)
}

// send performs a request with a body through the full chain.
func send(t *testing.T, srv *Server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Authorization", authHeader(t, knownTgUserID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	return rec
}

type roadmapBody struct {
	Goals []struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		TotalCards   int    `json:"total_cards"`
		DoneCards    int    `json:"done_cards"`
		Technologies []struct {
			ID              int64  `json:"id"`
			Name            string `json:"name"`
			MasteryCriteria string `json:"mastery_criteria"`
			Cards           []struct {
				ID         int64   `json:"id"`
				Text       string  `json:"text"`
				Kind       string  `json:"kind"`
				Difficulty int     `json:"difficulty"`
				Done       bool    `json:"done"`
				DoneAt     *string `json:"done_at"`
			} `json:"cards"`
		} `json:"technologies"`
	} `json:"goals"`
	OrphanTechnologies []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"orphan_technologies"`
}

func TestRoadmapReturnsTheWholeTree(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)
	body := getJSON[roadmapBody](t, srv, "/api/v1/roadmap")

	if len(body.Goals) != 1 {
		t.Fatalf("got %d goals, want 1", len(body.Goals))
	}
	goal := body.Goals[0]
	if goal.Name != "выйти на мидла" || goal.DoneCards != 1 || goal.TotalCards != 3 {
		t.Fatalf("goal = %+v", goal)
	}
	if len(goal.Technologies) != 1 {
		t.Fatalf("got %d technologies, want 1", len(goal.Technologies))
	}
	tech := goal.Technologies[0]
	if tech.MasteryCriteria == "" {
		t.Fatal("the mastery criteria was dropped — it is what the AI plan is generated from")
	}
	if len(tech.Cards) != 3 {
		t.Fatalf("got %d cards, want 3", len(tech.Cards))
	}

	// done_at is the only completion timestamp in this domain and nothing in the
	// bot has ever shown it; sending it is what lets the dashboard draw a
	// history at all.
	if tech.Cards[0].DoneAt == nil {
		t.Fatal("a finished card carries no done_at")
	}
	if tech.Cards[1].DoneAt != nil {
		t.Fatal("a pending card carries a done_at")
	}
	if tech.Cards[2].Kind != "book" || tech.Cards[2].Difficulty != 3 {
		t.Fatalf("card kind/difficulty lost: %+v", tech.Cards[2])
	}

	// Technologies with no goal are still real work.
	if len(body.OrphanTechnologies) != 1 || body.OrphanTechnologies[0].Name != "Docker" {
		t.Fatalf("orphans = %+v", body.OrphanTechnologies)
	}
}

// The write says what it wants rather than flipping, so sending it twice lands
// in the same place — which is the whole reason it is a PUT with a body.
func TestCardDoneIsIdempotent(t *testing.T) {
	srv, roadmapsvc, _ := newRoadmapServer(t)

	for range 2 {
		rec := send(t, srv, http.MethodPut, "/api/v1/roadmap/cards/101/done", `{"done":true}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
		}
		var body struct {
			CardID    int64 `json:"card_id"`
			RoadmapID int64 `json:"roadmap_id"`
			Done      bool  `json:"done"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.CardID != 101 || !body.Done {
			t.Fatalf("body = %+v", body)
		}
	}

	// Both writes asked for the same state; a toggle would have undone itself.
	if len(roadmapsvc.setDone) != 2 {
		t.Fatalf("%d writes recorded, want 2", len(roadmapsvc.setDone))
	}
	for i, w := range roadmapsvc.setDone {
		if !w.done {
			t.Fatalf("write %d asked for done=false", i)
		}
	}
}

func TestCardDoneCanUntick(t *testing.T) {
	srv, roadmapsvc, _ := newRoadmapServer(t)

	rec := send(t, srv, http.MethodPut, "/api/v1/roadmap/cards/100/done", `{"done":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d (body %q)", rec.Code, rec.Body.String())
	}
	if len(roadmapsvc.setDone) != 1 || roadmapsvc.setDone[0].done {
		t.Fatalf("writes = %+v", roadmapsvc.setDone)
	}
}

func TestCardDoneRejectsBadRequests(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)

	cases := map[string]struct {
		path, body string
		wantStatus int
	}{
		"no body":       {"/api/v1/roadmap/cards/101/done", "", http.StatusBadRequest},
		"empty object":  {"/api/v1/roadmap/cards/101/done", `{}`, http.StatusBadRequest},
		"unknown field": {"/api/v1/roadmap/cards/101/done", `{"done_at":"now"}`, http.StatusBadRequest},
		"wrong type":    {"/api/v1/roadmap/cards/101/done", `{"done":"yes"}`, http.StatusBadRequest},
		"bad card id":   {"/api/v1/roadmap/cards/abc/done", `{"done":true}`, http.StatusBadRequest},
		"unknown card":  {"/api/v1/roadmap/cards/9999/done", `{"done":true}`, http.StatusNotFound},
		"negative id":   {"/api/v1/roadmap/cards/-1/done", `{"done":true}`, http.StatusBadRequest},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := send(t, srv, http.MethodPut, tc.path, tc.body)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCardDoneNeedsAuth(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/roadmap/cards/101/done", strings.NewReader(`{"done":true}`))
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 — a write must never be reachable unauthenticated", rec.Code)
	}
}

type jobBody struct {
	ID           string          `json:"id"`
	Kind         string          `json:"kind"`
	Status       string          `json:"status"`
	Result       json.RawMessage `json:"result"`
	ErrorCode    string          `json:"error_code"`
	ErrorMessage string          `json:"error_message"`
}

// awaitJob polls until the job settles, the way a client does.
func awaitJob(t *testing.T, srv *Server, id string) jobBody {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		body := getJSON[jobBody](t, srv, "/api/v1/ai/jobs/"+id)
		if body.Status != "pending" {
			return body
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s never settled", id)
	return jobBody{}
}

func startJob(t *testing.T, srv *Server, path string, body string) jobBody {
	t.Helper()
	rec := send(t, srv, http.MethodPost, path, body)
	// 202: the work is running, here is what to poll.
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202 (body %q)", rec.Code, rec.Body.String())
	}
	var out jobBody
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.ID == "" || out.Status != "pending" {
		t.Fatalf("job = %+v", out)
	}
	return out
}

func TestPlanRunsAsAJob(t *testing.T) {
	srv, _, aisvc := newRoadmapServer(t)

	started := startJob(t, srv, "/api/v1/roadmap/technologies/10/plan", "")
	if started.Kind != "plan" {
		t.Fatalf("kind = %q", started.Kind)
	}

	done := awaitJob(t, srv, started.ID)
	if done.Status != "done" {
		t.Fatalf("status = %q, error %q", done.Status, done.ErrorMessage)
	}
	var result struct {
		Added    int `json:"added"`
		Rejected int `json:"rejected"`
	}
	if err := json.Unmarshal(done.Result, &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	// Rejected drafts are reported, not hidden: "added 6" when it proposed 7 is
	// worth knowing.
	if result.Added != 6 || result.Rejected != 1 {
		t.Fatalf("result = %+v", result)
	}
	// The user's language has to reach the model, or the plan comes back in the
	// wrong one.
	if aisvc.gotLang != "ru" {
		t.Fatalf("language passed to the service = %q, want ru", aisvc.gotLang)
	}
}

func TestQuizAndGradeRunAsJobs(t *testing.T) {
	srv, _, aisvc := newRoadmapServer(t)

	quizJob := awaitJob(t, srv, startJob(t, srv, "/api/v1/roadmap/cards/101/quiz", "").ID)
	if quizJob.Status != "done" {
		t.Fatalf("quiz status = %q, error %q", quizJob.Status, quizJob.ErrorMessage)
	}
	var quiz struct {
		CardID   int64  `json:"card_id"`
		CardText string `json:"card_text"`
		Question string `json:"question"`
	}
	if err := json.Unmarshal(quizJob.Result, &quiz); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if quiz.CardID != 101 || quiz.Question == "" {
		t.Fatalf("quiz = %+v", quiz)
	}

	// The question comes back from the client, so grading holds no server-side
	// state and survives a reload or a restart between the two calls.
	gradeJob := awaitJob(t, srv, startJob(t, srv,
		"/api/v1/roadmap/cards/101/quiz/grade",
		`{"card_text":"Группы консьюмеров","question":"`+quiz.Question+`","answer":"ребалансировка переназначает партиции"}`,
	).ID)
	if gradeJob.Status != "done" {
		t.Fatalf("grade status = %q, error %q", gradeJob.Status, gradeJob.ErrorMessage)
	}
	var grade struct {
		Verdict  string `json:"verdict"`
		Feedback string `json:"feedback"`
	}
	if err := json.Unmarshal(gradeJob.Result, &grade); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if grade.Verdict != "partial" {
		t.Fatalf("verdict = %q", grade.Verdict)
	}
	if aisvc.gotAnswer == "" {
		t.Fatal("the answer never reached the service")
	}
}

func TestGradeRequiresQuestionAndAnswer(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)

	for _, body := range []string{`{"question":"","answer":"x"}`, `{"question":"q","answer":""}`, `{}`} {
		rec := send(t, srv, http.MethodPost, "/api/v1/roadmap/cards/101/quiz/grade", body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body %s: status = %d, want 400", body, rec.Code)
		}
	}
}

// A failure inside the job is reported through the job, not as a failed start:
// the request that launched it succeeded.
func TestJobCarriesTheFailure(t *testing.T) {
	srv, _, aisvc := newRoadmapServer(t)
	aisvc.enabled = false
	aisvc.err = models.ErrRoadmapAIDisabled

	done := awaitJob(t, srv, startJob(t, srv, "/api/v1/roadmap/technologies/10/plan", "").ID)
	if done.Status != "failed" {
		t.Fatalf("status = %q, want failed", done.Status)
	}
	if done.ErrorCode != codeAIDisabled {
		t.Fatalf("error_code = %q, want %q", done.ErrorCode, codeAIDisabled)
	}
	if done.ErrorMessage == "" {
		t.Fatal("no message for the user")
	}
}

func TestJobIsScopedToItsOwner(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)
	started := startJob(t, srv, "/api/v1/roadmap/technologies/10/plan", "")

	// A different Telegram user, genuinely authenticated, must not be able to
	// read someone else's generated plan by holding its id.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/jobs/"+started.ID, nil)
	req.Header.Set("Authorization", authHeader(t, 999999))
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)

	// That user has no row here, so the 404 arrives at the identity step; the
	// ownership check is exercised directly below.
	if rec.Code == http.StatusOK {
		t.Fatal("another user read the job")
	}

	if _, ok := srv.jobs.get(started.ID, 999); ok {
		t.Fatal("jobs.get returned a job to the wrong owner")
	}
	if _, ok := srv.jobs.get("does-not-exist", 7); ok {
		t.Fatal("jobs.get invented a job")
	}
}

func TestUnknownJob(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)

	rec := call(t, srv, "/api/v1/ai/jobs/deadbeef", authHeader(t, knownTgUserID))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestExpiredJobIsGone(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)
	started := startJob(t, srv, "/api/v1/roadmap/technologies/10/plan", "")
	awaitJob(t, srv, started.ID)

	// Wind the store's clock past the retention window.
	srv.jobs.now = func() time.Time { return time.Now().Add(2 * jobTTL) }

	if _, ok := srv.jobs.get(started.ID, 7); ok {
		t.Fatal("an expired job is still readable")
	}
}

// Many jobs at once, read while they finish. This is what caught the original
// bug: start used to hand back the same pointer the worker writes to, so the
// handler's read of status raced the worker's write. Run with -race.
func TestConcurrentJobsAreRaceFree(t *testing.T) {
	srv, _, _ := newRoadmapServer(t)

	const n = 12
	ids := make([]string, 0, n)
	for range n {
		ids = append(ids, startJob(t, srv, "/api/v1/roadmap/technologies/10/plan", "").ID)
	}
	for _, id := range ids {
		if done := awaitJob(t, srv, id); done.Status != "done" {
			t.Fatalf("job %s: status %q", id, done.Status)
		}
	}
}
