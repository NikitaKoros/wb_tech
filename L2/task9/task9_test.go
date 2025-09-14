package main

import "testing"

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"a4bc2d5e", "aaaabccddddde", false},
		{"abcd", "abcd", false},
		{"", "", false},

		{"45", "", true},
		{"\\", "", true},
		{"3abc", "", true},
		{"qwe\\", "", true},

		{"qwe\\4\\5", "qwe45", false},
		{"qwe\\45", "qwe44444", false},
		{"qwe\\415", "qwe444444444444444", false},

		{"a0bc", "bc", false},
	}

	for _, tt := range tests {
		got, err := Unpack(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("Unpack(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			continue
		}

		if got != tt.expected {
			t.Errorf("Unpack(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
