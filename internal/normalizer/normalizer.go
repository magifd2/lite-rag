// Package normalizer provides text normalization for mixed Japanese/English content.
package normalizer

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var (
	reControlChars    = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	reConsecutiveSpace = regexp.MustCompile(`[ \t]+`)
	reMarkdownImage   = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`)
	reMarkdownLink    = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)
	reHTMLTag         = regexp.MustCompile(`<[^>]+>`)
	reCodeFence       = regexp.MustCompile("(?m)^```[^\n]*\n")
	reInlineCode      = regexp.MustCompile("`([^`]+)`")
)

// Normalize applies NFKC normalization, whitespace cleanup, and control-character
// removal to the input text. It does NOT strip Markdown syntax — call
// StripMarkdown separately before generating embeddings.
func Normalize(text string) string {
	// Step 1: Unicode NFKC normalization
	// - Full-width ASCII → half-width (Ａ→A, １→1)
	// - Half-width Katakana → full-width (ｶﾅ→カナ)
	// - Compatibility characters decomposed (㍉→ミリ)
	text = norm.NFKC.String(text)

	// Step 2: Full-width space → ASCII space
	text = strings.ReplaceAll(text, "\u3000", " ")

	// Step 3: Normalize line endings to LF
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// Step 4: Remove C0/C1 control characters (keep LF=0x0A, TAB=0x09)
	text = reControlChars.ReplaceAllString(text, "")

	// Step 5: Collapse consecutive spaces/tabs on each line
	text = reConsecutiveSpace.ReplaceAllString(text, " ")

	return text
}

// StripMarkdown removes Markdown syntax artifacts from text before embedding.
// It should be called after Normalize, on the content of individual chunks.
func StripMarkdown(text string) string {
	// Images: ![alt](url) → alt
	text = reMarkdownImage.ReplaceAllString(text, "$1")
	// Links: [text](url) → text
	text = reMarkdownLink.ReplaceAllString(text, "$1")
	// HTML tags
	text = reHTMLTag.ReplaceAllString(text, "")
	// Fenced code block markers (``` lines) — keep the code content
	text = reCodeFence.ReplaceAllString(text, "")
	// Inline code: `code` → code
	text = reInlineCode.ReplaceAllString(text, "$1")

	return strings.TrimSpace(text)
}

// EstimateTokens returns a conservative token-count estimate for mixed
// Japanese/English text.
//
// Coefficients:
//   - CJK characters (Kanji, Hiragana, Katakana, etc.): 1 char ≈ 2 tokens
//   - ASCII/Latin words: 1 word ≈ 1.3 tokens
func EstimateTokens(text string) int {
	var cjkChars, asciiWords int
	inWord := false

	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		i += size

		if isCJK(r) {
			cjkChars++
			inWord = false
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if !inWord {
				asciiWords++
				inWord = true
			}
		} else {
			inWord = false
		}
	}

	return cjkChars*2 + int(float64(asciiWords)*1.3+0.5)
}

// isCJK reports whether r is a CJK unified ideograph, Hiragana, Katakana,
// or other East Asian script character.
func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		(r >= 0xFF01 && r <= 0xFF60) || // Fullwidth forms
		(r >= 0x3000 && r <= 0x303F) // CJK symbols and punctuation
}
