package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New() zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	return logger
}
