package logger

import (
	"io"
	"log/slog"
	"os"
)

// New creates a new logger with the given log level and io.writer
func New(level slog.Level, writers ...io.Writer) *slog.Logger {
    // If no writers provided, default to stdout
    if len(writers) == 0 {
        return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: level,
        }))
    }

	// Always use stdout
    writers = append(writers, os.Stdout)

    // Combine all writers
    multiWriter := io.MultiWriter(writers...)
    
	// Create logger
    return slog.New(slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
        Level: level,
    }))
}
