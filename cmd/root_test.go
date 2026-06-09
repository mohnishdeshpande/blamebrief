package cmd

import "testing"

func TestParseLines(t *testing.T) {
	tests := []struct {
		input     string
		wantStart int
		wantEnd   int
		wantErr   bool
	}{
		{"10-45", 10, 45, false},
		{"10:45", 10, 45, false},
		{"25", 25, 25, false},
		{"  30 - 40  ", 30, 40, false},
		{"-10", 0, 0, true},
		{"10-5", 0, 0, true},
		{"abc", 0, 0, true},
		{"10-abc", 0, 0, true},
		{"", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			start, end, err := parseLines(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLines(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if start != tt.wantStart || end != tt.wantEnd {
					t.Errorf("parseLines(%q) got = (%d, %d), want = (%d, %d)", tt.input, start, end, tt.wantStart, tt.wantEnd)
				}
			}
		})
	}
}
