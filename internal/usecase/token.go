package usecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// generatePlaintextToken returns a 40-byte random hex string (80 chars).
func generatePlaintextToken() (string, error) {
	buf := make([]byte, 40)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// FormatToken returns "<id>|<plain>", the Sanctum wire format.
func FormatToken(id int64, plain string) string {
	return fmt.Sprintf("%d|%s", id, plain)
}
