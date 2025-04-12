package helper

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

var log = zerolog.New(os.Stdout).With().Timestamp().Logger()

func CoverToTimestamp(timeStr, layout string) (time.Time, error) {
	log.Info().Str("timeStr", timeStr).Str("layout", layout).Msg("Attempting to convert string to timestamp")
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		log.Error().Err(err).Str("timeStr", timeStr).Msg("Failed to parse time string")
		return time.Time{}, err
	}

	log.Info().Str("timeStr", timeStr).Str("layout", layout).Msg("Successfully converted to timestamp")
	return t, nil
}
