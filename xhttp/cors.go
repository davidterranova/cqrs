package xhttp

import (
	"net/http"

	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
)

func CORS() func(http.Handler) http.Handler {
	return cors.AllowAll().Handler
}

type CORSLogger struct{}

func (c CORSLogger) Printf(format string, v ...interface{}) {
	log.Info().Msgf(format, v...)
}
