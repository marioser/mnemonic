package sync

import "testing"

func TestStripHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<b>Bold</b> and <i>italic</i>", "Bold and italic"},
		{"No HTML here", "No HTML here"},
		{"", ""},
		{"<div><p>Nested</p></div>", "Nested"},
		{"Multiple   spaces", "Multiple spaces"},
	}

	for _, tt := range tests {
		got := StripHTML(tt.input)
		if got != tt.expected {
			t.Errorf("StripHTML(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestStripHTMLTruncation(t *testing.T) {
	long := "<p>" + string(make([]byte, 600)) + "</p>"
	got := StripHTML(long)
	if len(got) > 500 {
		t.Errorf("expected max 500 chars, got %d", len(got))
	}
}
