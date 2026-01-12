package utils

import (
	"testing"

	"github.com/pzsp-teams/lib/chats"
	"github.com/stretchr/testify/require"
)

func TestGetChatRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ref      string
		wantNil  bool
		wantType any
		wantRef  string
	}{
		{
			name:    "empty string -> nil",
			ref:     "",
			wantNil: true,
		},
		{
			name:    "whitespace only -> nil",
			ref:     "   \t\n",
			wantNil: true,
		},
		{
			name:     "email -> OneOnOneChatRef",
			ref:      "john.doe@example.com",
			wantType: chats.OneOnOneChatRef{},
			wantRef:  "john.doe@example.com",
		},
		{
			name:     "group chat id -> GroupChatRef",
			ref:      "19:abcdef1234567890@thread.v2",
			wantType: chats.GroupChatRef{},
			wantRef:  "19:abcdef1234567890@thread.v2",
		},
		{
			name:     "one-on-one chat id -> OneOnOneChatRef",
			ref:      "someone@example.com",
			wantType: chats.OneOnOneChatRef{},
			wantRef:  "someone@example.com",
		},
		{
			name:     "random non-empty string -> GroupChatRef",
			ref:      "not-an-email",
			wantType: chats.GroupChatRef{},
			wantRef:  "not-an-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GetChatRef(tt.ref)

			if tt.wantNil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)

			switch tt.wantType.(type) {
			case chats.OneOnOneChatRef:
				r, ok := got.(chats.OneOnOneChatRef)
				require.True(t, ok, "expected OneOnOneChatRef, got %T", got)
				require.Equal(t, tt.wantRef, r.Ref)

			case chats.GroupChatRef:
				r, ok := got.(chats.GroupChatRef)
				require.True(t, ok, "expected GroupChatRef, got %T", got)
				require.Equal(t, tt.wantRef, r.Ref)

			default:
				t.Fatalf("unsupported wantType: %T", tt.wantType)
			}
		})
	}
}
