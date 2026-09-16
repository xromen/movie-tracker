package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handlers) Help(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: `
		Доступные команды:

		/start — начать работу
		/help — помощь`,
	})
}