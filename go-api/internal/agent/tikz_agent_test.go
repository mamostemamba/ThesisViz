package agent

import "testing"

func TestExtractFixedCode_WithMarker(t *testing.T) {
	input := "Some explanation text\n=== FIXED CODE ===\n\\begin{tikzpicture}\n\\end{tikzpicture}"
	got := extractFixedCode(input)
	want := "\n\\begin{tikzpicture}\n\\end{tikzpicture}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractFixedCode_WithoutMarker(t *testing.T) {
	input := "\\begin{tikzpicture}\n\\end{tikzpicture}"
	got := extractFixedCode(input)
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestExtractFixedCode_MarkerAtEnd(t *testing.T) {
	input := "Some text\n=== FIXED CODE ==="
	got := extractFixedCode(input)
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}
