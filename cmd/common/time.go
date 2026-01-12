package common

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	coreretriever "github.com/pzsp-teams/teams-cli/internal/core/retriever"
)

// ParseTimeRange parses dates given in natural language or other formats
// and returns a TimeRange struct
func ParseTimeRange(start, end string) (coreretriever.TimeRange, error) {
	var startTime, endTime *time.Time

	if start != "" {
		t, err := dateparse.ParseAny(start)
		if err != nil {
			return coreretriever.TimeRange{}, fmt.Errorf("invalid start time: %w", err)
		}
		startTime = &t
	}

	if end != "" {
		t, err := dateparse.ParseAny(end)
		if err != nil {
			return coreretriever.TimeRange{}, fmt.Errorf("invalid end time: %w", err)
		}
		endTime = &t
	}

	if err := applyDefaultTimeRange(&startTime, &endTime); err != nil {
		return coreretriever.TimeRange{}, err
	}

	return coreretriever.TimeRange{Start: *startTime, End: *endTime}, nil
}

func applyDefaultTimeRange(start, end **time.Time) error {
	now := time.Now()

	switch {
	case *start == nil && *end == nil:
		e := now
		s := now.Add(-24 * time.Hour)
		*start = &s
		*end = &e

	case *start != nil && *end == nil:
		*end = &now

	case *start == nil && *end != nil:
		s := (*end).Add(-24 * time.Hour)
		*start = &s
	}

	if !(*start).Before(**end) {
		return fmt.Errorf("start time must be before end time")
	}

	return nil
}
