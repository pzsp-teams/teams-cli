package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/templates"
)

// ParseTemplateAndData parses a template file and data file to generate messages
func ParseTemplateAndData(templatePath, dataPath string) (map[string]string, error) {
	templateFile, err := os.Open(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template file: %w", err)
	}

	dataFile, err := os.Open(dataPath)
	if err != nil {
		_ = templateFile.Close()
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}

	extension := GetFileExtension(dataPath)
	parser, err := GetDecodeFunc(extension)
	if err != nil {
		_ = templateFile.Close()
		_ = dataFile.Close()
		return nil, err
	}

	messageParser, err := templates.NewMessageParser(templateFile, dataFile, parser)
	_ = templateFile.Close()
	_ = dataFile.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to create message parser: %w", err)
	}

	return messageParser.Parse()
}

// GetFileExtension returns the extension of a file without the leading dot
func GetFileExtension(filename string) string {
	return strings.TrimPrefix(filepath.Ext(filename), ".")
}

// GetDecodeFunc returns the decoder function for a given file extension
func GetDecodeFunc(extension string) (file_readers.DecodeFunc, error) {
	return file_readers.GetDecoderByExtension(extension)
}
