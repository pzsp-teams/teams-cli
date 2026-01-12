package creator

import (
	"fmt"
	"io"

	"github.com/pzsp-teams/teams-cli/internal/file_readers"
	"github.com/pzsp-teams/teams-cli/internal/initializers"
)

// ParseChannelsData parses channel creation data from the provided reader
func ParseChannelsData(r io.Reader, decodeFunc file_readers.DecodeFunc) (map[string]ChannelData, error) {
	channelsData := make(map[string]ChannelData, 0)
	err := decodeFunc(r, &channelsData)
	if err != nil {
		initializers.Logger.Error(errDataParseFailed.Error(), "error", err)
		return nil, fmt.Errorf("%w: %w", errDataParseFailed, err)
	}
	initializers.Logger.Info("Channels data parsed", "channel_count", len(channelsData))
	return channelsData, nil
}

// ParseChannelsDataByExtension parses channel creation data based on file extension
func ParseChannelsDataByExtension(r io.Reader, extension string) (map[string]ChannelData, error) {
	if extension == "csv" {
		return parseChannelsDataFromCSV(r)
	}

	decoder, err := file_readers.GetDecoderByExtension(extension)
	if err != nil {
		return nil, err
	}

	return ParseChannelsData(r, decoder)
}

func parseChannelsDataFromCSV(r io.Reader) (map[string]ChannelData, error) {
	var rows []map[string]string
	err := file_readers.DecodeCSV(r, &rows)
	if err != nil {
		initializers.Logger.Error(errDataParseFailed.Error(), "error", err)
		return nil, fmt.Errorf("%w: %w", errDataParseFailed, err)
	}

	channelsData := transformCSVRowsToChannelData(rows)
	initializers.Logger.Info("Channels data parsed from CSV", "channel_count", len(channelsData))
	return channelsData, nil
}

func transformCSVRowsToChannelData(rows []map[string]string) map[string]ChannelData {
	grouped := file_readers.GroupBy(rows, func(row map[string]string) string {
		return row["channel_ref"]
	})

	result := make(map[string]ChannelData, len(grouped))
	for channelRef, channelRows := range grouped {
		channelData := ChannelData{
			"members": []string{},
			"owners":  []string{},
		}
		for _, row := range channelRows {
			role := row["role"]
			userRef := row["user_ref"]
			switch role {
			case "member":
				channelData["members"] = append(channelData["members"], userRef)
			case "owner":
				channelData["owners"] = append(channelData["owners"], userRef)
			}
		}
		result[channelRef] = channelData
	}

	return result
}
