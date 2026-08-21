package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Groq is the OpenAI-compatible chat-completions provider. SEO kept a
// codec/transport split here so two providers could share one HTTP client;
// with Claude now on its own SDK, Groq is the only user of this path, so it
// is written out flat instead.
const (
	groqAPIURL   = "https://api.groq.com/openai/v1/chat/completions"
	ProviderGroq = "groq"

	// requestTimeout is per attempt, not per call: with retries the wall
	// clock can reach several times this. Generous because a full plan
	// generation is a long single reply.
	groqRequestTimeout = 120 * time.Second

	// maxResponseBytes bounds what a misbehaving endpoint can make us read
	// into memory.
	maxResponseBytes = 1 << 20
	// maxErrorBodyBytes bounds what we copy into an error message and the log.
	maxErrorBodyBytes = 4 << 10
)

type groqClient struct {
	key   string
	model string
	http  *http.Client
}

// NewGroq builds a Groq client. Both key and model must be non-empty; the
// factory checks that.
func NewGroq(key, model string) Client {
	return &groqClient{
		key:   key,
		model: model,
		http:  &http.Client{Timeout: groqRequestTimeout},
	}
}

func (c *groqClient) Provider() string { return ProviderGroq }
func (c *groqClient) Model() string    { return c.model }

// providerError carries the status so the retry policy can classify it.
type providerError struct {
	provider   string
	status     int
	body       string
	retryAfter time.Duration
	// cause preserves the transport error so callers can errors.Is/As into
	// net.Timeout, context.Canceled and friends.
	cause error
}

func (e *providerError) Error() string {
	if e.status == 0 {
		return e.provider + " transport error: " + e.body
	}
	return fmt.Sprintf("%s returned %d: %s", e.provider, e.status, e.body)
}

func (e *providerError) Unwrap() error { return e.cause }

func (e *providerError) HTTPStatus() int { return e.status }

func (e *providerError) RetryAfter() (time.Duration, bool) {
	return e.retryAfter, e.retryAfter > 0
}

func truncateBody(body []byte) string {
	if len(body) <= maxErrorBodyBytes {
		return string(body)
	}
	return string(body[:maxErrorBodyBytes]) + "...(truncated)"
}

func (c *groqClient) Complete(ctx context.Context, req Request) (string, Usage, error) {
	return c.call(ctx, req, nil)
}

func (c *groqClient) CompleteJSON(ctx context.Context, req Request, schema Schema, out any) (Usage, error) {
	// Groq has no strict-schema mode across all models, so the schema is
	// stated in the prompt and the reply is forced to be a JSON object.
	// Unlike Claude's strict tool use this is advisory, hence the unmarshal
	// error below is a real possibility rather than a should-never-happen.
	req.System = joinSystem(req.System, schemaInstruction(schema))

	body, u, err := c.call(ctx, req, map[string]any{"type": "json_object"})
	if err != nil {
		return u, err
	}
	if err := json.Unmarshal([]byte(body), out); err != nil {
		return u, fmt.Errorf("groq returned unparseable JSON (%d bytes, finish %q): %w",
			len(body), u.FinishReason, err)
	}
	return u, nil
}

func (c *groqClient) call(ctx context.Context, req Request, responseFormat map[string]any) (string, Usage, error) {
	var (
		content string
		u       Usage
	)

	start := time.Now()
	err := doWithRetry(ctx, ProviderGroq, func() error {
		payload := map[string]any{
			"model":      c.model,
			"max_tokens": maxTokensOr(req.MaxTokens),
			"messages":   groqMessages(req),
		}
		if responseFormat != nil {
			payload["response_format"] = responseFormat
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, groqAPIURL, bytes.NewReader(raw))
		if err != nil {
			return err
		}
		httpReq.Header.Set("Authorization", "Bearer "+c.key)
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(httpReq)
		if err != nil {
			return &providerError{provider: ProviderGroq, body: err.Error(), cause: err}
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
		if err != nil {
			return &providerError{provider: ProviderGroq, status: resp.StatusCode, body: err.Error(), cause: err}
		}
		if resp.StatusCode != http.StatusOK {
			return &providerError{
				provider:   ProviderGroq,
				status:     resp.StatusCode,
				body:       truncateBody(respBody),
				retryAfter: parseRetryAfter(resp.Header),
			}
		}

		var parsed struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			return fmt.Errorf("groq: decode response: %w", err)
		}
		if len(parsed.Choices) == 0 {
			return fmt.Errorf("groq returned no choices (%d bytes)", len(respBody))
		}

		content = parsed.Choices[0].Message.Content
		u = Usage{
			InputTokens:  parsed.Usage.PromptTokens,
			OutputTokens: parsed.Usage.CompletionTokens,
			FinishReason: parsed.Choices[0].FinishReason,
		}
		return nil
	})
	if err != nil {
		return "", Usage{}, fmt.Errorf("groq request failed (model %s): %w", c.model, err)
	}

	logCall(ProviderGroq, c.model, u, time.Since(start))
	return content, u, nil
}

func groqMessages(req Request) []map[string]string {
	msgs := make([]map[string]string, 0, 2)
	if req.System != "" {
		msgs = append(msgs, map[string]string{"role": "system", "content": req.System})
	}
	return append(msgs, map[string]string{"role": "user", "content": req.Prompt})
}

// schemaInstruction renders a Schema as prompt text, for providers without a
// strict-schema mode.
func schemaInstruction(schema Schema) string {
	props, err := json.Marshal(schema.Properties)
	if err != nil {
		props = []byte("{}")
	}
	required, err := json.Marshal(schema.Required)
	if err != nil {
		required = []byte("[]")
	}
	return fmt.Sprintf(
		"Reply with a single JSON object and nothing else. %s\n"+
			"Its properties: %s\nRequired keys: %s\nAdd no other keys.",
		schema.Description, props, required,
	)
}

func joinSystem(parts ...string) string {
	var buf bytes.Buffer
	for _, p := range parts {
		if p == "" {
			continue
		}
		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(p)
	}
	return buf.String()
}

func logCall(provider, model string, u Usage, latency time.Duration) {
	ev := log.Info()
	if u.Truncated() {
		// A truncated reply is not an error the provider reports — the JSON
		// simply arrives half-written. Worth a warning so a too-low
		// MaxTokens is visible in the logs rather than only as a parse error.
		ev = log.Warn().Str("finish_reason", u.FinishReason)
	}
	ev.Str("provider", provider).
		Str("model", model).
		Dur("latency", latency).
		Int("input_tokens", u.InputTokens).
		Int("output_tokens", u.OutputTokens).
		Msg("llm call done")
}
