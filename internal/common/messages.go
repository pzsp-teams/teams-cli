package common

import (
	"fmt"
	"os"

	"github.com/pzsp-teams/cli/internal/templates"
)

// GetMessageContent reads message content from either a string or file and processes it with RawToHTML
// Returns the processed message content or an error if file reading fails
func GetMessageContent(messageStr, messageFilePath string) (string, error) {
	if messageStr != "" && messageFilePath != "" {
		return "", fmt.Errorf("cannot specify both message and message-file")
	}
	if messageStr == "" && messageFilePath == "" {
		return "", fmt.Errorf("must specify either message or message-file")
	}

	if messageStr != "" {
		return templates.RawToHTML([]byte(messageStr)), nil
	}

	content, err := os.ReadFile(messageFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read message file: %w", err)
	}

	return templates.RawToHTML(content), nil
}

// CreateMessagesFromString creates a map of messages for multiple recipients from a single message string
func CreateMessagesFromString(message string, recipients []string) map[string]string {
	messages := make(map[string]string, len(recipients))
	message = templates.RawToHTML([]byte(message))
	for _, recipient := range recipients {
		messages[recipient] = message
	}
	return messages
}

// CreateMessagesFromFile creates a map of messages for multiple recipients from a file
func CreateMessagesFromFile(filePath string, recipients []string) (map[string]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read message file: %w", err)
	}

	message := templates.RawToHTML(content)
	messages := make(map[string]string, len(recipients))
	for _, recipient := range recipients {
		messages[recipient] = message
	}
	return messages, nil
}
