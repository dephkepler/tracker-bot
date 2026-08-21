// Package llm is the bot's LLM access layer: one Client interface over
// several providers, plus a factory that resolves which provider and model a
// given task should use.
//
// Ported from the multiagent-seo backend, with one deliberate difference —
// Claude goes through the official anthropic-sdk-go (retries, typed content
// blocks and strict tool use come for free), while OpenAI-compatible
// providers like Groq keep the hand-rolled transport below, since they have
// no official Go SDK.
package llm

import (
	"context"
	"errors"
)

// ErrNoProvider is what the factory returns when no provider is configured
// at all. Callers treat it as "the AI features are switched off" rather than
// as a failure: the bot must stay fully usable without an API key.
var ErrNoProvider = errors.New("llm: no provider configured")

// Usage is the token accounting of a single call, for logging and for the
// admin cost report.
type Usage struct {
	InputTokens  int
	OutputTokens int
	// FinishReason is the provider's raw stop reason ("end_turn",
	// "max_tokens", "tool_use", "stop", "length", ...). Compare with
	// Truncated rather than against literals.
	FinishReason string
}

// Truncated reports whether the reply was cut off by the token ceiling —
// worth a warning, because a truncated JSON body fails to parse and a
// truncated plan is silently short.
func (u Usage) Truncated() bool {
	return u.FinishReason == "max_tokens" || u.FinishReason == "length"
}

// Effort trades reply quality against cost and latency. Only Claude honours
// it; Groq ignores it.
type Effort string

const (
	EffortLow    Effort = "low"
	EffortMedium Effort = "medium"
	EffortHigh   Effort = "high"
)

// Request is one call. Prompt is required; everything else has a usable zero
// value.
type Request struct {
	// System is the role/rules preamble. Kept separate from Prompt because
	// Claude caches it as a prefix.
	System string
	Prompt string
	// MaxTokens caps the reply. Zero means DefaultMaxTokens.
	MaxTokens int
	Effort    Effort
}

// DefaultMaxTokens is deliberately generous: a generated learning plan for
// one technology runs long, and hitting the ceiling truncates the JSON into
// something unparseable.
const DefaultMaxTokens = 16000

// Schema constrains a reply to a JSON object. Properties is a JSON Schema
// "properties" map; Required lists the mandatory keys. The object is always
// closed (additionalProperties: false) so a provider cannot invent fields.
type Schema struct {
	// Name and Description are shown to the model and should read as an
	// instruction, e.g. "learning_plan" / "Return the generated plan".
	Name        string
	Description string
	Properties  map[string]any
	Required    []string
}

// Client is one configured provider+model pair. Implementations are
// safe for concurrent use.
type Client interface {
	// Complete returns the reply as free text.
	Complete(ctx context.Context, req Request) (string, Usage, error)
	// CompleteJSON constrains the reply to schema and unmarshals it into
	// out, which must be a non-nil pointer.
	CompleteJSON(ctx context.Context, req Request, schema Schema, out any) (Usage, error)
	// Provider and Model report what this client is, for logs and for the
	// "which model answered" line in admin output.
	Provider() string
	Model() string
}

func maxTokensOr(n int) int {
	if n <= 0 {
		return DefaultMaxTokens
	}
	return n
}
