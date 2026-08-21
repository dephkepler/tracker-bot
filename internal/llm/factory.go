package llm

import (
	"fmt"
	"sort"
	"strings"

	"tracker-bot/internal/config"
)

// Task is one AI-backed job. Each resolves its own provider and model so the
// cheap ones need not run on the expensive model.
type Task string

const (
	// TaskPlan generates a card checklist for a technology under a goal.
	// The most demanding of the four: long output, needs real knowledge of
	// what learning a technology actually involves.
	TaskPlan Task = "plan"
	// TaskTagging assigns kind and difficulty to pasted card lines.
	TaskTagging Task = "tagging"
	// TaskDigest writes the "what to take next and why" note on a push.
	TaskDigest Task = "digest"
	// TaskQuiz asks a question about a card and grades the answer.
	TaskQuiz Task = "quiz"
)

// AllTasks is the set the registry builds clients for.
var AllTasks = []Task{TaskPlan, TaskTagging, TaskDigest, TaskQuiz}

// Registry holds one ready client per task. Built once at startup: each
// client owns an HTTP client, so building per request would leak connection
// pools.
type Registry struct {
	clients map[Task]Client
}

// NewRegistry resolves a client for every task. A configuration with no
// provider at all yields an empty registry and no error — the AI features
// are simply off. A configuration that names a provider but cannot be
// completed (unknown name, missing key or model) is a real error and stops
// startup, because silently disabling a feature the operator asked for is
// worse than refusing to boot.
func NewRegistry(cfg config.LLM) (*Registry, error) {
	reg := &Registry{clients: make(map[Task]Client, len(AllTasks))}
	if !cfg.AIEnabled() {
		return reg, nil
	}

	// Clients are shared between tasks that resolve to the same
	// provider+model, so the common single-provider setup builds one client
	// rather than four.
	shared := make(map[string]Client)
	for _, task := range AllTasks {
		provider, model := defaultsFor(cfg, task)
		client, err := build(cfg, provider, model, shared)
		if err != nil {
			return nil, fmt.Errorf("llm: task %s: %w", task, err)
		}
		reg.clients[task] = client
	}
	return reg, nil
}

func build(cfg config.LLM, provider, model string, shared map[string]Client) (Client, error) {
	name := normalizeProvider(provider)
	if name == "" {
		return nil, ErrNoProvider
	}
	key := keyFor(cfg, name)
	if key == "" {
		return nil, fmt.Errorf("api key is empty for provider %q", provider)
	}
	if model == "" {
		return nil, fmt.Errorf("model is empty for provider %q", provider)
	}

	cacheKey := name + "|" + model
	if c, ok := shared[cacheKey]; ok {
		return c, nil
	}

	var client Client
	switch name {
	case ProviderClaude:
		client = NewClaude(key, model)
	case ProviderGroq:
		client = NewGroq(key, model)
	default:
		return nil, fmt.Errorf("unknown provider %q (supported: %s)", provider, supportedProviders())
	}
	shared[cacheKey] = client
	return client, nil
}

// For returns the client for a task. The false return is the "AI is off"
// path every caller has to handle, not an error case.
func (r *Registry) For(task Task) (Client, bool) {
	if r == nil {
		return nil, false
	}
	c, ok := r.clients[task]
	return c, ok
}

// Enabled reports whether any AI feature is available.
func (r *Registry) Enabled() bool {
	return r != nil && len(r.clients) > 0
}

func defaultsFor(cfg config.LLM, task Task) (provider, model string) {
	switch task {
	case TaskPlan:
		return resolveTask(cfg, cfg.PlanProvider, cfg.PlanModel)
	case TaskTagging:
		return resolveTask(cfg, cfg.TaggingProvider, cfg.TaggingModel)
	case TaskDigest:
		return resolveTask(cfg, cfg.DigestProvider, cfg.DigestModel)
	case TaskQuiz:
		return resolveTask(cfg, cfg.QuizProvider, cfg.QuizModel)
	}
	return defaultProvider(cfg), cfg.Model
}

// resolveTask layers the task override over the global default. The subtle
// case is a task that switches provider without naming a model: it must get
// the new provider's default model, not the other provider's model.
func resolveTask(cfg config.LLM, taskProvider, taskModel string) (provider, model string) {
	provider = firstNonEmpty(taskProvider, defaultProvider(cfg))
	if taskModel != "" {
		return provider, taskModel
	}
	if normalizeProvider(provider) != normalizeProvider(defaultProvider(cfg)) {
		return provider, modelFor(cfg, provider)
	}
	return provider, firstNonEmpty(cfg.Model, modelFor(cfg, provider))
}

// defaultProvider falls back to whichever provider has a key when
// LLM_PROVIDER is unset, so a setup with only LLM_CLAUDE_API_KEY works
// without naming the provider twice.
func defaultProvider(cfg config.LLM) string {
	if cfg.Provider != "" {
		return cfg.Provider
	}
	if cfg.ClaudeAPIKey != "" {
		return ProviderClaude
	}
	if cfg.GroqAPIKey != "" {
		return ProviderGroq
	}
	return ""
}

func modelFor(cfg config.LLM, provider string) string {
	switch normalizeProvider(provider) {
	case ProviderClaude:
		return firstNonEmpty(cfg.ClaudeDefaultModel, DefaultClaudeModel)
	case ProviderGroq:
		return cfg.GroqDefaultModel
	}
	return ""
}

func keyFor(cfg config.LLM, provider string) string {
	switch normalizeProvider(provider) {
	case ProviderClaude:
		if cfg.ClaudeAPIKey != "" {
			return cfg.ClaudeAPIKey
		}
	case ProviderGroq:
		if cfg.GroqAPIKey != "" {
			return cfg.GroqAPIKey
		}
	}
	// LLM_API_KEY is the single-provider shorthand: it only applies to the
	// provider it was set alongside.
	if normalizeProvider(provider) == normalizeProvider(defaultProvider(cfg)) {
		return cfg.APIKey
	}
	return ""
}

// normalizeProvider accepts "anthropic" as a synonym for "claude", since
// that is what the env var reads like to anyone coming from the API docs.
func normalizeProvider(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "anthropic" {
		return ProviderClaude
	}
	return s
}

func supportedProviders() string {
	names := []string{ProviderClaude, ProviderGroq}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
