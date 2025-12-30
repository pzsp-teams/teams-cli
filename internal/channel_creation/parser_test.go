package channelcreation

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

var _ logger.Logger = noopLogger{}

func TestParseChannelsData_WhenDecodeFails_ReturnsWrappedError(t *testing.T) {
	origErr := errors.New("boom")

	decodeFn := file_readers.DecodeFunc(func(_ io.Reader, _ any) error {
		return origErr
	})

	got, err := ParseChannelsData(bytes.NewBufferString("x"), decodeFn)
	require.Nil(t, got)
	require.Error(t, err)

	require.ErrorIs(t, err, errDataParseFailed)
	require.ErrorIs(t, err, origErr)
}

func TestParseChannelsData_WhenDecodeOK_ReturnsData(t *testing.T) {
	expected := []map[string]string{
		{"team_ref": "t1", "channel_ref": "c1", "role": "member", "user_ref": "u1"},
		{"team_ref": "t1", "channel_ref": "c1", "role": "owner", "user_ref": "u2"},
	}

	decodeFn := file_readers.DecodeFunc(func(_ io.Reader, v any) error {
		ptr, ok := v.(*[]map[string]string)
		require.True(t, ok, "decodeFn should receive *[]map[string]string")
		*ptr = append(*ptr, expected...)
		return nil
	})

	got, err := ParseChannelsData(bytes.NewBufferString("x"), decodeFn)
	require.NoError(t, err)
	require.Equal(t, expected, got)
}
