package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/xromen/movietracker/internal/domain"
)

const telegramBindingUrlFormat = "https://t.me/%s?start=%s"

type TelegramConfig struct {
	BotUserName     string
	BindingTokenTTL time.Duration
}

type TelegramService interface {
	HandleBindingToken(ctx context.Context, code string, telegramID int64) (*UserOutput, error)
	GenerateBindingUrl(ctx context.Context, userID int64) (*BindingUrlOutput, error)
}

type TelegramWorker interface {
	Run(ctx context.Context) error
}

type telegramRepository interface {
	CreateBindingToken(ctx context.Context, token *domain.BindingToken) error
	GetBindingToken(ctx context.Context, userID int64) (*domain.BindingToken, error)
	GetUserByBindingToken(ctx context.Context, code string) (*domain.User, error)
	SetTelegramID(ctx context.Context, userID, telegramID int64) error

	GetDueReportMessages(ctx context.Context, limit int) ([]domain.DueReportMessage, error)
	MarkMessageSuccess(ctx context.Context, messageID, telegramMessageID int64) error
	MarkMessageFailure(ctx context.Context, messageID int64, nextRetry time.Time, error string) error
}

type telegramService struct {
	logger          *slog.Logger
	repo            telegramRepository
	botUserName     string
	bindingTokenTTL time.Duration
}

type telegramWorker struct {
	logger *slog.Logger
	repo   telegramRepository
	bot    *bot.Bot
}

type BindingUrlOutput struct {
	URL       string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type UserOutput struct {
	ID         int64
	Username   string
	Email      string
	TelegramID int64
}

func NewTelegramService(repo telegramRepository, logger *slog.Logger, config TelegramConfig) TelegramService {
	return &telegramService{
		logger:          logger,
		repo:            repo,
		botUserName:     config.BotUserName,
		bindingTokenTTL: config.BindingTokenTTL,
	}
}

func NewTelegramWorker(repo telegramRepository, logger *slog.Logger, bot *bot.Bot) TelegramWorker {
	return &telegramWorker{
		logger: logger,
		repo:   repo,
		bot:    bot,
	}
}

func (w *telegramWorker) Run(ctx context.Context) error {
	w.logger.Info("telegram worker started")

	w.runOnce(ctx)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (s *telegramService) HandleBindingToken(ctx context.Context, token string, telegramID int64) (*UserOutput, error) {
	user, err := s.repo.GetUserByBindingToken(ctx, token)

	if err != nil {
		return nil, fmt.Errorf("handle binding token: %w", err)
	}

	err = s.repo.SetTelegramID(ctx, user.ID, telegramID)

	if err != nil {
		return nil, fmt.Errorf("handle binding token: %w", err)
	}

	return &UserOutput{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		TelegramID: telegramID,
	}, nil
}

func (s *telegramService) GenerateBindingUrl(ctx context.Context, userID int64) (*BindingUrlOutput, error) {
	token, err := s.repo.GetBindingToken(ctx, userID)

	if err == nil {
		return &BindingUrlOutput{
			URL:       fmt.Sprintf(telegramBindingUrlFormat, s.botUserName, token.Token),
			ExpiresAt: token.ExpiresAt,
			CreatedAt: token.CreatedAt,
		}, nil
	}

	token = &domain.BindingToken{
		UserID:    userID,
		Token:     rand.Text()[0:10],
		ExpiresAt: time.Now().Add(s.bindingTokenTTL),
	}

	err = s.repo.CreateBindingToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("generate binding url: %w", err)
	}

	return &BindingUrlOutput{
		URL:       fmt.Sprintf(telegramBindingUrlFormat, s.botUserName, token.Token),
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	}, nil
}

func (w *telegramWorker) runOnce(ctx context.Context) {
	for {
		messages, err := w.repo.GetDueReportMessages(ctx, batchSize)
		if err != nil {
			w.logger.Error("failed to get due report messages", "error", err)
			return
		}

		if len(messages) == 0 {
			return
		}

		for _, message := range messages {
			if telegramMessage, err := w.bot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    message.UserTelegramID,
				Text:      message.Message,
				ParseMode: models.ParseModeHTML,
			}); err != nil {
				_ = w.repo.MarkMessageFailure(ctx, message.ID, time.Now().Add(retryDelay), err.Error())
			} else {
				_ = w.repo.MarkMessageSuccess(ctx, message.ID, int64(telegramMessage.ID))
			}
		}
	}
}
