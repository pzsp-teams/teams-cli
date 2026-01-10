package common

import (
	"errors"

	"github.com/pzsp-teams/cli/internal/common"
)

// MessageInputFlags represents the flags for message input
type MessageInputFlags struct {
	Template     string
	TemplateData string
	Message      string
	MessageFile  string
}

// ProcessedMessages contains the processed messages and metadata
type ProcessedMessages struct {
	Messages map[string]string
	Source   string // "template", "message", or "file"
}

// ValidateMessageInput validates that exactly one message input method is specified
func ValidateMessageInput(flags MessageInputFlags) error {
	inputMethods := 0
	if flags.Template != "" {
		inputMethods++
	}
	if flags.Message != "" {
		inputMethods++
	}
	if flags.MessageFile != "" {
		inputMethods++
	}

	if inputMethods == 0 {
		return errors.New("must specify one of: --template, --message, or --message-file")
	}
	if inputMethods > 1 {
		return errors.New("cannot use --template, --message, and --message-file together")
	}

	return nil
}

// ProcessMessageFlags processes message input flags and returns processed messages
func ProcessMessageFlags(
	flags MessageInputFlags,
	recipients []string,
	templateParser func(template, data string) (map[string]string, error),
) (*ProcessedMessages, error) {
	if err := ValidateMessageInput(flags); err != nil {
		return nil, err
	}

	switch {
	case flags.Template != "":
		if flags.TemplateData == "" {
			return nil, errors.New("--data is required when using --template")
		}
		messages, err := templateParser(flags.Template, flags.TemplateData)
		if err != nil {
			return nil, err
		}
		return &ProcessedMessages{Messages: messages, Source: "template"}, nil

	case flags.Message != "":
		if len(recipients) == 0 {
			return nil, errors.New("--channels or --chats is required when using --message")
		}
		messages := common.CreateMessagesFromString(flags.Message, recipients)
		return &ProcessedMessages{Messages: messages, Source: "message"}, nil

	case flags.MessageFile != "":
		if len(recipients) == 0 {
			return nil, errors.New("--channels or --chats is required when using --message-file")
		}
		messages, err := common.CreateMessagesFromFile(flags.MessageFile, recipients)
		if err != nil {
			return nil, err
		}
		return &ProcessedMessages{Messages: messages, Source: "file"}, nil
	}

	return nil, errors.New("internal error: no input method selected")
}
