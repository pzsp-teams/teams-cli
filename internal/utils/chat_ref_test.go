package utils

import (
	"testing"

	"github.com/pzsp-teams/lib/chats"
)

func TestGetChatRef(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  any
		wantValue string
	}{
		{"email goes to one-on-one", "test@example.com", chats.OneOnOneChatRef{}, "test@example.com"},
		{"email with spaces", "  user+tag@gmail.com ", chats.OneOnOneChatRef{}, "  user+tag@gmail.com "},
		{"group ref normal", "dev-team", chats.GroupChatRef{}, "dev-team"},
		{"group ref with @ but not email", "team@channel", chats.GroupChatRef{}, "team@channel"},
		{"empty string", "", chats.GroupChatRef{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := GetChatRef(tt.input)

			switch tt.wantType.(type) {
			case chats.OneOnOneChatRef:
				v, ok := ref.(chats.OneOnOneChatRef)
				if !ok {
					t.Fatalf("expected OneOnOneChatRef, got %T", ref)
				}
				if v.Ref != tt.wantValue {
					t.Errorf("Ref = %q, want %q", v.Ref, tt.wantValue)
				}

			case chats.GroupChatRef:
				v, ok := ref.(chats.GroupChatRef)
				if !ok {
					t.Fatalf("expected GroupChatRef, got %T", ref)
				}
				if v.Ref != tt.wantValue {
					t.Errorf("Ref = %q, want %q", v.Ref, tt.wantValue)
				}
			}
		})
	}
}
