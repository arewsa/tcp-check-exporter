package main

import (
    "io"
    "log/slog"
    "os"
)

func SetupLogging(cfg *ParsedConfig) error {
    var level slog.Level
    switch cfg.LogLevel {
    case "debug":
        level = slog.LevelDebug
    case "info":
        level = slog.LevelInfo
    case "warn":
        level = slog.LevelWarn
    case "error":
        level = slog.LevelError
    default:
        level = slog.LevelInfo
    }

    var output io.Writer
    switch cfg.LogOutput {
    case "stdout":
        output = os.Stdout
    case "stderr":
        output = os.Stderr
    case "file":
        if cfg.LogFile == "" {
            cfg.LogFile = "/var/log/host-exporter.log"
        }
        file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
        if err != nil {
            return err
        }
        output = file
    default:
        output = os.Stdout
    }

    var handler slog.Handler
    opts := &slog.HandlerOptions{Level: level}

    switch cfg.LogFormat {
    case "json":
        handler = slog.NewJSONHandler(output, opts)
    default:
        handler = slog.NewTextHandler(output, opts)
    }

    slog.SetDefault(slog.New(handler))

    slog.Info("Logging configured",
        "level", cfg.LogLevel,
        "format", cfg.LogFormat,
        "output", cfg.LogOutput,
    )

    return nil
}