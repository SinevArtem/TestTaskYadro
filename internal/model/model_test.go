package model

import (
	"testing"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{SUCCESS, "SUCCESS"},
		{FAIL, "FAIL"},
		{DISQUAL, "DISQUAL"},
		{Status(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		result := tt.status.String()
		if result != tt.expected {
			t.Errorf("Status(%d).String() = %s, want %s", tt.status, result, tt.expected)
		}
	}
}
