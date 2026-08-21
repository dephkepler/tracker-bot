package llm

import (
	"testing"

	"tracker-bot/internal/config"
)

func TestNewRegistryDisabledWithoutProvider(t *testing.T) {
	reg, err := NewRegistry(config.LLM{})
	if err != nil {
		t.Fatalf("empty config must not error, got %v", err)
	}
	if reg.Enabled() {
		t.Fatal("registry must be disabled when nothing is configured")
	}
	if _, ok := reg.For(TaskPlan); ok {
		t.Fatal("For must report false when AI is off")
	}
}

func TestNewRegistryProviderWithoutKeyFails(t *testing.T) {
	// Naming a provider and forgetting its key is an operator mistake, and
	// booting with the feature quietly off would hide it.
	_, err := NewRegistry(config.LLM{Provider: "claude", ClaudeDefaultModel: "claude-opus-5"})
	if err == nil {
		t.Fatal("expected an error when the provider has no key")
	}
}

func TestNewRegistryUnknownProviderFails(t *testing.T) {
	_, err := NewRegistry(config.LLM{Provider: "gpt", APIKey: "k", Model: "m"})
	if err == nil {
		t.Fatal("expected an error for an unknown provider")
	}
}

func TestRegistrySharesClientAcrossTasks(t *testing.T) {
	reg, err := NewRegistry(config.LLM{
		Provider:           "claude",
		ClaudeAPIKey:       "k",
		ClaudeDefaultModel: "claude-opus-5",
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	plan, _ := reg.For(TaskPlan)
	quiz, _ := reg.For(TaskQuiz)
	if plan != quiz {
		t.Fatal("tasks resolving to the same provider+model must share one client")
	}
}

func TestDefaultProviderInferredFromKey(t *testing.T) {
	// Only a Claude key set, LLM_PROVIDER unset: everything must land on
	// Claude without the provider being named twice.
	reg, err := NewRegistry(config.LLM{ClaudeAPIKey: "k", ClaudeDefaultModel: "claude-opus-5"})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	c, ok := reg.For(TaskDigest)
	if !ok {
		t.Fatal("digest client missing")
	}
	if c.Provider() != ProviderClaude {
		t.Fatalf("provider = %q, want %q", c.Provider(), ProviderClaude)
	}
}

func TestTaskOverrideSwitchesProviderAndPicksItsDefaultModel(t *testing.T) {
	// The regression this guards: a task that switches provider without
	// naming a model must not inherit the other provider's model.
	reg, err := NewRegistry(config.LLM{
		Provider:           "claude",
		Model:              "claude-opus-5",
		ClaudeAPIKey:       "ck",
		GroqAPIKey:         "gk",
		ClaudeDefaultModel: "claude-opus-5",
		GroqDefaultModel:   "llama-3.3-70b-versatile",
		TaggingProvider:    "groq",
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	tagging, _ := reg.For(TaskTagging)
	if tagging.Provider() != ProviderGroq {
		t.Fatalf("tagging provider = %q, want groq", tagging.Provider())
	}
	if tagging.Model() != "llama-3.3-70b-versatile" {
		t.Fatalf("tagging model = %q, want the groq default", tagging.Model())
	}

	plan, _ := reg.For(TaskPlan)
	if plan.Provider() != ProviderClaude || plan.Model() != "claude-opus-5" {
		t.Fatalf("plan = %s/%s, want claude/claude-opus-5", plan.Provider(), plan.Model())
	}
}

func TestAnthropicIsASynonymForClaude(t *testing.T) {
	reg, err := NewRegistry(config.LLM{
		Provider:           "anthropic",
		ClaudeAPIKey:       "k",
		ClaudeDefaultModel: "claude-opus-5",
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	c, _ := reg.For(TaskPlan)
	if c.Provider() != ProviderClaude {
		t.Fatalf("provider = %q, want claude", c.Provider())
	}
}

func TestTaskModelOverrideKeepsProvider(t *testing.T) {
	reg, err := NewRegistry(config.LLM{
		Provider:           "claude",
		ClaudeAPIKey:       "k",
		ClaudeDefaultModel: "claude-opus-5",
		TaggingModel:       "claude-haiku-4-5",
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	tagging, _ := reg.For(TaskTagging)
	if tagging.Provider() != ProviderClaude || tagging.Model() != "claude-haiku-4-5" {
		t.Fatalf("tagging = %s/%s, want claude/claude-haiku-4-5", tagging.Provider(), tagging.Model())
	}
}
