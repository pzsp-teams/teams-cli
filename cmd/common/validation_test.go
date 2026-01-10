package common

import (
	"errors"
	"testing"
)

func TestValidateMessageInput(t *testing.T) {
	tests := []struct {
		name    string
		flags   MessageInputFlags
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no input specified",
			flags:   MessageInputFlags{},
			wantErr: true,
			errMsg:  "must specify one of: --template, --message, or --message-file",
		},
		{
			name: "template only",
			flags: MessageInputFlags{
				Template: "template.txt",
			},
			wantErr: false,
		},
		{
			name: "message only",
			flags: MessageInputFlags{
				Message: "hello",
			},
			wantErr: false,
		},
		{
			name: "message-file only",
			flags: MessageInputFlags{
				MessageFile: "message.txt",
			},
			wantErr: false,
		},
		{
			name: "template and message",
			flags: MessageInputFlags{
				Template: "template.txt",
				Message:  "hello",
			},
			wantErr: true,
			errMsg:  "cannot use --template, --message, and --message-file together",
		},
		{
			name: "template and message-file",
			flags: MessageInputFlags{
				Template:    "template.txt",
				MessageFile: "message.txt",
			},
			wantErr: true,
			errMsg:  "cannot use --template, --message, and --message-file together",
		},
		{
			name: "message and message-file",
			flags: MessageInputFlags{
				Message:     "hello",
				MessageFile: "message.txt",
			},
			wantErr: true,
			errMsg:  "cannot use --template, --message, and --message-file together",
		},
		{
			name: "all three specified",
			flags: MessageInputFlags{
				Template:    "template.txt",
				Message:     "hello",
				MessageFile: "message.txt",
			},
			wantErr: true,
			errMsg:  "cannot use --template, --message, and --message-file together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessageInput(tt.flags)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateMessageInput() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateMessageInput() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("ValidateMessageInput() unexpected error = %v", err)
			}
		})
	}
}

func TestProcessMessageFlags(t *testing.T) {
	mockParser := func(template, data string) (map[string]string, error) {
		if template == "error.txt" {
			return nil, errors.New("template parse error")
		}
		return map[string]string{"channel1": "message1", "channel2": "message2"}, nil
	}

	tests := []struct {
		name       string
		flags      MessageInputFlags
		recipients []string
		wantErr    bool
		errMsg     string
		wantSource string
	}{
		{
			name:       "no input specified",
			flags:      MessageInputFlags{},
			recipients: []string{"channel1"},
			wantErr:    true,
			errMsg:     "must specify one of: --template, --message, or --message-file",
		},
		{
			name: "template without data",
			flags: MessageInputFlags{
				Template: "template.txt",
			},
			recipients: []string{},
			wantErr:    true,
			errMsg:     "--data is required when using --template",
		},
		{
			name: "template with data",
			flags: MessageInputFlags{
				Template:     "template.txt",
				TemplateData: "data.yaml",
			},
			recipients: []string{},
			wantErr:    false,
			wantSource: "template",
		},
		{
			name: "template parse error",
			flags: MessageInputFlags{
				Template:     "error.txt",
				TemplateData: "data.yaml",
			},
			recipients: []string{},
			wantErr:    true,
			errMsg:     "template parse error",
		},
		{
			name: "message without recipients",
			flags: MessageInputFlags{
				Message: "hello",
			},
			recipients: []string{},
			wantErr:    true,
			errMsg:     "--channels or --chats is required when using --message",
		},
		{
			name: "message with recipients",
			flags: MessageInputFlags{
				Message: "hello",
			},
			recipients: []string{"channel1", "channel2"},
			wantErr:    false,
			wantSource: "message",
		},
		{
			name: "message-file without recipients",
			flags: MessageInputFlags{
				MessageFile: "message.txt",
			},
			recipients: []string{},
			wantErr:    true,
			errMsg:     "--channels or --chats is required when using --message-file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessMessageFlags(tt.flags, tt.recipients, mockParser)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ProcessMessageFlags() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("ProcessMessageFlags() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ProcessMessageFlags() unexpected error = %v", err)
					return
				}
				if result == nil {
					t.Errorf("ProcessMessageFlags() returned nil result")
					return
				}
				if result.Source != tt.wantSource {
					t.Errorf("ProcessMessageFlags() source = %v, want %v", result.Source, tt.wantSource)
				}
				if result.Messages == nil {
					t.Errorf("ProcessMessageFlags() messages map is nil")
				}
			}
		})
	}
}
