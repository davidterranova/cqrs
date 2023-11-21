package xhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type HTTPConfig struct {
	Host              string        `envconfig:"HOST" default:""`
	Port              int           `envconfig:"PORT" default:"8080"`
	ReadTimeout       time.Duration `envconfig:"READ_TIMEOUT" default:"5s"`
	ReadHeaderTimeout time.Duration `envconfig:"READ_HEADER_TIMEOUT" default:"1s"`
	WriteTimeout      time.Duration `envconfig:"WRITE_TIMEOUT" default:"5s"`
}

// Server is a filestorage http server
type Server struct {
	host              string
	port              int
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	handler           http.Handler
}

// NewServer creates a new http server given a handler and a configuration
func NewServer(handler http.Handler, cfg HTTPConfig) *Server {
	return &Server{
		host:              cfg.Host,
		port:              cfg.Port,
		readTimeout:       cfg.ReadTimeout,
		readHeaderTimeout: cfg.ReadHeaderTimeout,
		writeTimeout:      cfg.WriteTimeout,
		handler:           handler,
	}
}

// Address returns the host and port expected from an http server
func (s Server) Address() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

// Serve starts the server
func (s Server) Serve(ctx context.Context) error {
	srv := http.Server{
		Addr:              s.Address(),
		Handler:           CORS()(s.handler),
		WriteTimeout:      s.writeTimeout,
		ReadTimeout:       s.readTimeout,
		ReadHeaderTimeout: s.readHeaderTimeout,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.
				Fatal().
				Err(err).
				Msg("http server crashed")
		}
	}()

	log.
		Info().
		Str("address", s.Address()).
		Msg("http server started")

	<-ctx.Done()

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctxShutDown)
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to shutdown http server properly: %s", err)
	}

	return nil
}
