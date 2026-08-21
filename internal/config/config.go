package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Telegram      Telegram
	PostreSQL     PgConfig
	LLM           LLM
	Web           Web
	AdminUsername string `env:"ADMIN_USERNAME" env-default:"alaamov"`
}
type PgConfig struct {
	Host    string `env:"HOST_DB"`
	Port    uint16 `env:"PORT_DB"`
	Name    string `env:"NAME_DB"`
	User    string `env:"USER_DB"`
	Pass    string `env:"PASSWORD_DB"`
	SSLMode string `env:"SSL_MODE" env-default:"disable"`
}

// LLM configures the AI features of the roadmap (plan generation, card
// tagging, digest advice, card quizzes). Provider empty means no AI: every
// feature degrades to its manual path and the bot works exactly as before,
// so a missing key is a normal state rather than a misconfiguration.
//
// Resolution per task: LLM_<TASK>_PROVIDER falls back to LLM_PROVIDER, and
// LLM_<TASK>_MODEL falls back to that provider's default model. The point is
// being able to run cheap tasks (tagging) on Groq or Haiku while the
// expensive one (plan generation) stays on the default model, without a
// rebuild.
type LLM struct {
	Provider string `env:"LLM_PROVIDER"`
	// APIKey is the key for Provider, for the single-provider setup. With
	// both providers in play, use the per-provider keys below instead.
	APIKey string `env:"LLM_API_KEY"`
	Model  string `env:"LLM_MODEL"`

	ClaudeAPIKey string `env:"LLM_CLAUDE_API_KEY"`
	GroqAPIKey   string `env:"LLM_GROQ_API_KEY"`

	ClaudeDefaultModel string `env:"LLM_CLAUDE_DEFAULT_MODEL" env-default:"claude-opus-5"`
	GroqDefaultModel   string `env:"LLM_GROQ_DEFAULT_MODEL" env-default:"llama-3.3-70b-versatile"`

	PlanProvider string `env:"LLM_PLAN_PROVIDER"`
	PlanModel    string `env:"LLM_PLAN_MODEL"`

	TaggingProvider string `env:"LLM_TAGGING_PROVIDER"`
	TaggingModel    string `env:"LLM_TAGGING_MODEL"`

	DigestProvider string `env:"LLM_DIGEST_PROVIDER"`
	DigestModel    string `env:"LLM_DIGEST_MODEL"`

	QuizProvider string `env:"LLM_QUIZ_PROVIDER"`
	QuizModel    string `env:"LLM_QUIZ_MODEL"`
}

// AIEnabled reports whether any provider is configured at all.
func (l LLM) AIEnabled() bool {
	return l.Provider != "" || l.ClaudeAPIKey != "" || l.GroqAPIKey != ""
}

