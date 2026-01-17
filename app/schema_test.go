package app

import "testing"

func TestValidateFlags(t *testing.T) {
	cases := []struct {
		name    string
		defs    []FlagDef
		flags   map[string]any
		wantErr bool
	}{
		{
			name: "invalid choice",
			defs: []FlagDef{{
				Name:    "format",
				Type:    InputChoice,
				Options: []string{"json", "yaml"},
			}},
			flags:   map[string]any{"format": "xml"},
			wantErr: true,
		},
		{
			name: "missing required flag",
			defs: []FlagDef{{
				Name:          "template",
				RequiresFlags: []string{"data"},
			}},
			flags:   map[string]any{"template": "msg.txt"},
			wantErr: true,
		},
		{
			name: "conflicting flags",
			defs: []FlagDef{{
				Name:          "file",
				ConflictsWith: []string{"message"},
			}},
			flags: map[string]any{
				"file":    "out.txt",
				"message": "hello",
			},
			wantErr: true,
		},
		{
			name: "valid flags",
			defs: []FlagDef{
				{
					Name:          "template",
					RequiresFlags: []string{"data"},
				},
				{
					Name:    "format",
					Type:    InputChoice,
					Options: []string{"json", "yaml"},
				},
			},
			flags: map[string]any{
				"template": "msg.txt",
				"data":     "data.yaml",
				"format":   "json",
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFlags(tc.flags, tc.defs)
			if (err != nil) != tc.wantErr {
				t.Fatalf("expected error=%v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestIsSet(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "string empty", value: "", want: false},
		{name: "string value", value: "hi", want: true},
		{name: "bool false", value: false, want: false},
		{name: "bool true", value: true, want: true},
		{name: "int zero", value: 0, want: true},
		{name: "nil value", value: nil, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			flags := map[string]any{"value": tc.value}
			if got := isSet(flags, "value"); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
