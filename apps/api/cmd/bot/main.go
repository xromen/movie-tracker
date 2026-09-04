package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-telegram/bot"
	"github.com/joho/godotenv"
	"github.com/xromen/movietracker/internal/config"
	handler "github.com/xromen/movietracker/internal/handler/telegram"
	"github.com/xromen/movietracker/internal/platform/database"
	"github.com/xromen/movietracker/internal/platform/logger"
	"github.com/xromen/movietracker/internal/repository"
	"github.com/xromen/movietracker/internal/service"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger, closeLogWriter := logger.CreateLogger(cfg.Logging.FilePath, "bot")
	defer closeLogWriter()

	slog.SetDefault(logger)
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.Database.PoolConfig())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	telegramRepo := repository.NewTelegramRepository(pool)
	telegramService := service.NewTelegramService(telegramRepo, logger, service.TelegramConfig{
		BotUserName:     cfg.Telegram.BotUsername,
		BindingTokenTTL: cfg.Telegram.BindingTokenTTL,
	})

	// Telegram handlers
	botHandlers := handler.NewHandlers(logger, telegramService)

	// Telegram Bot
	opts := []bot.Option{
		bot.WithDefaultHandler(botHandlers.Default),
		bot.WithServerURL(cfg.Telegram.TelegramApiBaseUrl),
	}

	tgBot, err := bot.New(
		cfg.Telegram.BotToken,
		opts...,
	)
	if err != nil {
		logger.Error("create telegram bot", "error", err)
		os.Exit(1)
	}

	botHandlers.Register(tgBot)

	logger.Info("telegram bot started")

	tgBot.Start(ctx)

	logger.Info("telegram bot stopped")
}
