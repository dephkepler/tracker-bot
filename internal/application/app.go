package application

import (
	"context"
	"fmt"
	"tracker-bot/internal/config"
	"tracker-bot/internal/dispatcher"
	"tracker-bot/internal/handlers"
	"tracker-bot/internal/repo"
	"tracker-bot/internal/scheduler"
	"tracker-bot/internal/service"
	"tracker-bot/internal/utils/pgclient"
	"tracker-bot/internal/utils/tgclient"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Application struct {
	cfg               *config.Config
	db                *pgclient.Client
	bot               *tgbotapi.BotAPI
	dispatcher        *dispatcher.Dispatcher
	timerScheduler    *scheduler.TimerScheduler
	learningScheduler *scheduler.LearningScheduler
}

func NewApplication(cfg *config.Config) *Application {
	return &Application{cfg: cfg}
}

// Build wires repositories, services, handlers and schedulers.
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

	//repositories
	entryRepo := repo.NewEntryRepository(app.db.Pool())
	profileRepo := repo.NewProfileRepository(app.db.Pool())
	trackRepo := repo.NewTrackerRepository(app.db.Pool())
	learningRepo := repo.NewLearningRepository(app.db.Pool())
	subscriptionRepo := repo.NewSubscriptionRepository(app.db.Pool())
	timerRepo := repo.NewTimerRepository(app.db.Pool())
	customTimerRepo := repo.NewCustomTimerRepository(app.db.Pool())
	sessionRepo := repo.NewSessionRepository(app.db.Pool())
	uistateRepo := repo.NewUIStateRepository(app.db.Pool())

	//services
	entrysvc := service.NewEntryService(entryRepo)
	provilesvc := service.NewProfileService(profileRepo)
	tracksvc := service.NewTrackerService(trackRepo)
	timersvc := service.NewTimerService(timerRepo, sessionRepo, customTimerRepo)
	learningsvc := service.NewLearningService(learningRepo)
	subscriptionsvc := service.NewSubscriptionService(subscriptionRepo)
	uistatesvc := service.NewUIStateService(uistateRepo)

	//handlers and dispatcher
	module := handlers.New(app.bot, entrysvc, provilesvc, tracksvc, timersvc, learningsvc, subscriptionsvc, app.cfg.AdminUsername)
	app.dispatcher = dispatcher.New(app.bot, ctx, entrysvc, provilesvc, uistatesvc, module, module, module, module, module)
	app.timerScheduler = scheduler.NewTimerScheduler(ctx, timersvc, module)
	app.learningScheduler = scheduler.NewLearningScheduler(ctx, learningsvc, module)

	return nil
}

// Run starts background jobs and blocks on dispatcher loop.
func (app *Application) Run() error {
	if app.dispatcher == nil || app.timerScheduler == nil || app.learningScheduler == nil {
		return fmt.Errorf("run application: app is not built")
	}
	app.timerScheduler.Run()
	app.learningScheduler.Run()
	app.dispatcher.Run()
	return nil
}
