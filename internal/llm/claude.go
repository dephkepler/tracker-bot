package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const ProviderClaude = "claude"

// DefaultClaudeModel is what an unset LLM_CLAUDE_DEFAULT_MODEL resolves to.
// Opus is the capable end of the range and priced accordingly; the per-task
// env overrides (LLM_<TASK>_PROVIDER / LLM_<TASK>_MODEL) exist so cheap tasks
// like card tagging can be pointed at Haiku or Groq without touching code.
const DefaultClaudeModel = "claude-opus-5"

type claudeClient struct {
	api   anthropic.Client
	model string
}

// NewClaude builds a Claude client. Retries, timeouts and connection reuse
// are the SDK's job here — unlike the Groq path, which rolls its own.
func NewClaude(key, model string) Client {
	return &claudeClient{
		api:   anthropic.NewClient(option.WithAPIKey(key)),
		model: model,
	}
}

func (c *claudeClient) Provider() string { return ProviderClaude }
func (c *claudeClient) Model() string    { return c.model }

func (c *claudeClient) Complete(ctx context.Context, req Request) (string, Usage, error) {
	return c.call(ctx, req, nil)
}

func (c *claudeClient) CompleteJSON(ctx context.Context, req Request, schema Schema, out any) (Usage, error) {
	// Structured outputs constrain the reply server-side, so unlike the Groq
	// path there is no need to describe the schema in the prompt and no
	// realistic chance of unparseable JSON coming back.
	body, u, err := c.call(ctx, req, jsonSchemaOf(schema))
	if err != nil {
		return u, err
	}
	if err := json.Unmarshal([]byte(body), out); err != nil {
		return u, fmt.Errorf("claude returned unparseable JSON (%d bytes, finish %q): %w",
			len(body), u.FinishReason, err)
	}
	return u, nil
}

func (c *claudeClient) call(ctx context.Context, req Request, format map[string]any) (string, Usage, error) {
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: int64(maxTokensOr(req.MaxTokens)),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
		// Adaptive thinking: the model decides how much reasoning a given
		// request needs. Effort below is the cost dial, not this.
		Thinking: anthropic.ThinkingConfigParamUnion{
			OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{},
		},
	}
	if req.System != "" {
		// No cache_control here on purpose: these system prompts are a few
		// hundred tokens, well under the ~1024-token minimum cacheable
		// prefix, so a breakpoint would be inert.
		params.System = []anthropic.TextBlockParam{{Text: req.System}}
	}
	if req.Effort != "" {
		params.OutputConfig.Effort = anthropic.OutputConfigEffort(req.Effort)
	}
	if format != nil {
		params.OutputConfig.Format = anthropic.JSONOutputFormatParam{Schema: format}
	}

	start := time.Now()
	resp, err := c.api.Messages.New(ctx, params)
	if err != nil {
		return "", Usage{}, fmt.Errorf("claude request failed (model %s): %w", c.model, err)
	}

	u := Usage{
		InputTokens:  int(resp.Usage.InputTokens),
		OutputTokens: int(resp.Usage.OutputTokens),
		FinishReason: string(resp.StopReason),
	}

	// A refusal arrives as HTTP 200 with empty content, so it has to be
	// checked before reading the blocks or it looks like an empty reply.
	if resp.StopReason == anthropic.StopReasonRefusal {
		return "", u, fmt.Errorf("claude declined to answer (category %q): %s",
			resp.StopDetails.Category, resp.StopDetails.Explanation)
	}

	var text strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			text.WriteString(tb.Text)
		}
	}
	if text.Len() == 0 {
		return "", u, fmt.Errorf("claude returned no text (%d blocks, finish %q)",
			len(resp.Content), resp.StopReason)
	}

	logCall(ProviderClaude, c.model, u, time.Since(start))
	return text.String(), u, nil
}

// jsonSchemaOf renders a Schema as the JSON Schema object the structured
// outputs parameter expects. Closed object: the model cannot add keys.
func jsonSchemaOf(schema Schema) map[string]any {
	return map[string]any{
		"type":                 "object",
		"title":                schema.Name,
		"description":          schema.Description,
		"properties":           schema.Properties,
		"required":             schema.Required,
		"additionalProperties": false,
	}
}
