package config

import (
	"time"

	"github.com/adriansr/github-api-service/util"
)

// Duration is a wrapper needed to decode a `time.Duration` from json
type Duration struct {
	Duration time.Duration
}

// UnmarshalJSON uses `time.ParseDuration` to parse a json string as a
// duration
func (d *Duration) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		unquoted := string(b[1 : len(b)-1])
		parsed, err := time.ParseDuration(unquoted)
		if err != nil {
			return err
		}
		d.Duration = parsed
		return nil
	}
	return util.NewError("expected a string to decode a Duration")
}
