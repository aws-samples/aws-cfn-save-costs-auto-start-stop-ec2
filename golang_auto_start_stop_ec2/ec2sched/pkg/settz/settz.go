package settz

import (
	"os"
	"time"
)

func SetRegion(tz string) (string, error) {
	if tz == "" {
		tz = "UTC"
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", err
	}

	os.Setenv("TZ", loc.String())
	return loc.String(), nil
}
