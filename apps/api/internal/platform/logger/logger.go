package logger

import (
	"context"
	"log/slog"

	"github.com/xromen/movietracker/internal/handler"
)

func FromContext(ctx context.Context) *slog.Logger {
	logger := slog.Default()

	if traceID := handler.TraceIDFromContext(ctx); traceID != "" {
		logger = logger.With("trace_id", traceID)
	}

	return logger
}
