package xlogs

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	DefaultLogLevel = zerolog.ErrorLevel

	LogFormatterText = "text"
	LogFormatterJSON = "json"
)

type LoggerConfig struct {
	LogLevel     string `envconfig:"LEVEL" default:"info"`
	LogFormatter string `envconfig:"FORMATTER" default:"text"`
	Host         string `envconfig:"HOST" default:"localhost"`
}

type LoggerOpts func() zerolog.Logger

func Setup(opts ...LoggerOpts) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.DurationFieldInteger = true
	zerolog.MessageFieldName = "msg"
	zerolog.SetGlobalLevel(DefaultLogLevel)
	log.Logger = log.Output(os.Stderr).With().Caller().Logger()

	for _, opt := range opts {
		log.Logger = opt()
	}
}

func WithLevel(level string) func() zerolog.Logger {
	return func() zerolog.Logger {
		logLevel, err := zerolog.ParseLevel(strings.ToLower(level))
		if err != nil {
			log.
				Error().
				Err(err).
				Msg("Error parsing provided log level")
		}
		zerolog.SetGlobalLevel(logLevel)

		return log.Logger
	}
}

func WithFormatter(formatter string) func() zerolog.Logger {
	return func() zerolog.Logger {
		if formatter == LogFormatterText {
			output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
			return log.Output(output).With().Caller().Logger()
		}

		return log.Output(os.Stderr).With().Logger()
	}
}

func WithHostname(hostname string) func() zerolog.Logger {
	return func() zerolog.Logger {
		return log.With().Str("host", hostname).Logger()
	}
}
