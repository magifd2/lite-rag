package normalizer_test

import (
	"testing"

	"lite-rag/internal/normalizer"
)

func TestNormalize_NFKC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "full-width ASCII to half-width",
			input: "Ａ１Ｂ２",
			want:  "A1B2",
		},
		{
			name:  "half-width Katakana to full-width",
			input: "ｶﾅ",
			want:  "カナ",
		},
		{
			name:  "compatibility character",
			input: "㍉",
			want:  "ミリ",
		},
		{
			name:  "full-width space to ASCII space",
			input: "hello\u3000world",
			want:  "hello world",
		},
		{
			name:  "CRLF to LF",
			input: "line1\r\nline2",
			want:  "line1\nline2",
		},
		{
			name:  "consecutive spaces collapsed",
			input: "a   b\t\tc",
			want:  "a b c",
		},
		{
			name:  "control characters removed",
			input: "hello\x01\x07world",
			want:  "helloworld",
		},
		{
			name:  "LF preserved; inline TAB collapsed to space",
			input: "line1\nline2\t\ttabbed",
			want:  "line1\nline2 tabbed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizer.Normalize(tt.input)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "image syntax",
			input: "![alt text](https://example.com/image.png)",
			want:  "alt text",
		},
		{
			name:  "link syntax",
			input: "[click here](https://example.com)",
			want:  "click here",
		},
		{
			name:  "HTML tag",
			input: "<strong>bold</strong>",
			want:  "bold",
		},
		{
			name:  "fenced code block marker",
			input: "```go\nfmt.Println()\n```",
			want:  "fmt.Println()\n```",
		},
		{
			name:  "inline code",
			input: "use `fmt.Println` to print",
			want:  "use fmt.Println to print",
		},
		{
			name:  "mixed Markdown",
			input: "See [docs](https://example.com) and ![logo](logo.png) for details.",
			want:  "See docs and logo for details.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizer.StripMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("StripMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantMin int
		wantMax int
	}{
		{
			name:    "empty string",
			input:   "",
			wantMin: 0,
			wantMax: 0,
		},
		{
			name:    "ASCII only",
			input:   "hello world foo bar", // 4 words × 1.3 ≈ 5
			wantMin: 4,
			wantMax: 8,
		},
		{
			name:    "CJK only",
			input:   "日本語テスト", // 6 chars × 2 = 12
			wantMin: 10,
			wantMax: 14,
		},
		{
			name:    "mixed JP/EN",
			input:   "GoでRAGを実装する", // ~3 words + 6 CJK chars
			wantMin: 10,
			wantMax: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizer.EstimateTokens(tt.input)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("EstimateTokens(%q) = %d, want [%d, %d]",
					tt.input, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}
