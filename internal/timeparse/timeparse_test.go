package timeparse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseTimeRange(t *testing.T) {
	t.Parallel()

	startFixed := "2024-01-01T10:00:00Z"
	endFixed := "2024-01-01T12:00:00Z"

	startT := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	endT := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		start      string
		end        string
		assertFn   func(t *testing.T, gotStart, gotEnd time.Time)
		wantErrMsg string
	}{
		{
			name:  "start+end provided -> returns exact values",
			start: startFixed,
			end:   endFixed,
			assertFn: func(t *testing.T, gotStart, gotEnd time.Time) {
				require.Equal(t, startT, gotStart)
				require.Equal(t, endT, gotEnd)
			},
		},
		{
			name:  "both empty -> defaults to last 24h (end=now)",
			start: "",
			end:   "",
			assertFn: func(t *testing.T, gotStart, gotEnd time.Time) {
				now := time.Now()
				require.WithinDuration(t, now, gotEnd, 2*time.Second)
				require.WithinDuration(t, now.Add(-24*time.Hour), gotStart, 2*time.Second)
				require.True(t, gotStart.Before(gotEnd))
			},
		},
		{
			name:  "end empty -> end defaults to now",
			start: startFixed,
			end:   "",
			assertFn: func(t *testing.T, gotStart, gotEnd time.Time) {
				require.Equal(t, startT, gotStart)
				require.WithinDuration(t, time.Now(), gotEnd, 2*time.Second)
				require.True(t, gotStart.Before(gotEnd))
			},
		},
		{
			name:  "start empty -> start defaults to end-24h",
			start: "",
			end:   endFixed,
			assertFn: func(t *testing.T, gotStart, gotEnd time.Time) {
				require.Equal(t, endT, gotEnd)
				require.Equal(t, endT.Add(-24*time.Hour), gotStart)
				require.True(t, gotStart.Before(gotEnd))
			},
		},
		{
			name:       "invalid start -> returns wrapped error",
			start:      "not-a-date",
			end:        endFixed,
			wantErrMsg: "invalid start time:",
		},
		{
			name:       "invalid end -> returns wrapped error",
			start:      startFixed,
			end:        "not-a-date",
			wantErrMsg: "invalid end time:",
		},
		{
			name:       "start == end -> validation error",
			start:      startFixed,
			end:        startFixed,
			wantErrMsg: "start time must be before end time",
		},
		{
			name:       "start in the future with end empty -> end=now, validation error",
			start:      "2099-01-01T00:00:00Z",
			end:        "",
			wantErrMsg: "start time must be before end time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseTimeRange(tt.start, tt.end)

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotZero(t, got.Start)
			require.NotZero(t, got.End)

			tt.assertFn(t, got.Start, got.End)
		})
	}
}
