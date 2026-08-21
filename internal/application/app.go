package application

import (
	"context"
	"fmt"
	"tracker-bot/internal/config"
	"tracker-bot/internal/dispatcher"
	"tracker-bot/internal/handlers"
	"tracker-bot/internal/llm"
	"tracker-bot/internal/repo"
	"tracker-bot/internal/scheduler"
	"tracker-bot/internal/service"
	"tracker-bot/internal/utils/pgclient"
	"tracker-bot/internal/utils/tgclient"
	"tracker-bot/internal/web"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type Application struct {
	cfg                *config.Config
	db                 *pgclient.Client
	bot                *tgbotapi.BotAPI
	dispatcher         *dispatcher.Dispatcher
	timerScheduler     *scheduler.TimerScheduler
	learningScheduler  *scheduler.LearningScheduler
	challengeScheduler *scheduler.ChallengeScheduler
	roadmapScheduler   *scheduler.RoadmapScheduler
	// webServer is nil unless WEB_ENABLED is set — the dashboard is optional
	// and the bot must run identically without it.
	webServer *web.Server
}

func NewApplication(cfg *config.Config) *Application {
	return &Application{cfg: cfg}
}

func (app *Application) Build(ctx context.Context) error {
	if app.cfg == nil {
		return fmt.Errorf("build application: nil config")
	}

	db, err := pgclient.New(ctx, app.cfg.PostgresDSN())
	if err != nil {
		return fmt.Errorf("init pg client: %w", err)
	}
	app.db = db

	bot, err := tgclient.New(app.cfg.Telegram.TelegramToken)
	if err != nil {
		return fmt.Errorf("init telegram bot: %w", err)
	}
	bot.Debug = app.cfg.Telegram.TelegramBotDebug
	app.bot = bot

	entryRepo := repo.NewEntryRepository(app.db.Pool())
	profileRepo := repo.NewProfileRepository(app.db.Pool())
	trackRepo := repo.NewTrackerRepository(app.db.Pool())
	learningRepo := repo.NewLearningRepository(app.db.Pool())
	subscriptionRepo := repo.NewSubscriptionRepository(app.db.Pool())
	timerRepo := repo.NewTimerRepository(app.db.Pool())
	customTimerRepo := repo.NewCustomTimerRepository(app.db.Pool())
	sessionRepo := repo.NewSessionRepository(app.db.Pool())
	uistateRepo := repo.NewUIStateRepository(app.db.Pool())
	adminRepo := repo.NewAdminRepository(app.db.Pool())
	challengeRepo := repo.NewChallengeRepository(app.db.Pool())
	roadmapRepo := repo.NewRoadmapRepository(app.db.Pool())

	entrysvc := service.NewEntryService(entryRepo)
	provilesvc := service.NewProfileService(profileRepo)
	tracksvc := service.NewTrackerService(trackRepo)
	timersvc := service.NewTimerService(timerRepo, sessionRepo, customTimerRepo)
	learningsvc := service.NewLearningService(learningRepo)
	subscriptionsvc := service.NewSubscriptionService(subscriptionRepo)
	uistatesvc := service.NewUIStateService(uistateRepo)
	adminsvc := service.NewAdminService(adminRepo)
	challengesvc := service.NewChallengeService(challengeRepo)
	roadmapsvc := service.NewRoadmapService(roadmapRepo)

	// An unconfigured LLM is not an error: the registry comes back empty and
	// every AI feature reports itself off, which is what hides the buttons.
	// A configured-but-broken one does stop startup — see llm.NewRegistry.
	llmRegistry, err := llm.NewRegistry(app.cfg.LLM)
	if err != nil {
		return fmt.Errorf("init llm: %w", err)
	}
	if llmRegistry.Enabled() {
		log.Info().Msg("roadmap ai enabled")
	}
	roadmapaisvc := service.NewRoadmapAIService(roadmapRepo, llmRegistry)

	module := handlers.New(app.bot, entrysvc, provilesvc, tracksvc, timersvc, learningsvc, subscriptionsvc, adminsvc, challengesvc, roadmapsvc, roadmapaisvc, app.cfg.MiniAppURL(), app.cfg.AdminUsername)
	app.dispatcher = dispatcher.New(app.bot, ctx, entrysvc, provilesvc, uistatesvc, module, module, module, module, module, module, module)
	app.timerScheduler = scheduler.NewTimerScheduler(ctx, timersvc, module)
	app.learningScheduler = scheduler.NewLearningScheduler(ctx, learningsvc, module)
	app.challengeScheduler = scheduler.NewChallengeScheduler(ctx, challengesvc, module)
	app.roadmapScheduler = scheduler.NewRoadmapScheduler(ctx, roadmapsvc, module)

	// The dashboard is opt-in. A misconfigured one fails the build rather than
	// starting the bot with the feature silently off — same rule as the LLM
	// registry above.
	if app.cfg.Web.Enabled {
		webServer, err := web.NewServer(ctx, app.cfg.Web, web.Deps{
			BotToken:  app.cfg.Telegram.TelegramToken,
			Entry:     entrysvc,
			Profile:   provilesvc,
			Tracker:   tracksvc,
			Roadmap:   roadmapsvc,
			RoadmapAI: roadmapaisvc,
		})
		if err != nil {
			return fmt.Errorf("init web server: %w", err)
		}
		app.webServer = webServer
		log.Info().Str("addr", app.cfg.Web.Addr).Msg("dashboard enabled")

		// Put the dashboard on the button beside the message field. Best effort
		// on purpose: this is a convenience, and Telegram being briefly
		// unreachable must not stop the bot from starting.
		if url := app.cfg.Web.DashboardURL(); url != "" {
			if err := tgclient.SetWebAppMenuButton(ctx, app.cfg.Telegram.TelegramToken,
				app.cfg.Web.MenuButtonText, url); err != nil {
				log.Warn().Err(err).Msg("could not set the Mini App menu button")
			} else {
				log.Info().Str("url", url).Msg("Mini App menu button set")
			}
		}
	}

	return nil
}

// blocks until the dispatcher loop exits; the schedulers above already run in the background
func (app *Application) Run() error {
	if app.dispatcher == nil || app.timerScheduler == nil || app.learningScheduler == nil || app.challengeScheduler == nil || app.roadmapScheduler == nil {
		return fmt.Errorf("run application: app is not built")
	}
	app.timerScheduler.Run()
	app.learningScheduler.Run()
	app.challengeScheduler.Run()
	app.roadmapScheduler.Run()
	if app.webServer != nil {
		// A fifth background worker, same shape as the schedulers: it serves
		// until the app context is cancelled and never returns an error here,
		// because a dead listener must not take the bot down with it.
		go app.webServer.Run()
	}
	app.dispatcher.Run()
	return nil
}
