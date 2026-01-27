package core

import (
	"io"
	"log"
	"log/slog"
	"os"
)

func SetupLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	file, err := os.OpenFile("logs/goredis.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	writer := io.MultiWriter(os.Stdout, file)
	slog.SetDefault(slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: level,
	})))
}
