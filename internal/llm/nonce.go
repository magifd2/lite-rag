package llm

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// maxNonceRetries caps the number of regeneration attempts when a collision
// is detected. A 16-byte (32 hex char) nonce appearing in real text is
// astronomically unlikely, but we cap retries to make the guarantee explicit.
const maxNonceRetries = 32

// NewNonce returns a 16-byte cryptographically random hex string.
func NewNonce() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failure is a fatal runtime condition on any supported OS.
		panic("llm: crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// NewNonceNotIn returns a nonce that does not appear as a substring in any of
// the provided texts. If the nonce were present in a passage or query, the
// nonce-tagged XML boundaries could be forged by crafted document content.
//
// Retries up to maxNonceRetries times on collision. If all attempts collide
// (should never happen in practice) it panics rather than silently using a
// colliding nonce.
func NewNonceNotIn(texts ...string) string {
	for range maxNonceRetries {
		n := NewNonce()
		if !nonceAppearsIn(n, texts) {
			return n
		}
	}
	panic("llm: could not generate a collision-free nonce after retries")
}

// nonceAppearsIn reports whether nonce is a substring of any text in the slice.
func nonceAppearsIn(nonce string, texts []string) bool {
	for _, t := range texts {
		if strings.Contains(t, nonce) {
			return true
		}
	}
	return false
}
