package helper

import (
	"time"
)

func CoverToTimestamp(timeStr, layout string) (time.Time, error) {
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}
