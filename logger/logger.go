package logger

import (
	"io"
	"log/slog"
	"os"
)

// New creates a new logger with the given log level and io.writer
func New(level slog.Level, writers ...io.Writer) *slog.Logger {
	// Always include stdout writer
	validWriters := []io.Writer{os.Stdout}

	// Filter out nil writers
	for _, w := range writers {
		if w != nil {
			validWriters = append(validWriters, w)
		}
	}

	// Combine writers into multiwriter
	mWriter := io.MultiWriter(validWriters...)
	// Set log level
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Create new logger
	return slog.New(slog.NewTextHandler(mWriter, opts))
}