// Web configures the read-only dashboard served as a Telegram Mini App.
//
// Enabled=false opens no listener at all and the bot behaves exactly as before,
// so every existing .env keeps working untouched — the same "unconfigured is a
// normal state" rule the LLM block follows.
type Web struct {
	Enabled bool   `env:"WEB_ENABLED" env-default:"false"`
	Addr    string `env:"WEB_ADDR" env-default:":8090"`

	// PublicOrigin is the https origin the Mini App is served on. Also the
	// guard that stops DevTgUserID from ever being live — see Validate.
	PublicOrigin string `env:"WEB_PUBLIC_ORIGIN"`

	// BotUsername and MiniAppShortName build the t.me deep link behind the
	// profile-menu button. An empty BotUsername hides that button entirely,
	// which is what keeps a dead button off the screen before the Mini App is
	// registered with BotFather.
	BotUsername      string `env:"WEB_BOT_USERNAME"`
	MiniAppShortName string `env:"WEB_MINIAPP_SHORT_NAME" env-default:"dashboard"`

	// InitDataMaxAge bounds how long after Telegram minted it a launch stays
	// valid. Generous on purpose: auth_date is stamped once when the Mini App
	// opens and never refreshes, so a short window signs out anyone who leaves
	// the dashboard on screen.
	InitDataMaxAge time.Duration `env:"WEB_INITDATA_MAX_AGE" env-default:"24h"`

	ReadTimeout   time.Duration `env:"WEB_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout  time.Duration `env:"WEB_WRITE_TIMEOUT" env-default:"15s"`
	IdleTimeout   time.Duration `env:"WEB_IDLE_TIMEOUT" env-default:"60s"`
	ShutdownGrace time.Duration `env:"WEB_SHUTDOWN_GRACE" env-default:"10s"`

	// MaxInflight bounds concurrent API requests. The dashboard shares one
	// pgxpool with the bot, so without a ceiling a runaway frontend would take
	// the bot's database connections with it.
	MaxInflight int `env:"WEB_MAX_INFLIGHT" env-default:"8"`

	// MenuButtonText labels the button beside the message field. A single
	// global default, so it cannot follow each user's language the way the rest
	// of the UI does; configurable so changing it needs no rebuild.
	MenuButtonText string `env:"WEB_MENU_BUTTON_TEXT" env-default:"Панель"`

	// DevTgUserID skips signature verification and pins every request to this
	// Telegram user. Local development only — it is the only way to open the
	// dashboard in an ordinary browser, and Validate refuses to let it start
	// alongside an https origin.
	DevTgUserID int64 `env:"WEB_DEV_TG_USER_ID"`
}

// Validate is called from ParseConfig only when the dashboard is enabled.
func (w Web) Validate() error {
	if strings.TrimSpace(w.Addr) == "" {
		return fmt.Errorf("web: WEB_ADDR is empty")
	}
	if _, _, err := net.SplitHostPort(w.Addr); err != nil {
		return fmt.Errorf("web: WEB_ADDR %q is not host:port: %w", w.Addr, err)
	}
	if w.MaxInflight <= 0 {
		return fmt.Errorf("web: WEB_MAX_INFLIGHT must be positive, got %d", w.MaxInflight)
	}
	if w.InitDataMaxAge <= 0 {
		return fmt.Errorf("web: WEB_INITDATA_MAX_AGE must be positive")
	}
	if w.PublicOrigin != "" {
		u, err := url.Parse(w.PublicOrigin)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("web: WEB_PUBLIC_ORIGIN %q is not an absolute URL", w.PublicOrigin)
		}
	}
	// The dev bypass hands every request the same identity. Reaching it from a
	// public origin would serve one person's data to anyone who found the URL,
	// so this refuses to boot rather than warning about it.
	if w.DevTgUserID != 0 && strings.HasPrefix(strings.ToLower(w.PublicOrigin), "https://") {
		return fmt.Errorf("web: WEB_DEV_TG_USER_ID must not be set with an https WEB_PUBLIC_ORIGIN")
	}
	return nil
}

type Telegram struct {
	TelegramToken    string `env:"TELEGRAM_TOKEN"`
	TelegramBotDebug bool   `env:"TELEGRAM_BOT_DEBUG"`
}

const (
	envConfigPath     = "CONFIG_PATH"
	defaultConfigPath = ".env"
)

func ParseConfig() (*Config, error) {
	var cfg Config

	// HOST_DB set means env-injected config (e.g. Docker); otherwise load from file (local dev).
	if os.Getenv("HOST_DB") != "" {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("read env: %w", err)
		}
		if err := cfg.validate(); err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	path := os.Getenv(envConfigPath)
	if path == "" {
		path = defaultConfigPath
	}

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validate checks the sections that can be misconfigured in a way worth
// refusing to start over. Only the dashboard has such a case today: silently
// disabling something the operator asked for is worse than failing loudly, and
// one of its settings is a security hazard if it reaches production.
func (c Config) validate() error {
	if c.Web.Enabled {
		return c.Web.Validate()
	}
	return nil
}

// DashboardURL is the page the Mini App opens — the https origin itself, as
// opposed to MiniAppURL's t.me deep link. Telegram needs the former for a
// web_app button and the latter for a link.
func (w Web) DashboardURL() string {
	if w.PublicOrigin == "" {
		return ""
	}
	return strings.TrimSuffix(w.PublicOrigin, "/") + "/"
}

// MiniAppURL is the deep link that opens the dashboard inside Telegram. Empty
// when no bot username is configured, which is how the profile menu knows not
// to draw the button yet.
func (c Config) MiniAppURL() string {
	if c.Web.BotUsername == "" || c.Web.MiniAppShortName == "" {
		return ""
	}
	return fmt.Sprintf("https://t.me/%s/%s",
		strings.TrimPrefix(c.Web.BotUsername, "@"), c.Web.MiniAppShortName)
}

func (c Config) PostgresDSN() string {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.PostreSQL.Host, c.PostreSQL.Port, c.PostreSQL.User, c.PostreSQL.Pass, c.PostreSQL.Name, c.PostreSQL.SSLMode,
	)
	return dsn
}

func (c Config) PostgresURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.PostreSQL.User,
		c.PostreSQL.Pass,
		c.PostreSQL.Host,
		c.PostreSQL.Port,
		c.PostreSQL.Name,
		c.PostreSQL.SSLMode,
	)
}
