package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func CreateLogger(filePath, appName string) (*slog.Logger, func()) {
	logWriter, closeLogWriter, err := createLogWriter(filePath)
	if err != nil {
		slog.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	logger = logger.With("app", appName)

	return logger, closeLogWriter
}

func createLogWriter(filePath string) (io.Writer, func(), error) {
	if filePath == "" {
		return os.Stdout, func() {}, nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return nil, nil, err
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}

	return io.MultiWriter(os.Stdout, file), func() {
		_ = file.Close()
	}, nil
}

// func FromContext(ctx context.Context) *slog.Logger {
// 	logger := slog.Default()

// 	if traceID := handler.TraceIDFromContext(ctx); traceID != "" {
// 		logger = logger.With("trace_id", traceID)
// 	}

// 	return logger
// }
