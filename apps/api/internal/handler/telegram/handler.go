package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/xromen/movietracker/internal/service"
)

type Handlers struct {
	logger          *slog.Logger
	telegramService service.TelegramService
}

func NewHandlers(
	logger *slog.Logger,
	telegramService service.TelegramService,
) *Handlers {
	return &Handlers{
		logger:          logger,
		telegramService: telegramService,
	}
}

func (h *Handlers) Register(b *bot.Bot) {
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"start",
		bot.MatchTypeCommandStartOnly,
		h.Start,
	)

	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"help",
		bot.MatchTypeCommand,
		h.Help,
	)
}

func (h *Handlers) Default(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	h.logger.Warn(
		"unknown message",
		"from", update.Message.From.ID,
		"message", update.Message.Text,
	)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Я не знаю такую команду.",
	})

	if err != nil {
		h.logger.Error("send message", "err", err)
	}
}
