package logger

import (
    "io"
    "os"
    "strings"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func Init() {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
    levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
    level, err := zerolog.ParseLevel(levelStr)
    if err != nil {
        level = zerolog.InfoLevel
    }

    format := strings.ToLower(os.Getenv("LOG_FORMAT"))
    var output io.Writer
    if format == "json" {
        output = os.Stdout
    } else {
        output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
    }

    log.Logger = zerolog.New(output).Level(level).With().Timestamp().Caller().Logger()
    log.Info().Str("level", level.String()).Str("format", format).Msg("Logger initialized")
}
