package utils

import "testing"

func TestIsLikelyGroupChatID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"whitespace only", "   \t\n", false},
		{"valid thread v2", "19:abcdef@thread.v2", true},
		{"valid thread msg", "19:something@thread.skype", true},
		{"trimmed valid", "  19:abc@thread.v2  ", true},
		{"missing 19 prefix", "18:abc@thread.v2", false},
		{"missing @thread dot", "19:abc@thread", false},
		{"wrong marker @unq", "19:abc@unq.v2", false},
		{"contains thread but no prefix", "xx19:abc@thread.v2", false},
		{"prefix ok but thread missing", "19:abc@unq.v2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsLikelyGroupChatID(tt.in); got != tt.want {
				t.Fatalf("IsLikelyGroupChatID(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsLikelyOneOnOneChatID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"whitespace only", "   \t\n", false},
		{"valid unq", "19:abcdef@unq.v2", true},
		{"valid unq other", "19:xyz@unq.gbl.spaces", true},
		{"trimmed valid", "  19:abc@unq.v2  ", true},
		{"missing 19 prefix", "18:abc@unq.v2", false},
		{"missing @unq dot", "19:abc@unq", false},
		{"wrong marker @thread", "19:abc@thread.v2", false},
		{"contains unq but no prefix", "xx19:abc@unq.v2", false},
		{"prefix ok but unq missing", "19:abc@thread.v2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsLikelyOneOnOneChatID(tt.in); got != tt.want {
				t.Fatalf("IsLikelyOneOnOneChatID(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsLikelyEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"whitespace only", "   \t\n", false},
		{"simple valid", "john.doe@example.com", true},
		{"trimmed valid", "  john.doe@example.com  ", true},
		{"plus and dash", "a.b-c+tag_1@sub.example-domain.com", true},
		{"missing at", "john.doe.example.com", false},
		{"missing domain", "john@", false},
		{"missing local part", "@example.com", false},
		{"missing dot tld", "john@example", false},
		{"tld too short", "john@example.c", false},
		{"double at", "a@@example.com", false},
		{"spaces inside", "john doe@example.com", false},
		{"unicode local part", "żółw@example.com", false},
		{"underscore in domain not allowed by regex", "john@exa_mple.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsLikelyEmail(tt.in); got != tt.want {
				t.Fatalf("IsLikelyEmail(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
