package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

func InitLogger(level string, pretty bool, logDir string) (*slog.Logger, error) {
	logLevel := slog.LevelInfo
	switch level {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: logLevel}

	var terminal slog.Handler
	if pretty {
		terminal = NewHandler(WithColor(true), WithLevel(logLevel), WithEncoder(JSON), WithWriter(os.Stdout))
	} else {
		terminal = slog.NewJSONHandler(os.Stdout, opts)
	}

	if logDir == "" {
		return slog.New(terminal), nil
	}

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	logFile := fmt.Sprintf("%s/server.log", logDir)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	file := slog.NewJSONHandler(f, opts)

	return slog.New(multiHandler{terminal, file}), nil
}

// multiHandler fans out log records to multiple slog.Handler implementations.
type multiHandler []slog.Handler

func (m multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make(multiHandler, len(m))
	for i, h := range m {
		next[i] = h.WithAttrs(attrs)
	}
	return next
}

func (m multiHandler) WithGroup(name string) slog.Handler {
	next := make(multiHandler, len(m))
	for i, h := range m {
		next[i] = h.WithGroup(name)
	}
	return next
}
