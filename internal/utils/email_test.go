package utils

import "testing"

func TestIsLikelyEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid simple", "test@example.com", true},
		{"valid with dots", "first.last@domain.co.uk", true},
		{"valid with plus", "user+tag@gmail.com", true},
		{"leading/trailing spaces", "  test@example.com  ", true},

		{"missing @", "testexample.com", false},
		{"missing domain", "test@", false},
		{"missing user", "@example.com", false},
		{"invalid chars", "test@exa mple.com", false},
		{"short tld", "test@domain.c", false},
		{"empty", "", false},
		{"subdomain", "a@b.coffee", true},
		{"numbers everywhere", "123.45+test@99-domain123.io", true},
		{"uppercase", "USER@EXAMPLE.COM", true},

		{"double dot local", "a..b@example.com", true},
		{"trailing dot domain", "a@b.com.", false},
		{"leading dot local", ".abc@test.com", true},
		{"hyphen domain start", "a@-test.com", true},

		{"unicode", "üser@test.com", false},
		{"quoted local", `"user"@test.com`, false},
		{"commented", "user(test)@example.com", false},
		{"newline injection", "test@example.com\n", true},
		{"tab injection", "test@example.com\t", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLikelyEmail(tt.input); got != tt.want {
				t.Errorf("IsLikelyEmail(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
