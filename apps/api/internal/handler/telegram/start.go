package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/xromen/movietracker/internal/domain"
)

func (h *Handlers) Start(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
) {
	if update.Message == nil {
		return
	}

	command := update.Message.Text
	splittedCommand := strings.Split(command, " ")

	user := update.Message.From

	if len(splittedCommand) == 2 {
		h.bindingToken(ctx, b, update, splittedCommand[1])
		return
	}

	text := fmt.Sprintf(
		"Привет, %s! 👋",
		user.FirstName,
	)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   text,
	})

	if err != nil {
		h.logger.Error("send start message", "err", err)
	}
}

func (h *Handlers) bindingToken(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
	token string,
) {
	user := update.Message.From
	bdUser, err := h.telegramService.HandleBindingToken(ctx, token, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrBindingTokenExpired) {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Данный токен привязки устарел",
			})
		} else {
			h.logger.Error("binding token", "error", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Произошла непредвиденная ошибка",
			})
		}
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Телеграм успешно привязан к аккаунту приложения %s", bdUser.Username),
	})
}
