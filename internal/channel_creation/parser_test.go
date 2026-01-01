package channelcreation

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/stretchr/testify/require"
)

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
	expected := map[string]ChannelData{
		"c1": {
			"members": []string{"u1"},
			"owners":  []string{"u2"},
		},
		"c2": {
			"members": []string{},
			"owners":  []string{"u3", "u4"},
		},
	}

	decodeFn := file_readers.DecodeFunc(func(_ io.Reader, v any) error {
		ptr, ok := v.(*map[string]ChannelData)
		require.True(t, ok, "decodeFn should receive *map[string]ChannelData")
		*ptr = expected
		return nil
	})

	got, err := ParseChannelsData(bytes.NewBufferString("x"), decodeFn)
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func TestParseChannelsDataFromCSV_Success(t *testing.T) {
	csv := "channel_ref,role,user_ref\ntest-general2,owner,kmarszalek@teamspzsp.onmicrosoft.com\ntest-general2,member,ddsouza@teamspzsp.onmicrosoft.com\ntest-random,owner,msuski@teamspzsp.onmicrosoft.com\ntest-random,owner,kmarszalek@teamspzsp.onmicrosoft.com\n"

	got, err := parseChannelsDataFromCSV(bytes.NewBufferString(csv))
	require.NoError(t, err)

	require.Len(t, got, 2)
	require.Equal(t, ChannelData{
		"members": []string{"ddsouza@teamspzsp.onmicrosoft.com"},
		"owners":  []string{"kmarszalek@teamspzsp.onmicrosoft.com"},
	}, got["test-general2"])

	require.Equal(t, ChannelData{
		"members": []string{},
		"owners":  []string{"msuski@teamspzsp.onmicrosoft.com", "kmarszalek@teamspzsp.onmicrosoft.com"},
	}, got["test-random"])
}

func TestParseChannelsDataFromCSV_InvalidCSV_ReturnsError(t *testing.T) {
	csv := "channel_ref,role,user_ref\nc1,member\n"

	got, err := parseChannelsDataFromCSV(bytes.NewBufferString(csv))
	require.Error(t, err)
	require.Nil(t, got)
}

func TestParseChannelsDataByExtension_CSV(t *testing.T) {
	csv := "channel_ref,role,user_ref\nc1,member,u1\nc1,owner,u2\n"

	got, err := ParseChannelsDataByExtension(bytes.NewBufferString(csv), "csv")
	require.NoError(t, err)

	require.Len(t, got, 1)
	require.Equal(t, ChannelData{
		"members": []string{"u1"},
		"owners":  []string{"u2"},
	}, got["c1"])
}

func TestParseChannelsDataByExtension_JSON(t *testing.T) {
	json := `{"c1": {"members": ["u1"], "owners": ["u2"]}}`

	got, err := ParseChannelsDataByExtension(bytes.NewBufferString(json), "json")
	require.NoError(t, err)

	require.Len(t, got, 1)
	require.Equal(t, ChannelData{
		"members": []string{"u1"},
		"owners":  []string{"u2"},
	}, got["c1"])
}

func TestParseChannelsDataByExtension_YAML(t *testing.T) {
	yaml := "c1:\n  members:\n    - u1\n  owners:\n    - u2\n"

	got, err := ParseChannelsDataByExtension(bytes.NewBufferString(yaml), "yaml")
	require.NoError(t, err)

	require.Len(t, got, 1)
	require.Equal(t, ChannelData{
		"members": []string{"u1"},
		"owners":  []string{"u2"},
	}, got["c1"])
}

func TestParseChannelsDataByExtension_UnsupportedExtension(t *testing.T) {
	got, err := ParseChannelsDataByExtension(bytes.NewBufferString("data"), "xml")
	require.Error(t, err)
	require.Nil(t, got)
	require.Contains(t, err.Error(), "unsupported file extension")
}

func TestTransformCSVRowsToChannelData_GroupsByChannel(t *testing.T) {
	rows := []map[string]string{
		{"channel_ref": "c1", "role": "member", "user_ref": "u1"},
		{"channel_ref": "c1", "role": "owner", "user_ref": "u2"},
		{"channel_ref": "c2", "role": "member", "user_ref": "u3"},
		{"channel_ref": "c3", "role": "owner", "user_ref": "u4"},
		{"channel_ref": "c3", "role": "owner", "user_ref": "u5"},
	}

	got := transformCSVRowsToChannelData(rows)

	require.Len(t, got, 3)
	require.Equal(t, ChannelData{
		"members": []string{"u1"},
		"owners":  []string{"u2"},
	}, got["c1"])
	require.Equal(t, ChannelData{
		"members": []string{"u3"},
		"owners":  []string{},
	}, got["c2"])
	require.Equal(t, ChannelData{
		"members": []string{},
		"owners":  []string{"u4", "u5"},
	}, got["c3"])
}
