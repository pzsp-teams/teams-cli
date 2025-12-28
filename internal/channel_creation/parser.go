package channelcreation

import (
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/initializers"
)

func ParseChannelsData(r io.Reader, decodeFunc file_readers.DecodeFunc) ([]map[string]string, error) {
	channelsData := make([]map[string]string, 0)
	err := decodeFunc(r, &channelsData)
	if err != nil {
		initializers.Logger.Error(errDataParseFailed.Error(), "error", err)
		return nil, fmt.Errorf("%w: %w", errDataParseFailed, err)
	}
	initializers.Logger.Info("Channels data parsed", "channel_count", len(channelsData))
	return channelsData, nil
}